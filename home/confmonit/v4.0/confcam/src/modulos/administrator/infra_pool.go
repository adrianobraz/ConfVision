package administrator

import (
	"confvision/src/auxiliar"
	"confvision/src/conexao"
	"confvision/src/modulos/visdata"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var reNonDigit = regexp.MustCompile(`\D+`)

func ApiInfraPoolStatus(w http.ResponseWriter, r *http.Request) {
	idFra := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	if idFra == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errMsg("id_franqueado obrigatorio"))
		return
	}
	out, err := visdata.InfraPoolStatus(r.Context(), idFra)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	payload, _ := json.Marshal(map[string]any{"status": "ok", "dados": out})
	auxiliar.RespostaAPP(w, payload)
}

func ApiInfraMediamtxNodes(w http.ResponseWriter, r *http.Request) {
	out, err := visdata.ListMediamtxNodesAtivos(r.Context())
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	payload, _ := json.Marshal(map[string]any{"status": "ok", "dados": out})
	auxiliar.RespostaAPP(w, payload)
}

func ApiInfraPoolMediamtxSave(w http.ResponseWriter, r *http.Request) {
	payload, err := readJSONBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	out, err := visdata.UpsertFranqueadoMediamtxPool(r.Context(), payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	resp, _ := json.Marshal(map[string]any{"status": "ok", "dados": out})
	auxiliar.RespostaAPP(w, resp)
}

func ApiInfraPoolWorkerSave(w http.ResponseWriter, r *http.Request) {
	payload, err := readJSONBody(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	out, err := visdata.UpsertFranqueadoWorkerPool(r.Context(), payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	resp, _ := json.Marshal(map[string]any{"status": "ok", "dados": out})
	auxiliar.RespostaAPP(w, resp)
}

func ApiInfraFranqueadosBusca(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	lista, err := fetchFranqueadosInfraBusca(context.Background(), q)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	payload, _ := json.Marshal(map[string]any{"status": "ok", "dados": lista})
	auxiliar.RespostaAPP(w, payload)
}

func readJSONBody(r *http.Request) (map[string]any, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func errMsg(s string) error {
	return &simpleErr{msg: s}
}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }

func fetchFranqueadosInfraBusca(ctx context.Context, q string) ([]map[string]any, error) {
	_ = ctx
	db, err := conexao.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	qRaw := strings.TrimSpace(q)
	qLower := strings.ToLower(qRaw)
	qDigits := reNonDigit.ReplaceAllString(qRaw, "")
	sqlQ := `
SELECT f.ID_Franqueado, f.RazaoSocial, f.Cnpj, u.Email1
FROM franqueado f
LEFT JOIN usuarios u ON u.ID_Usuario = f.ID_UsuarioMaster
WHERE 1=1`
	var args []any
	if qRaw != "" {
		sqlQ += ` AND (
	LOWER(f.RazaoSocial) LIKE ?
	OR f.ID_Franqueado LIKE ?
	OR LOWER(COALESCE(f.Cnpj, '')) LIKE ?
	OR LOWER(COALESCE(u.Email1, '')) LIKE ?`
		like := "%" + qLower + "%"
		args = append(args, like, "%"+qRaw+"%", like, like)
		if len(qDigits) >= 3 {
			sqlQ += `
	OR REPLACE(REPLACE(REPLACE(COALESCE(f.Cnpj, ''), '.', ''), '/', ''), '-', '') LIKE ?`
			args = append(args, "%"+qDigits+"%")
		}
		sqlQ += `
)`
	}
	sqlQ += ` ORDER BY f.RazaoSocial ASC LIMIT 80`
	rows, err := db.Query(sqlQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, nome, cnpj, email sql.NullString
		if err := rows.Scan(&id, &nome, &cnpj, &email); err != nil {
			continue
		}
		k := strings.TrimSpace(id.String)
		if k == "" {
			continue
		}
		n := strings.TrimSpace(nome.String)
		c := strings.TrimSpace(cnpj.String)
		em := strings.TrimSpace(email.String)
		hintParts := make([]string, 0, 2)
		if c != "" {
			hintParts = append(hintParts, "CNPJ "+c)
		}
		if em != "" {
			hintParts = append(hintParts, em)
		}
		out = append(out, map[string]any{
			"id_franqueado": k,
			"nome":          n,
			"cnpj":          c,
			"email":         em,
			"label":         fmt.Sprintf("%s (%s)", n, k),
			"hint":          strings.Join(hintParts, " · "),
		})
	}
	return out, nil
}
