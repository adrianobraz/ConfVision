package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"confservice/db"
	"confservice/internal/auth"
)

// EventoEnfileirar — chamado pelo Xano/ConfMonit quando chega alarm_events de cliente parceiro.
func (h *Handler) EventoEnfileirar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	var in struct {
		IDFranqueado  string         `json:"idFranqueado"`
		IDCliente     string         `json:"idCliente"`
		AlarmEventsID *int64         `json:"alarmEventsId"`
		Payload       map[string]any `json:"payload"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.IDFranqueado = strings.TrimSpace(in.IDFranqueado)
	in.IDCliente = strings.TrimSpace(in.IDCliente)
	if in.IDFranqueado == "" || in.IDCliente == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado e idCliente obrigatorios")
		return
	}
	if in.Payload == nil {
		in.Payload = map[string]any{}
	}

	var idVinculo, idParceiro string
	err := db.Conn.QueryRow(`
		SELECT id, id_parceiro FROM cs_cliente_vinculo
		WHERE id_franqueado = ? AND id_cliente = ? AND ativo = 1 LIMIT 1`,
		in.IDFranqueado, in.IDCliente,
	).Scan(&idVinculo, &idParceiro)
	if err == sql.ErrNoRows {
		writeErr(w, http.StatusNotFound, "cliente sem parceiro ativo")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}

	raw, _ := json.Marshal(in.Payload)
	res, err := db.Conn.Exec(`
		INSERT INTO cs_webhook_fila
		(id_vinculo, id_parceiro, id_franqueado, id_cliente, alarm_events_id, payload_json, status, tentativas, proximo_em)
		VALUES (?, ?, ?, ?, ?, ?, 'PENDENTE', 0, NOW(3))`,
		idVinculo, idParceiro, in.IDFranqueado, in.IDCliente, in.AlarmEventsID, string(raw),
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "falha ao enfileirar")
		return
	}
	filaID, _ := res.LastInsertId()

	// Disparo best-effort imediato
	go h.processarFilaItem(filaID)

	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "filaId": filaID})
}

func (h *Handler) WebhookProcessar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	n := h.ProcessarPendentes(20)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "processados": n})
}

// ProcessarPendentes exposto para o worker em background.
func (h *Handler) ProcessarPendentes(limite int) int {
	return h.processarPendentes(limite)
}

func (h *Handler) processarPendentes(limite int) int {
	rows, err := db.Conn.Query(`
		SELECT id FROM cs_webhook_fila
		WHERE status IN ('PENDENTE','ERRO') AND proximo_em <= NOW(3) AND tentativas < 8
		ORDER BY id ASC LIMIT ?`, limite)
	if err != nil {
		return 0
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	for _, id := range ids {
		h.processarFilaItem(id)
	}
	return len(ids)
}

func (h *Handler) processarFilaItem(id int64) {
	var idParceiro string
	var payload []byte
	var tentativas int
	err := db.Conn.QueryRow(`
		SELECT id_parceiro, payload_json, tentativas FROM cs_webhook_fila WHERE id = ?`, id).
		Scan(&idParceiro, &payload, &tentativas)
	if err != nil {
		return
	}

	var webhookURL, webhookToken string
	err = db.Conn.QueryRow(`SELECT webhook_url, webhook_token FROM cs_parceiro WHERE id = ? AND ativo = 1`, idParceiro).
		Scan(&webhookURL, &webhookToken)
	if err != nil || strings.TrimSpace(webhookURL) == "" {
		_, _ = db.Conn.Exec(`
			UPDATE cs_webhook_fila SET status='ERRO', tentativas=tentativas+1, ultimo_erro=?, proximo_em=DATE_ADD(NOW(3), INTERVAL 15 MINUTE)
			WHERE id=?`, "parceiro sem webhook_url", id)
		return
	}

	_, _ = db.Conn.Exec(`UPDATE cs_webhook_fila SET status='ENVIANDO', updated_at=NOW(3) WHERE id=?`, id)

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		h.marcarErro(id, tentativas, 0, err.Error(), "")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ConfService/1.0")
	if webhookToken != "" {
		req.Header.Set("Authorization", "Bearer "+webhookToken)
		req.Header.Set("X-Webhook-Token", webhookToken)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.marcarErro(id, tentativas, 0, err.Error(), "")
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_, _ = db.Conn.Exec(`
			UPDATE cs_webhook_fila SET status='OK', http_status=?, response_body=?, ultimo_erro=NULL, updated_at=NOW(3)
			WHERE id=?`, resp.StatusCode, string(body), id)
		return
	}
	h.marcarErro(id, tentativas, resp.StatusCode, "http "+resp.Status, string(body))
}

func (h *Handler) marcarErro(id int64, tentativas, httpStatus int, msg, body string) {
	mins := 1 << min(tentativas, 6) // 1,2,4,8,16,32,64
	_, _ = db.Conn.Exec(`
		UPDATE cs_webhook_fila SET
			status='ERRO', tentativas=tentativas+1, http_status=?, ultimo_erro=?, response_body=?,
			proximo_em=DATE_ADD(NOW(3), INTERVAL ? MINUTE), updated_at=NOW(3)
		WHERE id=?`, httpStatus, truncate(msg, 500), truncate(body, 8000), mins, id)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
