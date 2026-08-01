package meuPlano

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/licenca"
	"franqueadopro/src/seguranca"
	"franqueadopro/src/xanopro"
)

var Rotas = []auxiliar.Rota{
	{
		URI:    "/carregar-meu-plano",
		Metodo: http.MethodGet,
		Funcao: carregarPagina,
		Aberto: false,
	},
	{
		URI:    "/licencaResumo",
		Metodo: http.MethodPost,
		Funcao: licencaResumo,
		Aberto: false,
	},
	{
		URI:    "/licencaMinhasFaturas",
		Metodo: http.MethodPost,
		Funcao: licencaMinhasFaturas,
		Aberto: false,
	},
	{
		URI:    "/licencaCatalogoPlanos",
		Metodo: http.MethodPost,
		Funcao: licencaCatalogoPlanos,
		Aberto: false,
	},
	{
		URI:    "/licencaContratar",
		Metodo: http.MethodPost,
		Funcao: licencaContratar,
		Aberto: false,
	},
	{
		URI:    "/licencaCupomValidar",
		Metodo: http.MethodPost,
		Funcao: licencaCupomValidar,
		Aberto: false,
	},
	{
		URI:    "/licencaCupomAplicarFatura",
		Metodo: http.MethodPost,
		Funcao: licencaCupomAplicarFatura,
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
	{
		URI:    "/licencaPacotesCota",
		Metodo: http.MethodPost,
		Funcao: licencaPacotesCota,
		Aberto: false,
	},
	{
		URI:    "/licencaComprarCota",
		Metodo: http.MethodPost,
		Funcao: licencaComprarCota,
		Aberto: false,
	},
}

func carregarPagina(w http.ResponseWriter, r *http.Request) {
	cookie, _ := seguranca.LerCookies(r)
	if cookie != nil && strings.TrimSpace(cookie["idFranqueado"]) != "" {
		raw, err := xanopro.Post("/fp_assinatura_resumo", map[string]string{
			"id_franqueado": cookie["idFranqueado"],
			"produto":       "franqueadopro",
		})
		if err == nil && len(raw) > 0 {
			licenca.SincronizarCacheResumo(cookie["idFranqueado"], "franqueadopro", raw)
		}
	}

	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Meu Plano"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-principal"
	auxiliar.PreencherEhMaster(r, &d)
	auxiliar.ExecutarTemplate(w, "meu-plano.html", d)
}

func licencaCatalogoPlanos(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_catalogo_listar_publico")
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaResumo(w http.ResponseWriter, r *http.Request) {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, err)
		return
	}
	raw, err := proxyProRaw(w, r, "/fp_assinatura_resumo")
	if err != nil || len(raw) == 0 {
		return
	}
	licenca.SincronizarCacheResumo(cookie["idFranqueado"], "franqueadopro", raw)
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaMinhasFaturas(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_fatura_minhas")
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaContratar(w http.ResponseWriter, r *http.Request) {
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

func licencaCupomValidar(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_cupom_validar")
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaCupomAplicarFatura(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_cupom_aplicar_fatura")
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

func licencaPacotesCota(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_pacote_cota_listar_publico")
	if err != nil || len(raw) == 0 {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func licencaComprarCota(w http.ResponseWriter, r *http.Request) {
	raw, err := proxyProRaw(w, r, "/fp_cota_comprar_extra")
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
		// Mensagem já vem limpa do client Xano (só o "message" amigável)
		auxiliar.RespostaErro(w, http.StatusBadGateway, err)
		return nil, err
	}
	return raw, nil
}
