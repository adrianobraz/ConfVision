package login

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/setup"
	"terminal/src/templates"
	"terminal/src/tipos"

	"golang.org/x/crypto/bcrypt"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/",
		Metodo: http.MethodGet,
		Funcao: PaginaLogin,
	},
	{
		Uri:    "/login",
		Metodo: http.MethodGet,
		Funcao: PaginaLogin,
	},
	{
		Uri:    "/login/logar",
		Metodo: http.MethodPost,
		Funcao: logar,
	},
}

func PaginaLogin(w http.ResponseWriter, r *http.Request) {

	templates.ExecutarTemplate(w, "login.html", nil)
}

func logar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {

		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {

		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT
			usuarios.ID_Vinculo,
			usuarios.ID_Usuario,
			usuarios.Nome,
			usuarios.Nick,
			usuarios.Email1,
			usuarios.Senha,
			usuarios.Telefone1 ,
			usuarios.UsuarioTeminal,
			usuarios.Master,

			central.VoipToken,
			central.VoipKey,
			central.voipDeviceId,
			central.VoipAtivo,

			central.BenuvemEmail,
			central.BenuvemSenha,
			central.BenuvemAtivo,
			
			central.ID_Central,
			central.RazaoSocial,

			representante.ID_Representante,
			representante.RazaoSocial,
			
			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,
			franqueado.CodBenuvem

		FROM usuarios

		LEFT JOIN central
		ON usuarios.ID_Vinculo = central.ID_Central 

		LEFT JOIN representante
		ON usuarios.ID_Vinculo = representante.ID_Representante 

		LEFT JOIN franqueado
		ON usuarios.ID_Vinculo = franqueado.ID_Franqueado
		
		WHERE usuarios.Email1 = ?

		AND usuarios.UsuarioTeminal = 'S'
	`, obj.Email)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	if tab.Next() {
		var login struct {
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

			VoipToken    string `json:"voipToken"`
			VoipKey      string `json:"voipKey"`
			VoipDeviceId string `json:"voipDeviceId"`
			VoipAtivo    string `json:"voipAtivo"`

			BenuvemEmail string `json:"benuvemEmail"`
			BenuvemSenha string `json:"benuvemSenha"`
			BenuvemAtivo string `json:"benuvemAtivo"`
		}

		var (
			senhaComHash    string
			usuarioTerminal sql.NullString

			voipToken    sql.NullString
			voipKey      sql.NullString
			voipDeviceId sql.NullString
			voipAtvio    sql.NullString

			benuvemEmail sql.NullString
			benuvemSenha sql.NullString
			benuvemAtivo sql.NullString

			cenId   sql.NullString
			cenNome sql.NullString

			repId   sql.NullString
			repNome sql.NullString

			fraId      sql.NullString
			fraNome    sql.NullString
			benuvemFra sql.NullString
		)

		if erro := tab.Scan(
			&login.IdVinculo,
			&login.IdOperador,
			&login.UserNome,
			&login.UserNick,
			&login.UserEmail,
			&senhaComHash,
			&login.UserCelular,
			&usuarioTerminal,
			&login.UserMaster,

			&voipToken,
			&voipKey,
			&voipDeviceId,
			&voipAtvio,

			&benuvemEmail,
			&benuvemSenha,
			&benuvemAtivo,

			&cenId,
			&cenNome,

			&repId,
			&repNome,

			&fraId,
			&fraNome,
			&benuvemFra,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		login.VoipToken = voipToken.String
		login.VoipKey = voipKey.String
		login.VoipDeviceId = voipDeviceId.String
		login.VoipAtivo = voipAtvio.String

		login.BenuvemEmail = benuvemEmail.String
		login.BenuvemSenha = benuvemSenha.String
		login.BenuvemAtivo = benuvemAtivo.String

		// Verifica se o usuario pode usar o terminal
		if usuarioTerminal.String == "N" {
			err := errors.New("usuario não autorizado a acessar o terminal")
			aux.RespostaErro(w, http.StatusBadRequest, err)
		}

		// Verifica se usuario esta entre os grupo que podem usar o terminal
		if cenId.Valid {
			login.UserBenuvem = setup.CodBenuvemCentral
			login.UserTipo = "CEN"
			login.NomeVinculo = cenNome.String
		} else if repId.Valid {
			login.UserBenuvem = "00001" // inferir o codigo benuvem no cadastro do franqueado
			login.UserTipo = "REP"
			login.NomeVinculo = repNome.String
		} else if fraId.Valid {
			login.UserBenuvem = benuvemFra.String
			login.UserTipo = "FRA"
			login.NomeVinculo = fraNome.String
		} else {
			err := errors.New("usuario não autorizado a acessar o terminal")
			aux.RespostaErro(w, http.StatusBadRequest, err)
		}

		if erro := bcrypt.CompareHashAndPassword([]byte(senhaComHash), []byte(obj.Senha)); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		aux.RespostaJsonDados(w, http.StatusOK, login)
		return
	}

	erro = errors.New("não encontrado")
	aux.RespostaErro(w, http.StatusBadRequest, erro)

}
