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

type ParceiroPublico struct {
	ID                   string   `json:"id"`
	RazaoSocial          string   `json:"razaoSocial"`
	NomeFantasia         string   `json:"nomeFantasia"`
	Email                string   `json:"email"`
	Telefone             string   `json:"telefone"`
	Software             string   `json:"software"`
	WebhookURL           string   `json:"webhookUrl"`
	PrecoClienteQuinzena float64  `json:"precoClienteQuinzena"`
	ComissaoPct          *float64 `json:"comissaoPct,omitempty"`
	Ativo                bool     `json:"ativo"`
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "service": "confservice"})
}

func (h *Handler) ParceiroRegistrar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		RazaoSocial          string  `json:"razaoSocial"`
		NomeFantasia         string  `json:"nomeFantasia"`
		CNPJ                 string  `json:"cnpj"`
		Email                string  `json:"email"`
		Telefone             string  `json:"telefone"`
		Software             string  `json:"software"`
		WebhookURL           string  `json:"webhookUrl"`
		WebhookToken         string  `json:"webhookToken"`
		PrecoClienteQuinzena float64 `json:"precoClienteQuinzena"`
		Senha                string  `json:"senha"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	in.RazaoSocial = strings.TrimSpace(in.RazaoSocial)
	if in.Email == "" || in.RazaoSocial == "" || len(in.Senha) < 6 {
		writeErr(w, http.StatusBadRequest, "razaoSocial, email e senha(>=6) obrigatorios")
		return
	}
	if in.Software == "" {
		in.Software = "GENERICO"
	}
	hash, err := auth.HashSenha(in.Senha)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash")
		return
	}
	id := uuid.NewString()
	_, err = db.Conn.Exec(`
		INSERT INTO cs_parceiro
		(id, razao_social, nome_fantasia, cnpj, email, telefone, software, webhook_url, webhook_token, preco_cliente_quinzena, senha_hash, ativo)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		id, in.RazaoSocial, nullStr(in.NomeFantasia), nullStr(in.CNPJ), in.Email, nullStr(in.Telefone),
		in.Software, in.WebhookURL, in.WebhookToken, in.PrecoClienteQuinzena, hash,
	)
	if err != nil {
		writeErr(w, http.StatusConflict, "nao foi possivel cadastrar (email ja existe?)")
		return
	}
	token, err := auth.EmitirToken(id, in.Email, "parceiro", 24*time.Hour)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":    true,
		"token": token,
		"parceiro": ParceiroPublico{
			ID: id, RazaoSocial: in.RazaoSocial, NomeFantasia: in.NomeFantasia,
			Email: in.Email, Telefone: in.Telefone, Software: in.Software,
			WebhookURL: in.WebhookURL, PrecoClienteQuinzena: in.PrecoClienteQuinzena, Ativo: true,
		},
	})
}

func (h *Handler) ParceiroLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	var in struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	var id, hash string
	var ativo bool
	err := db.Conn.QueryRow(`SELECT id, senha_hash, ativo FROM cs_parceiro WHERE email = ?`, in.Email).
		Scan(&id, &hash, &ativo)
	if err == sql.ErrNoRows || !auth.ChecarSenha(hash, in.Senha) {
		writeErr(w, http.StatusUnauthorized, "credenciais invalidas")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	if !ativo {
		writeErr(w, http.StatusForbidden, "parceiro inativo")
		return
	}
	token, err := auth.EmitirToken(id, in.Email, "parceiro", 24*time.Hour)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "token": token})
}

func (h *Handler) ParceiroMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := h.requireParceiro(w, r)
	if !ok {
		return
	}
	p, err := carregarParceiro(claims.ParceiroID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "parceiro nao encontrado")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "parceiro": p})
}

func (h *Handler) ParceiroAtualizar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	claims, ok := h.requireParceiro(w, r)
	if !ok {
		return
	}
	var in struct {
		NomeFantasia         string  `json:"nomeFantasia"`
		Telefone             string  `json:"telefone"`
		Software             string  `json:"software"`
		WebhookURL           string  `json:"webhookUrl"`
		WebhookToken         string  `json:"webhookToken"`
		PrecoClienteQuinzena float64 `json:"precoClienteQuinzena"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	if in.Software == "" {
		in.Software = "GENERICO"
	}
	_, err := db.Conn.Exec(`
		UPDATE cs_parceiro SET
			nome_fantasia = ?, telefone = ?, software = ?,
			webhook_url = ?, webhook_token = ?, preco_cliente_quinzena = ?
		WHERE id = ?`,
		nullStr(in.NomeFantasia), nullStr(in.Telefone), in.Software,
		in.WebhookURL, in.WebhookToken, in.PrecoClienteQuinzena, claims.ParceiroID,
	)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	p, _ := carregarParceiro(claims.ParceiroID)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "parceiro": p})
}

func (h *Handler) ParceirosCatalogo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	// Catálogo público descontinuado: parceiro só aparece após liberação Central + REP.
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": []map[string]any{}})
}

func carregarParceiro(id string) (*ParceiroPublico, error) {
	var p ParceiroPublico
	var fantasia, tel sql.NullString
	var comissao sql.NullFloat64
	err := db.Conn.QueryRow(`
		SELECT id, razao_social, nome_fantasia, email, telefone, software, webhook_url,
		       preco_cliente_quinzena, comissao_pct, ativo
		FROM cs_parceiro WHERE id = ?`, id).
		Scan(&p.ID, &p.RazaoSocial, &fantasia, &p.Email, &tel, &p.Software, &p.WebhookURL,
			&p.PrecoClienteQuinzena, &comissao, &p.Ativo)
	if err != nil {
		return nil, err
	}
	p.NomeFantasia = fantasia.String
	p.Telefone = tel.String
	if comissao.Valid {
		v := comissao.Float64
		p.ComissaoPct = &v
	}
	return &p, nil
}

func (h *Handler) requireParceiro(w http.ResponseWriter, r *http.Request) (*auth.Claims, bool) {
	raw, err := auth.Bearer(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "nao autenticado")
		return nil, false
	}
	claims, err := auth.ParseToken(raw)
	if err != nil || claims.Role != "parceiro" {
		writeErr(w, http.StatusUnauthorized, "token invalido")
		return nil, false
	}
	return claims, true
}

func nullStr(s string) any {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}
