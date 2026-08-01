package administrator

import (
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
)

func CarregarLogin(w http.ResponseWriter, r *http.Request) {
	if cookie, err := seguranca.LerCookies(r); err == nil && seguranca.EhAdministrator(cookie) {
		http.Redirect(w, r, "/administrator/rtmp-falhas", http.StatusFound)
		return
	}

	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if res["MANUTENCAO"] == "S" {
		d := auxiliar.Pagina{TituloSite: config.TituloSite}
		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
		return
	}

	d := auxiliar.Pagina{TituloSite: config.TituloSite}
	auxiliar.ExecutarTemplate(w, "administrator.html", d)
}

func Login(w http.ResponseWriter, r *http.Request) {
	if config.AdministratorUser == "" || config.AdministratorPass == "" {
		auxiliar.RespostaErro(w, http.StatusServiceUnavailable, errors.New("administrator nao configurado no servidor"))
		return
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var req struct {
		Usuario string `json:"usuario"`
		Senha   string `json:"senha"`
	}
	_ = json.Unmarshal(body, &req)

	usuario := strings.TrimSpace(req.Usuario)
	senha := strings.TrimSpace(req.Senha)
	if usuario == "" || senha == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("usuario e senha obrigatorios"))
		return
	}

	if !credenciaisAdministratorValidas(usuario, senha) {
		auxiliar.RespostaErro(w, http.StatusUnauthorized, errors.New("usuario ou senha invalidos"))
		return
	}

	if erro := seguranca.SalvarSessao(w, seguranca.Sessao{
		IdUsuario: "administrator",
		IdVinculo: "",
		Token:     "administrator-rtmp",
		Tipo:      "ADM",
	}); erro != nil {
		auxiliar.RespostaErro(w, http.StatusInternalServerError, erro)
		return
	}

	out, _ := json.Marshal(map[string]any{
		"status": "ok",
		"dados": map[string]any{
			"tipo":     "ADM",
			"redirect": "/administrator/rtmp-falhas",
		},
	})
	auxiliar.RespostaAPP(w, out)
}

func credenciaisAdministratorValidas(usuario, senha string) bool {
	uOk := subtle.ConstantTimeCompare([]byte(usuario), []byte(config.AdministratorUser)) == 1
	pOk := subtle.ConstantTimeCompare([]byte(senha), []byte(config.AdministratorPass)) == 1
	return uOk && pOk
}

func CarregarRtmpFalhas(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if res["MANUTENCAO"] == "S" {
		d := auxiliar.Pagina{TituloSite: config.TituloSite}
		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
		return
	}

	d := auxiliar.Pagina{TituloSite: config.TituloSite, NomeTela: "Falhas RTMP"}
	auxiliar.ExecutarTemplate(w, "administrator-rtmp.html", d)
}
