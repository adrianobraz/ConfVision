package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"confservice/db"
	"confservice/internal/auth"
)

type eventoItem struct {
	ID            int64          `json:"id"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
	IDParceiro    string         `json:"idParceiro"`
	IDFranqueado  string         `json:"idFranqueado"`
	IDCliente     string         `json:"idCliente"`
	NomeCliente   string         `json:"nomeCliente"`
	AlarmEventsID *int64         `json:"alarmEventsId,omitempty"`
	Status        string         `json:"status"`
	Tentativas    int            `json:"tentativas"`
	HTTPStatus    *int           `json:"httpStatus,omitempty"`
	UltimoErro    string         `json:"ultimoErro,omitempty"`
	Grupo         string         `json:"grupo,omitempty"`
	Codigo        string         `json:"codigo,omitempty"`
	Descricao     string         `json:"descricao,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
}

type eventoResumo struct {
	Total    int `json:"total"`
	OK       int `json:"ok"`
	Erro     int `json:"erro"`
	Pendente int `json:"pendente"`
	Enviando int `json:"enviando"`
}

// EventosParceiroMe — relatorio de eventos/webhooks do parceiro (prova de envio).
func (h *Handler) EventosParceiroMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	claims, ok := h.requireParceiro(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	list, resumo, err := listarEventos(claims.ParceiroID, q.Get("de"), q.Get("ate"), q.Get("status"), q.Get("idCliente"), queryInt(q.Get("limite"), 100))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "resumo": resumo, "dados": list})
}

// EventosInternal — admin (X-Api-Key): filtro opcional por idParceiro.
func (h *Handler) EventosInternal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	q := r.URL.Query()
	list, resumo, err := listarEventos(strings.TrimSpace(q.Get("idParceiro")), q.Get("de"), q.Get("ate"), q.Get("status"), q.Get("idCliente"), queryInt(q.Get("limite"), 200))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "resumo": resumo, "dados": list})
}

func listarEventos(idParceiro, de, ate, status, idCliente string, limite int) ([]eventoItem, eventoResumo, error) {
	var res eventoResumo
	if limite < 1 {
		limite = 50
	}
	if limite > 500 {
		limite = 500
	}

	where := []string{"1=1"}
	args := []any{}
	if idParceiro != "" {
		where = append(where, "f.id_parceiro = ?")
		args = append(args, idParceiro)
	}
	if status != "" {
		where = append(where, "f.status = ?")
		args = append(args, strings.ToUpper(strings.TrimSpace(status)))
	}
	if idCliente != "" {
		where = append(where, "f.id_cliente = ?")
		args = append(args, strings.TrimSpace(idCliente))
	}
	if de != "" {
		if _, err := time.Parse("2006-01-02", de); err == nil {
			where = append(where, "f.created_at >= ?")
			args = append(args, de+" 00:00:00")
		}
	}
	if ate != "" {
		if _, err := time.Parse("2006-01-02", ate); err == nil {
			where = append(where, "f.created_at <= ?")
			args = append(args, ate+" 23:59:59.999")
		}
	}
	wSQL := strings.Join(where, " AND ")

	// resumo
	sumRows, err := db.Conn.Query(`
		SELECT f.status, COUNT(*) FROM cs_webhook_fila f
		WHERE `+wSQL+`
		GROUP BY f.status`, args...)
	if err != nil {
		return nil, res, err
	}
	defer sumRows.Close()
	for sumRows.Next() {
		var st string
		var n int
		if err := sumRows.Scan(&st, &n); err != nil {
			continue
		}
		res.Total += n
		switch st {
		case "OK":
			res.OK = n
		case "ERRO":
			res.Erro = n
		case "PENDENTE":
			res.Pendente = n
		case "ENVIANDO":
			res.Enviando = n
		}
	}

	qArgs := append([]any{}, args...)
	qArgs = append(qArgs, limite)
	rows, err := db.Conn.Query(`
		SELECT f.id, f.created_at, f.updated_at, f.id_parceiro, f.id_franqueado, f.id_cliente,
		       COALESCE(v.nome_cliente, ''), f.alarm_events_id, f.status, f.tentativas,
		       f.http_status, COALESCE(f.ultimo_erro, ''), f.payload_json
		FROM cs_webhook_fila f
		LEFT JOIN cs_cliente_vinculo v ON v.id = f.id_vinculo
		WHERE `+wSQL+`
		ORDER BY f.id DESC
		LIMIT ?`, qArgs...)
	if err != nil {
		return nil, res, err
	}
	defer rows.Close()

	list := []eventoItem{}
	for rows.Next() {
		var it eventoItem
		var created, updated time.Time
		var alarm sql.NullInt64
		var httpSt sql.NullInt64
		var payloadRaw []byte
		if err := rows.Scan(
			&it.ID, &created, &updated, &it.IDParceiro, &it.IDFranqueado, &it.IDCliente,
			&it.NomeCliente, &alarm, &it.Status, &it.Tentativas,
			&httpSt, &it.UltimoErro, &payloadRaw,
		); err != nil {
			continue
		}
		it.CreatedAt = created.Format(time.RFC3339)
		it.UpdatedAt = updated.Format(time.RFC3339)
		if alarm.Valid {
			v := alarm.Int64
			it.AlarmEventsID = &v
		}
		if httpSt.Valid {
			v := int(httpSt.Int64)
			it.HTTPStatus = &v
		}
		var payload map[string]any
		if len(payloadRaw) > 0 {
			_ = json.Unmarshal(payloadRaw, &payload)
		}
		if payload != nil {
			it.Payload = payload
			it.Grupo = strAny(payload["ctiGrupo"], payload["grupo"])
			it.Codigo = strAny(payload["codigo"])
			it.Descricao = strAny(payload["ctiDescricao"], payload["descricao"])
		}
		list = append(list, it)
	}
	return list, res, nil
}

func queryInt(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func strAny(vals ...any) string {
	for _, v := range vals {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				return t
			}
		case float64:
			return strconv.FormatFloat(t, 'f', -1, 64)
		case json.Number:
			return t.String()
		}
	}
	return ""
}
