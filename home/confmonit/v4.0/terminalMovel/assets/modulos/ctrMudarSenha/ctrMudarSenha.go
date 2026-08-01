package ctrMudarSenha

import (
	"encoding/json"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"

	"golang.org/x/crypto/bcrypt"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/ctrMudarSenha/gravar",
		Metodo: http.MethodPost,
		Funcao: gravar,
	},
}

func gravar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdOperador string `json:"idOperador"`
		Senha      string `json:"senha"`
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

	stm, erro := db.Prepare(`
		UPDATE usuarios 
		SET  usuarios.Senha = ?
		WHERE usuarios.ID_Usuario = ?
	`)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer stm.Close()

	senhaHesh, erro := codHash(obj.Senha)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if _, erro := stm.Exec(string(senhaHesh), obj.IdOperador); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	aux.RespostaJsonOK(w)
}

// Hash recebe uma string e coloca um hash nela
func codHash(senha string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
}

// VerificarSenha compara uma senha e um hash e retorna se elas são iguais
// func VerificarSenha(senhaComHash, senhaString string) error {
// 	return bcrypt.CompareHashAndPassword([]byte(senhaComHash), []byte(senhaString))
//}
