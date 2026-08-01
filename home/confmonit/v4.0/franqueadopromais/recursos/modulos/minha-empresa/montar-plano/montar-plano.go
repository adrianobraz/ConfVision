package montarPlano

import (
	"encoding/json"
	"io"
	"net/http"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/licenca"
	"franqueadopro/src/seguranca"
	"franqueadopro/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-montar-plano",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/licencaCatalogoPlanos",
		Metodo: http.MethodPost,
		Funcao: licencaCatalogoPlanos,
		Aberto: false,
	},
	{
		URI:    "/licencaCatalogoModulos",
		Metodo: http.MethodPost,
		Funcao: licencaCatalogoModulos,
		Aberto: false,
	},
	{
		URI:    "/licencaAlacarteSimular",
		Metodo: http.MethodPost,
		Funcao: licencaAlacarteSimular,
		Aberto: false,
	},
	{
		URI:    "/licencaContratarAlacarte",
		Metodo: http.MethodPost,
		Funcao: licencaContratarAlacarte,
		Aberto: false,
	},
	{
		URI:    "/licencaEcossistema",
		Metodo: http.MethodPost,
		Funcao: licencaEcossistema,
		Aberto: false,
	},
	{
		URI:    "/licencaContratarProduto",
		Metodo: http.MethodPost,
		Funcao: licencaContratarProduto,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Montar Meu Plano"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-gestao"
	auxiliar.PreencherEhMaster(r, &d)
	auxiliar.ExecutarTemplate(w, "montar-plano.html", d)
}

func licencaCatalogoPlanos(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_catalogo_listar_publico")
}

func licencaCatalogoModulos(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_modulo_catalogo_listar_publico")
}

func licencaAlacarteSimular(w http.ResponseWriter, r *http.Request) {
	proxyPro(w, r, "/fp_alacarte_simular")
}

func licencaContratarAlacarte(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_assinatura_contratar_alacarte")
	if err != nil || len(raw) == 0 {
		return
	}
	cookie, _ := seguranca.LerCookies(r)
	if cookie != nil {
		licenca.InvalidarCache(cookie["idFranqueado"], "franqueadopro")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaEcossistema(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_ecossistema_listar")
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaContratarProduto(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_assinatura_contratar")
	if err != nil || len(raw) == 0 {
		return
	}
	cookie, _ := seguranca.LerCookies(r)
	if cookie != nil {
		licenca.InvalidarCache(cookie["idFranqueado"], "franqueadopro")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func proxyPro(w http.ResponseWriter, r *http.Request, path string) {
	raw, err := proxyProRaw(w, r, path)
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func proxyProRaw(w http.ResponseWriter, r *http.Request, path string) ([]byte, error) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return nil, err
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, err)
		return nil, err
	}

	var payload map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	payload["id_franqueado"] = cookie["idFranqueado"]

	raw, err := xanopro.Post(path, payload)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return nil, err
	}
	return raw, nil
}
