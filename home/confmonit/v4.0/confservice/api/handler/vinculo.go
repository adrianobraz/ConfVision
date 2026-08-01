package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"confservice/db"
	"confservice/internal/auth"

	"github.com/google/uuid"
)

func (h *Handler) VinculoCriar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	// Franqueado Pro / Xano: autenticado por API key interna
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	var in struct {
		IDFranqueado string `json:"idFranqueado"`
		IDCliente    string `json:"idCliente"`
		NomeCliente  string `json:"nomeCliente"`
		IDParceiro   string `json:"idParceiro"`
		ContaExterna string `json:"contaExterna"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.IDFranqueado = strings.TrimSpace(in.IDFranqueado)
	in.IDCliente = strings.TrimSpace(in.IDCliente)
	in.IDParceiro = strings.TrimSpace(in.IDParceiro)
	if in.IDFranqueado == "" || in.IDCliente == "" || in.IDParceiro == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado, idCliente e idParceiro obrigatorios")
		return
	}

	var preco float64
	var ativo bool
	err := db.Conn.QueryRow(`SELECT preco_cliente_quinzena, ativo FROM cs_parceiro WHERE id = ?`, in.IDParceiro).
		Scan(&preco, &ativo)
	if err == sql.ErrNoRows || !ativo {
		writeErr(w, http.StatusBadRequest, "parceiro invalido ou inativo")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}

	permitido, err := parceiroPermitidoFranqueado(in.IDFranqueado, in.IDParceiro)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	if !permitido {
		writeErr(w, http.StatusForbidden, "parceiro nao liberado para este franqueado")
		return
	}

	// Desativa vinculo anterior ativo
	_, _ = db.Conn.Exec(`
		UPDATE cs_cliente_vinculo SET ativo = 0, fim_em = CURDATE()
		WHERE id_franqueado = ? AND id_cliente = ? AND ativo = 1`,
		in.IDFranqueado, in.IDCliente,
	)

	id := uuid.NewString()
	_, err = db.Conn.Exec(`
		INSERT INTO cs_cliente_vinculo
		(id, id_franqueado, id_cliente, nome_cliente, id_parceiro, conta_externa, preco_congelado, ativo, inicio_em)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, CURDATE())`,
		id, in.IDFranqueado, in.IDCliente, nullStr(in.NomeCliente), in.IDParceiro, nullStr(in.ContaExterna), preco,
	)
	if err != nil {
		writeErr(w, http.StatusConflict, "falha ao vincular")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true,
		"vinculo": map[string]any{
			"id":             id,
			"idFranqueado":   in.IDFranqueado,
			"idCliente":      in.IDCliente,
			"idParceiro":     in.IDParceiro,
			"precoCongelado": preco,
			"ativo":          true,
		},
	})
}

func (h *Handler) VinculoDesativar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	var in struct {
		IDFranqueado string `json:"idFranqueado"`
		IDCliente    string `json:"idCliente"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	res, err := db.Conn.Exec(`
		UPDATE cs_cliente_vinculo SET ativo = 0, fim_em = CURDATE()
		WHERE id_franqueado = ? AND id_cliente = ? AND ativo = 1`,
		strings.TrimSpace(in.IDFranqueado), strings.TrimSpace(in.IDCliente),
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "afetados": n})
}

func (h *Handler) VinculoPorCliente(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	idFranq := strings.TrimSpace(r.URL.Query().Get("idFranqueado"))
	idCli := strings.TrimSpace(r.URL.Query().Get("idCliente"))
	if idFranq == "" || idCli == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado e idCliente obrigatorios")
		return
	}
	var id, idParceiro string
	var nome, conta sql.NullString
	var preco float64
	err := db.Conn.QueryRow(`
		SELECT id, id_parceiro, nome_cliente, conta_externa, preco_congelado
		FROM cs_cliente_vinculo
		WHERE id_franqueado = ? AND id_cliente = ? AND ativo = 1
		LIMIT 1`, idFranq, idCli).
		Scan(&id, &idParceiro, &nome, &conta, &preco)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "vinculo": nil})
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"vinculo": map[string]any{
			"id":             id,
			"idFranqueado":   idFranq,
			"idCliente":      idCli,
			"idParceiro":     idParceiro,
			"nomeCliente":    nome.String,
			"contaExterna":   conta.String,
			"precoCongelado": preco,
			"ativo":          true,
		},
	})
}

func (h *Handler) MeusClientes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	claims, ok := h.requireParceiro(w, r)
	if !ok {
		return
	}
	rows, err := db.Conn.Query(`
		SELECT id, id_franqueado, id_cliente, nome_cliente, conta_externa, preco_congelado, inicio_em
		FROM cs_cliente_vinculo
		WHERE id_parceiro = ? AND ativo = 1
		ORDER BY nome_cliente, id_cliente`, claims.ParceiroID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, idFranq, idCli string
		var nome, conta sql.NullString
		var preco float64
		var inicio time.Time
		if err := rows.Scan(&id, &idFranq, &idCli, &nome, &conta, &preco, &inicio); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id":             id,
			"idFranqueado":   idFranq,
			"idCliente":      idCli,
			"nomeCliente":    nome.String,
			"contaExterna":   conta.String,
			"precoCongelado": preco,
			"inicioEm":       inicio.Format("2006-01-02"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}
