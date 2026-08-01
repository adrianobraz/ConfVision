package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"confservice/db"
	"confservice/internal/auth"
)

func (h *Handler) ParceirosGlobal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	list, err := listarParceirosAtivos()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) ParceirosRepListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	idCentral := strings.TrimSpace(r.URL.Query().Get("idCentral"))
	idRep := strings.TrimSpace(r.URL.Query().Get("idRepresentante"))
	if idCentral == "" || idRep == "" {
		writeErr(w, http.StatusBadRequest, "idCentral e idRepresentante obrigatorios")
		return
	}
	rows, err := db.Conn.Query(`
		SELECT p.id, p.razao_social, p.nome_fantasia, p.software, p.preco_cliente_quinzena, p.ativo
		FROM cs_parceiro_rep pr
		JOIN cs_parceiro p ON p.id = pr.id_parceiro
		WHERE pr.id_central = ? AND pr.id_representante = ? AND pr.ativo = 1 AND p.ativo = 1
		ORDER BY p.razao_social`, idCentral, idRep)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list, err := scanParceirosCatalogo(rows)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) ParceirosRepSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	var in struct {
		IDCentral       string   `json:"idCentral"`
		IDRepresentante string   `json:"idRepresentante"`
		IDsParceiro     []string `json:"idsParceiro"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.IDCentral = strings.TrimSpace(in.IDCentral)
	in.IDRepresentante = strings.TrimSpace(in.IDRepresentante)
	if in.IDCentral == "" || in.IDRepresentante == "" {
		writeErr(w, http.StatusBadRequest, "idCentral e idRepresentante obrigatorios")
		return
	}
	ids := normalizeIDs(in.IDsParceiro)
	if err := validarParceirosExistem(ids); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM cs_parceiro_rep WHERE id_central = ? AND id_representante = ?`,
		in.IDCentral, in.IDRepresentante); err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	for _, idParceiro := range ids {
		if _, err := tx.Exec(`
			INSERT INTO cs_parceiro_rep (id_central, id_representante, id_parceiro, ativo)
			VALUES (?, ?, ?, 1)`, in.IDCentral, in.IDRepresentante, idParceiro); err != nil {
			writeErr(w, http.StatusInternalServerError, "db")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "total": len(ids)})
}

func (h *Handler) ParceirosFranqueadoListar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	idFranq := strings.TrimSpace(r.URL.Query().Get("idFranqueado"))
	idRep := strings.TrimSpace(r.URL.Query().Get("idRepresentante"))
	if idFranq == "" || idRep == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado e idRepresentante obrigatorios")
		return
	}
	rows, err := db.Conn.Query(`
		SELECT p.id, p.razao_social, p.nome_fantasia, p.software, p.preco_cliente_quinzena, p.ativo
		FROM cs_parceiro_franqueado pf
		JOIN cs_parceiro p ON p.id = pf.id_parceiro
		WHERE pf.id_representante = ? AND pf.id_franqueado = ? AND pf.ativo = 1 AND p.ativo = 1
		ORDER BY p.razao_social`, idRep, idFranq)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer rows.Close()
	list, err := scanParceirosCatalogo(rows)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func (h *Handler) ParceirosFranqueadoSalvar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	var in struct {
		IDRepresentante string   `json:"idRepresentante"`
		IDFranqueado    string   `json:"idFranqueado"`
		IDsParceiro     []string `json:"idsParceiro"`
	}
	if err := decodeJSON(r, &in); err != nil {
		writeErr(w, http.StatusBadRequest, "json invalido")
		return
	}
	in.IDRepresentante = strings.TrimSpace(in.IDRepresentante)
	in.IDFranqueado = strings.TrimSpace(in.IDFranqueado)
	if in.IDRepresentante == "" || in.IDFranqueado == "" {
		writeErr(w, http.StatusBadRequest, "idRepresentante e idFranqueado obrigatorios")
		return
	}
	ids := normalizeIDs(in.IDsParceiro)
	if err := validarParceirosRepPermitidos(in.IDRepresentante, ids); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	tx, err := db.Conn.Begin()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM cs_parceiro_franqueado WHERE id_representante = ? AND id_franqueado = ?`,
		in.IDRepresentante, in.IDFranqueado); err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	for _, idParceiro := range ids {
		if _, err := tx.Exec(`
			INSERT INTO cs_parceiro_franqueado (id_representante, id_franqueado, id_parceiro, ativo)
			VALUES (?, ?, ?, 1)`, in.IDRepresentante, in.IDFranqueado, idParceiro); err != nil {
			writeErr(w, http.StatusInternalServerError, "db")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "total": len(ids)})
}

func (h *Handler) ParceirosCatalogoFranqueado(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method")
		return
	}
	if !auth.RequireAPIKey(r) {
		writeErr(w, http.StatusUnauthorized, "X-Api-Key invalido")
		return
	}
	idFranq := strings.TrimSpace(r.URL.Query().Get("idFranqueado"))
	if idFranq == "" {
		writeErr(w, http.StatusBadRequest, "idFranqueado obrigatorio")
		return
	}
	list, err := parceirosCatalogoFranqueado(idFranq)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "dados": list})
}

func parceirosCatalogoFranqueado(idFranqueado string) ([]map[string]any, error) {
	rows, err := db.Conn.Query(`
		SELECT p.id, p.razao_social, p.nome_fantasia, p.software, p.preco_cliente_quinzena, p.ativo
		FROM cs_parceiro_franqueado pf
		JOIN cs_parceiro p ON p.id = pf.id_parceiro
		WHERE pf.id_franqueado = ? AND pf.ativo = 1 AND p.ativo = 1
		ORDER BY p.razao_social`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParceirosCatalogo(rows)
}

func parceiroPermitidoFranqueado(idFranqueado, idParceiro string) (bool, error) {
	var n int
	err := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM cs_parceiro_franqueado pf
		JOIN cs_parceiro p ON p.id = pf.id_parceiro
		WHERE pf.id_franqueado = ? AND pf.id_parceiro = ? AND pf.ativo = 1 AND p.ativo = 1`,
		idFranqueado, idParceiro).Scan(&n)
	return n > 0, err
}

func listarParceirosAtivos() ([]map[string]any, error) {
	rows, err := db.Conn.Query(`
		SELECT id, razao_social, nome_fantasia, software, preco_cliente_quinzena, ativo
		FROM cs_parceiro WHERE ativo = 1 ORDER BY razao_social`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParceirosCatalogo(rows)
}

func scanParceirosCatalogo(rows *sql.Rows) ([]map[string]any, error) {
	list := []map[string]any{}
	for rows.Next() {
		var id, razao, software string
		var fantasia sql.NullString
		var preco float64
		var ativo bool
		if err := rows.Scan(&id, &razao, &fantasia, &software, &preco, &ativo); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"id":                   id,
			"razaoSocial":          razao,
			"nomeFantasia":         fantasia.String,
			"software":             software,
			"precoClienteQuinzena": preco,
			"ativo":                ativo,
		})
	}
	return list, rows.Err()
}

func normalizeIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func validarParceirosExistem(ids []string) error {
	for _, id := range ids {
		var ativo bool
		err := db.Conn.QueryRow(`SELECT ativo FROM cs_parceiro WHERE id = ?`, id).Scan(&ativo)
		if err == sql.ErrNoRows {
			return errors.New("parceiro invalido ou inativo: " + id)
		}
		if err != nil {
			return err
		}
		if !ativo {
			return errors.New("parceiro invalido ou inativo: " + id)
		}
	}
	return nil
}

func validarParceirosRepPermitidos(idRepresentante string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := validarParceirosExistem(ids); err != nil {
		return err
	}
	for _, idParceiro := range ids {
		var n int
		err := db.Conn.QueryRow(`
			SELECT COUNT(*) FROM cs_parceiro_rep
			WHERE id_representante = ? AND id_parceiro = ? AND ativo = 1`,
			idRepresentante, idParceiro).Scan(&n)
		if err != nil {
			return err
		}
		if n == 0 {
			return errors.New("parceiro nao liberado pela central para este representante: " + idParceiro)
		}
	}
	return nil
}
