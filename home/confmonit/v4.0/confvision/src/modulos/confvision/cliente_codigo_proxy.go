package confvision

import (
	"confvision/src/auxiliar"
	"confvision/src/modulos/visdata"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func ProxyClienteExtGet(w http.ResponseWriter, r *http.Request) {
	idFra := idFranqueadoSessao(r)
	idCli := strings.TrimSpace(r.URL.Query().Get("id_cliente"))
	if idCli == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errMsg("id_cliente obrigatorio"))
		return
	}
	codigo, err := visdata.GetClienteCodigoInterno(r.Context(), idFra, idCli)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}
	auxiliar.RespostaJSON(w, http.StatusOK, map[string]any{
		"id_franqueado":  idFra,
		"id_cliente":     idCli,
		"codigo_interno": codigo,
	})
}

func ProxyClienteExtPut(w http.ResponseWriter, r *http.Request) {
	idFra := idFranqueadoSessao(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	payload := map[string]any{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
	}
	idCli := strings.TrimSpace(strVal(payload, "id_cliente"))
	if idCli == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errMsg("id_cliente obrigatorio"))
		return
	}
	codigo := strVal(payload, "codigo_interno")
	if err := visdata.SetClienteCodigoInterno(r.Context(), idFra, idCli, codigo); err != nil {
		auxiliar.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}
	saved, err := visdata.GetClienteCodigoInterno(r.Context(), idFra, idCli)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}
	auxiliar.RespostaJSON(w, http.StatusOK, map[string]any{
		"id_franqueado":  idFra,
		"id_cliente":     idCli,
		"codigo_interno": saved,
	})
}

func strVal(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(toString(v))
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		s := strings.TrimSpace(string(b))
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
		return s
	}
}

func errMsg(msg string) error {
	return &simpleErr{msg: msg}
}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }
