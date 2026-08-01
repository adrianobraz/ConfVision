package login

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/config"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"

	"golang.org/x/crypto/bcrypt"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/",
		Metodo:   http.MethodGet,
		Controle: loginPage,
		Seguro:   false,
	},
	{
		Uri:      "/login",
		Metodo:   http.MethodGet,
		Controle: loginPage,
		Seguro:   false,
	},
	{
		Uri:      "/logar",
		Metodo:   http.MethodPost,
		Controle: logar,
		Seguro:   false,
	},
	{
		Uri:      "/logout",
		Metodo:   http.MethodGet,
		Controle: logout,
		Seguro:   false,
	},
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	if op, err := seguranca.LerOperador(r); err == nil && op["idOperador"] != "" {
		if seguranca.EhOperadorTerminal(r) {
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		// Cookie antigo (sem usuarioTerminal): limpa e exibe login
		seguranca.Deletar(w)
	}

	templ := []string{
		"public/templates/login/login.html",
		"public/templates/components/pagina.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	d := tipos.Page{
		Titulo: fmt.Sprintf("%s - Login Operador", config.TituloSite),
	}
	if err := page.ExecuteTemplate(w, "login.html", d); err != nil {
		fmt.Println(err)
	}
}

func logar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var obj objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
			usuarios.ID_Vinculo,
			usuarios.ID_Usuario,
			usuarios.Nome,
			usuarios.Nick,
			usuarios.Email1,
			usuarios.Senha,
			usuarios.Telefone1,
			usuarios.UsuarioTeminal,
			usuarios.Master,
			central.ID_Central,
			central.RazaoSocial,
			representante.ID_Representante,
			representante.RazaoSocial,
			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,
			franqueado.CodBenuvem
		FROM usuarios
		LEFT JOIN central ON usuarios.ID_Vinculo = central.ID_Central
		LEFT JOIN representante ON usuarios.ID_Vinculo = representante.ID_Representante
		LEFT JOIN franqueado ON usuarios.ID_Vinculo = franqueado.ID_Franqueado
		WHERE LOWER(usuarios.Email1) = LOWER(?)
		AND usuarios.UsuarioTeminal = 'S'
	`, strings.TrimSpace(obj.Email))
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	defer tab.Close()

	if !tab.Next() {
		resposta.Erro(w, http.StatusBadRequest, errors.New("usuario nao encontrado ou sem permissao de operador terminal"))
		return
	}

	var (
		loginData struct {
			IdVinculo   string `json:"idVinculo"`
			NomeVinculo string `json:"nomeVinculo"`
			IdOperador  string `json:"idOperador"`
			UserNome    string `json:"userNome"`
			UserNick    string `json:"userNick"`
			UserCelular string `json:"userCelular"`
			UserEmail   string `json:"userEmail"`
			UserMaster  string `json:"userMaster"`
			UserTipo    string `json:"userTipo"`
			UserBenuvem string `json:"userBenuvem"`
		}
		senhaComHash    string
		usuarioTerminal sql.NullString
		cenId           sql.NullString
		cenNome         sql.NullString
		repId           sql.NullString
		repNome         sql.NullString
		fraId           sql.NullString
		fraNome         sql.NullString
		benuvemFra      sql.NullString
	)

	if err := tab.Scan(
		&loginData.IdVinculo,
		&loginData.IdOperador,
		&loginData.UserNome,
		&loginData.UserNick,
		&loginData.UserEmail,
		&senhaComHash,
		&loginData.UserCelular,
		&usuarioTerminal,
		&loginData.UserMaster,
		&cenId,
		&cenNome,
		&repId,
		&repNome,
		&fraId,
		&fraNome,
		&benuvemFra,
	); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if usuarioTerminal.String != "S" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("usuario nao autorizado: requer UsuarioTerminal = S"))
		return
	}

	if cenId.Valid {
		loginData.UserTipo = "CEN"
		loginData.NomeVinculo = cenNome.String
	} else if repId.Valid {
		loginData.UserTipo = "REP"
		loginData.NomeVinculo = repNome.String
		loginData.UserBenuvem = "00001"
	} else if fraId.Valid {
		loginData.UserTipo = "FRA"
		loginData.NomeVinculo = fraNome.String
		loginData.UserBenuvem = benuvemFra.String
	} else {
		resposta.Erro(w, http.StatusBadRequest, errors.New("usuario nao autorizado"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(senhaComHash), []byte(obj.Senha)); err != nil {
		resposta.Erro(w, http.StatusBadRequest, errors.New("usuario ou senha invalidos"))
		return
	}

	if err := seguranca.SalvarOperador(w, map[string]string{
		"idOperador":       loginData.IdOperador,
		"idVinculo":        loginData.IdVinculo,
		"userTipo":         loginData.UserTipo,
		"userNome":         loginData.UserNome,
		"userNick":         loginData.UserNick,
		"nomeVinculo":      loginData.NomeVinculo,
		"userEmail":        loginData.UserEmail,
		"userMaster":       loginData.UserMaster,
		"usuarioTerminal":  "S",
	}); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.JsonDados(w, http.StatusOK, loginData)
}

func logout(w http.ResponseWriter, r *http.Request) {
	seguranca.Deletar(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}
