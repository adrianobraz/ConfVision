package handler

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"confservice/config"
	"confservice/db"
	"confservice/internal/auth"
)

func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		Usuario string `json:"usuario"`
		Senha   string `json:"senha"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.Usuario = strings.TrimSpace(in.Usuario)
	if in.Usuario != config.AdminUser || in.Senha != config.AdminPass {
		writeErr(w, http.StatusUnauthorized, "credenciais invalidas")
		return
	}
	token, err := auth.EmitirToken("admin", in.Usuario, "admin", 8*time.Hour)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "token": token})
}

func (h *Handler) AdminParceiros(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !h.requireAdmin(w, r) {
		return
	}
	rows, err := db.Conn.Query(`
		SELECT id, razao_social, nome_fantasia, email, telefone, software,
		       preco_cliente_quinzena, ativo, created_at
		FROM cs_parceiro ORDER BY razao_social`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list := []map[string]any{}
	for rows.Next() {
		var id, razao, email, software string
		var fantasia, tel sql.NullString
		var preco float64
		var ativo bool
		var created time.Time
		if err := rows.Scan(&id, &razao, &fantasia, &email, &tel, &software, &preco, &ativo, &created); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id": id, "razaoSocial": razao, "nomeFantasia": fantasia.String,
			"email": email, "telefone": tel.String, "software": software,
			"precoClienteQuinzena": preco, "ativo": ativo,
			"createdAt": created.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) AdminResetSenha(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !h.requireAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, "id obrigatorio")
		return
	}
	var in struct {
		Senha string `json:"senha"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	if len(in.Senha) < 6 {
		writeErr(w, http.StatusBadRequest, "senha deve ter no minimo 6 caracteres")
		return
	}
	hash, err := auth.HashSenha(in.Senha)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash")
		return
	}
	res, err := db.Conn.Exec(`UPDATE cs_parceiro SET senha_hash = ? WHERE id = ?`, hash, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeErr(w, http.StatusNotFound, "parceiro nao encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	raw, err := auth.Bearer(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "nao autenticado")
		return false
	}
	claims, err := auth.ParseToken(raw)
	if err != nil || claims.Role != "admin" {
		writeErr(w, http.StatusUnauthorized, "token invalido")
		return false
	}
	return true
}
