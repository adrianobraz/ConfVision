package whitelabel

import (
	"encoding/json"
	"io"
	"net/http"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
	"franqueadopro/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-whitelabel",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/whitelabelCarregar",
		Metodo: http.MethodPost,
		Funcao: whitelabelCarregar,
		Aberto: false,
	},
	{
		URI:    "/whitelabelSalvar",
		Metodo: http.MethodPost,
		Funcao: whitelabelSalvar,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "White Label"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-gestao"
	auxiliar.PreencherEhMaster(r, &d)
	auxiliar.ExecutarTemplate(w, "whitelabel.html", d)
}

func whitelabelCarregar(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_whitelabel_get")
}

func whitelabelSalvar(w http.ResponseWriter, r *http.Request) {
	_, payload, err := lerPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	raw, err := xanopro.Post("/fp_whitelabel_salvar", payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func lerPayload(r *http.Request) (map[string]string, map[string]any, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["id_franqueado"] = cookie["idFranqueado"]
	return cookie, payload, nil
}

func proxyPro(w http.ResponseWriter, r *http.Request, path string) {
	_, payload, err := lerPayload(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}

	raw, err := xanopro.Post(path, payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}
