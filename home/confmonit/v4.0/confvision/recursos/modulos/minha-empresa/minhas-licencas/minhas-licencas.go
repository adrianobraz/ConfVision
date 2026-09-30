package minhasLicencas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/modulos/confvision"
	"confvision/src/modulos/visdata"
	"confvision/src/xanopro"

	"github.com/joho/godotenv"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/minhas-licencas",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaResumo",
		Metodo: http.MethodPost,
		Funcao: licencaResumo,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaPlanos",
		Metodo: http.MethodPost,
		Funcao: licencaPlanos,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaComprar",
		Metodo: http.MethodPost,
		Funcao: licencaComprar,
		Aberto: false,
	},
	{
		URI:    "/cvLicencaFaturas",
		Metodo: http.MethodPost,
		Funcao: licencaFaturas,
		Aberto: false,
	},
	{
		URI:    "/cvCapacidadeResumo",
		Metodo: http.MethodPost,
		Funcao: capacidadeResumo,
		Aberto: false,
	},
	{
		URI:    "/cvCapacidadeCotacao",
		Metodo: http.MethodPost,
		Funcao: capacidadeCotacao,
		Aberto: false,
	},
	{
		URI:    "/cvCapacidadeContratar",
		Metodo: http.MethodPost,
		Funcao: capacidadeContratar,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina
		d.TituloSite = config.TituloSite
		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
		return
	}

	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Minhas licenças"
	d.LinkRetorno = "/carregar-menu-confvision"
	auxiliar.ExecutarTemplate(w, "minhas-licencas.html", d)
}

func licencaResumo(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if !config.VisPostgresEnabled {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, errPostgresOff)
		return
	}

	idFranqueado := idFranqueadoPayload(payload)
	xanoRes, xanoErr := confvision.FetchXanoResumoFranqueado(idFranqueado)

	faturas := []any{}
	if xanoErr == nil && xanoRes["faturas_abertas"] != nil {
		if arr, ok := xanoRes["faturas_abertas"].([]any); ok {
			faturas = arr
		}
	}

	out, err := visdata.ResumoPortalWithFaturas(r.Context(), idFranqueado, faturas)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if config.VisPostgresEnabled {
		if capRes, capErr := visdata.ResumoCapacidade(r.Context(), idFranqueado, strPayload(payload, "id_central"), strPayload(payload, "id_representante")); capErr == nil {
			out["capacidade"] = capRes
		}
	}

	if xanoErr == nil {
		if lib, ok := xanoRes["liberado"].(bool); ok && lib {
			out["liberado"] = true
		}
		if motivo, ok := xanoRes["motivo"].(string); ok && motivo != "" {
			out["motivo"] = motivo
		}
		if usa, ok := xanoRes["usa_confvision"].(string); ok && usa != "" {
			out["usa_confvision"] = usa
		}
	}

	writeJSON(w, out)
}

func licencaPlanos(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	idFranqueado := idFranqueadoPayload(payload)
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}

	raw, err := xanopro.Post("/fp_confvision_planos_listar_franqueado", map[string]any{
		"id_franqueado": idFranqueado,
	})
	if err != nil {
		if config.VisPostgresEnabled {
			writeJSON(w, visdata.ListPlanosCatalog(r.Context(), idFranqueado))
			return
		}
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, out)
}

func licencaComprar(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	idFranqueado := idFranqueadoPayload(payload)
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}

	itensRaw, _ := payload["itens"].([]any)
	if len(itensRaw) == 0 {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("itens obrigatorio"))
		return
	}

	if config.VisPostgresEnabled {
		if err := visdata.CheckCapacidadeDisponivel(r.Context(), idFranqueado, false); err != nil {
			auxiliar.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
	}

	body := map[string]any{
		"id_franqueado": idFranqueado,
		"itens":         itensRaw,
		"admin_usuario": strings.TrimSpace(strPayload(payload, "admin_usuario")),
	}

	raw, err := xanopro.Post("/fp_confvision_comprar", body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, out)
}

func licencaFaturas(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	idFranqueado := idFranqueadoPayload(payload)
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}

	req := map[string]any{"id_franqueado": idFranqueado}
	if st := strings.TrimSpace(strPayload(payload, "status")); st != "" {
		req["status"] = st
	}

	raw, err := xanopro.Post("/fp_fatura_minhas", req)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, out)
}

func capacidadeResumo(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if !config.VisPostgresEnabled {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, errPostgresOff)
		return
	}
	idFranqueado := idFranqueadoPayload(payload)
	if idFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("id_franqueado obrigatorio"))
		return
	}
	out, err := visdata.ResumoCapacidade(r.Context(), idFranqueado, strPayload(payload, "id_central"), strPayload(payload, "id_representante"))
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, out)
}

func capacidadeCotacao(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if !config.VisPostgresEnabled {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, errPostgresOff)
		return
	}
	idFranqueado := idFranqueadoPayload(payload)
	qtd := intPayload(payload, "quantidade")
	if qtd <= 0 {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("quantidade obrigatoria"))
		return
	}
	out, err := visdata.CotacaoCapacidade(r.Context(), idFranqueado, qtd, strPayload(payload, "id_central"), strPayload(payload, "id_representante"))
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, out)
}

func capacidadeContratar(w http.ResponseWriter, r *http.Request) {
	payload, err := readPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if !config.VisPostgresEnabled {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, errPostgresOff)
		return
	}
	idFranqueado := idFranqueadoPayload(payload)
	qtd := intPayload(payload, "quantidade")
	if qtd <= 0 {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("quantidade obrigatoria"))
		return
	}

	body := map[string]any{
		"id_franqueado": idFranqueado,
		"admin_usuario": strings.TrimSpace(strPayload(payload, "admin_usuario")),
		"itens": []map[string]any{{
			"plano":      "capacidade_processamento",
			"quantidade": qtd,
		}},
	}

	raw, err := xanopro.Post("/fp_confvision_comprar", body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, out)
}

func intPayload(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case string:
		var n int
		fmt.Sscan(strings.TrimSpace(t), &n)
		return n
	default:
		return 0
	}
}

func readPayload(r *http.Request) (map[string]any, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func idFranqueadoPayload(payload map[string]any) string {
	if v := strPayload(payload, "id_franqueado"); v != "" {
		return v
	}
	return strPayload(payload, "idFranqueado")
}

func strPayload(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(strings.Trim(fmtAny(v), `"`))
}

func fmtAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}

func asMapSlice(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, el := range t {
			if m, ok := el.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]any:
		return t
	default:
		return nil
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	raw, err := json.Marshal(v)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

var errPostgresOff = &postgresOffErr{}

type postgresOffErr struct{}

func (e *postgresOffErr) Error() string {
	return "Postgres operacional nao configurado (POSTGRES_URL)"
}
