package login

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"confvision/src/xanopro"

	"github.com/joho/godotenv"
)

// CarregarPaginaLogin carrega a pagina para efetuar o login
func CarregarLogin(w http.ResponseWriter, r *http.Request) {

	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
	} else {
		cookie, _ := seguranca.LerCookies(r)
		if cookie["token"] != "" {
			http.Redirect(w, r, "/carregar-menu-confvision", 302)
			return
		}

		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite
		aplicarWhitelabelLogin(&d, r.Host)

		auxiliar.ExecutarTemplate(w, "login.html", d)
	}
}

func normalizarHost(host string) string {
	h := strings.ToLower(strings.TrimSpace(host))
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	if i := strings.Index(h, "/"); i >= 0 {
		h = h[:i]
	}
	if i := strings.Index(h, ":"); i >= 0 {
		h = h[:i]
	}
	return h
}

func aplicarWhitelabelLogin(d *auxiliar.Pagina, hostRaw string) {
	host := normalizarHost(hostRaw)
	if host == "" || host == "localhost" || host == "127.0.0.1" {
		return
	}
	if config.XanoApiPro == "" {
		return
	}

	raw, err := xanopro.Post("/fp_whitelabel_by_fqdn", map[string]any{"fqdn": host})
	if err != nil {
		return
	}

	var resp struct {
		Dados *struct {
			IDFranqueado string `json:"id_franqueado"`
			LogoData     string `json:"logo_data"`
		} `json:"dados"`
	}
	if json.Unmarshal(raw, &resp) != nil || resp.Dados == nil {
		return
	}
	d.IdFranqueadoWhitelabel = strings.TrimSpace(resp.Dados.IDFranqueado)
	_ = resp.Dados.LogoData
}

func resolverTenantDominio(r *http.Request, informado string) string {
	tenantDominio := strings.TrimSpace(informado)
	if tenantDominio != "" {
		return tenantDominio
	}
	host := normalizarHost(r.Host)
	if host == "" || config.XanoApiPro == "" {
		return ""
	}
	raw, err := xanopro.Post("/fp_whitelabel_by_fqdn", map[string]any{"fqdn": host})
	if err != nil {
		return ""
	}
	var resp struct {
		Dados *struct {
			IDFranqueado string `json:"id_franqueado"`
		} `json:"dados"`
	}
	if json.Unmarshal(raw, &resp) == nil && resp.Dados != nil {
		return strings.TrimSpace(resp.Dados.IDFranqueado)
	}
	return ""
}

func postJSON(url string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, raw, nil
}

// LoginLogar efetua o login no sistema (franqueado FRA ou cliente CLI)
func LoginLogar(w http.ResponseWriter, r *http.Request) {

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var reqLogin struct {
		Email               string `json:"email"`
		Senha               string `json:"senha"`
		IDFranqueadoDominio string `json:"id_franqueado_dominio"`
	}
	_ = json.Unmarshal(body, &reqLogin)

	email := strings.TrimSpace(reqLogin.Email)
	senha := strings.TrimSpace(reqLogin.Senha)
	if email == "" || senha == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("email e senha obrigatorios"))
		return
	}

	tenantDominio := resolverTenantDominio(r, reqLogin.IDFranqueadoDominio)

	// 1) Tenta login como franqueado
	switch tentarLoginFranqueado(w, email, senha, tenantDominio) {
	case loginResponded:
		return
	case loginOK:
		return
	}

	// 2) Tenta login como cliente
	switch tentarLoginCliente(w, email, senha, tenantDominio) {
	case loginResponded:
		return
	case loginOK:
		return
	}

	auxiliar.RespostaErro(w, http.StatusUnauthorized, errors.New("usuario ou senha invalidos"))
}

type loginOutcome int

const (
	loginSkip loginOutcome = iota
	loginOK
	loginResponded
)

func tentarLoginFranqueado(w http.ResponseWriter, email, senha, tenantDominio string) loginOutcome {
	status, corpo, erro := postJSON(fmt.Sprintf("%s/v4/franqueado/logar", config.ApiUrl), map[string]string{
		"email": email,
		"senha": senha,
	})
	if erro != nil || status >= 400 {
		return loginSkip
	}

	var login struct {
		Dados struct {
			Token     string `json:"token"`
			IdUsuario string `json:"idUsuario"`
			IdVinculo string `json:"idVinculo"`
			Tipo      string `json:"tipo"`
			Nome      string `json:"nome"`
			Nick      string `json:"nick"`
			Email1    string `json:"email1"`
		} `json:"dados"`
	}
	if json.Unmarshal(corpo, &login) != nil {
		return loginSkip
	}
	if login.Dados.Tipo != "FRA" || strings.TrimSpace(login.Dados.IdVinculo) == "" {
		return loginSkip
	}

	if tenantDominio != "" && tenantDominio != strings.TrimSpace(login.Dados.IdVinculo) {
		auxiliar.RespostaErro(w, http.StatusForbidden, errors.New("usuario nao autorizado neste dominio"))
		return loginResponded
	}

	// Fonte canonica Xano: plano ConfVision, incluso no FP Pro+ ou licenca de camera
	okAcesso, erro := confVisionAcessoPermitido(login.Dados.IdVinculo)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return loginResponded
	}
	if !okAcesso {
		auxiliar.RespostaErro(w, http.StatusForbidden, errors.New("modulo ConfVision nao contratado — precisa de plano ConfVision, Pro+ do FranqueadoPro ou licenca de camera"))
		return loginResponded
	}

	if erro = seguranca.SalvarSessao(w, seguranca.Sessao{
		IdUsuario: login.Dados.IdUsuario,
		IdVinculo: login.Dados.IdVinculo,
		Token:     login.Dados.Token,
		Tipo:      "FRA",
	}); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return loginResponded
	}

	auxiliar.RespostaAPP(w, corpo)
	return loginOK
}

func tentarLoginCliente(w http.ResponseWriter, email, senha, tenantDominio string) loginOutcome {
	status, corpo, erro := postJSON(fmt.Sprintf("%s/v4/cliente/logar", config.ApiUrl), map[string]string{
		"email1": email,
		"senha":  senha,
	})
	if erro != nil || status >= 400 {
		return loginSkip
	}

	var root map[string]any
	if json.Unmarshal(corpo, &root) != nil {
		return loginSkip
	}
	dados, _ := root["dados"].(map[string]any)
	if dados == nil {
		return loginSkip
	}

	token := anyToTrimString(dados["token"])
	idUsuario := anyToTrimString(dados["idUsuario"])
	idCliente := anyToTrimString(dados["idCliente"])
	idFranqueado := anyToTrimString(dados["idFranqueado"])
	nome := anyToTrimString(dados["nome"])
	nick := anyToTrimString(dados["nick"])
	email1 := anyToTrimString(dados["email1"])
	fraRazao := anyToTrimString(dados["fraRazao"])
	ativo := anyToTrimString(dados["ativo"])

	if idCliente == "" || idFranqueado == "" || token == "" {
		return loginSkip
	}

	if strings.ToUpper(ativo) == "N" {
		auxiliar.RespostaErro(w, http.StatusForbidden, errors.New("usuario bloqueado"))
		return loginResponded
	}

	if tenantDominio != "" && tenantDominio != idFranqueado {
		auxiliar.RespostaErro(w, http.StatusForbidden, errors.New("usuario nao autorizado neste dominio"))
		return loginResponded
	}

	usaConfVision, erro := getUsaConfVision(idFranqueado)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return loginResponded
	}
	if usaConfVision != "S" {
		auxiliar.RespostaErro(w, http.StatusForbidden, errors.New("modulo ConfVision nao contratado"))
		return loginResponded
	}

	if erro = seguranca.SalvarSessao(w, seguranca.Sessao{
		IdUsuario: idUsuario,
		IdVinculo: idFranqueado,
		Token:     token,
		Tipo:      "CLI",
		IdCliente: idCliente,
	}); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return loginResponded
	}

	out, _ := json.Marshal(map[string]any{
		"dados": map[string]any{
			"tipo":           "CLI",
			"token":          token,
			"idUsuario":      idUsuario,
			"idVinculo":      idFranqueado,
			"idCliente":      idCliente,
			"idFranqueado":   idFranqueado,
			"nome":           nome,
			"nick":           nick,
			"email1":         email1,
			"nomeFranqueado": fraRazao,
		},
	})
	auxiliar.RespostaAPP(w, out)
	return loginOK
}

func anyToTrimString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSpace(fmt.Sprintf("%.0f", t))
	case json.Number:
		return strings.TrimSpace(t.String())
	case bool:
		if t {
			return "S"
		}
		return "N"
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "<nil>" {
			return ""
		}
		return s
	}
}

func CarregarDados(w http.ResponseWriter, r *http.Request) {
	json, erro := seguranca.CarregarDados(r)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, json)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := seguranca.LerCookies(r)
	eraAdministrator := seguranca.EhAdministrator(cookie)
	seguranca.Deletar(w)
	if eraAdministrator {
		http.Redirect(w, r, "/administrator", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func getFranqDadosById(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf(`%s/v4/franqueado/getDadosById`, config.ApiUrl)

	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	corpo, erro := io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaAPP(w, corpo)
}
