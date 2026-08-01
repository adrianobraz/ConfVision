package modDiscador

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/modDiscador/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdDispositivo string `json:"idDispositivo"`
		SenhaVerbal   string `json:"senhaVerbal"`
		ContraSenha   string `json:"contraSenha"`
		CliTelefone1  string `json:"cliTelefone1"`
		CliTelefone2  string `json:"cliTelefone2"`
		Supervisor    string `json:"supervisor"`
		SupTelefone1  string `json:"supTelefone1"`
		SupTelefone2  string `json:"supTelefone2"`
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
	// cliente.senhaVerbal,
	// cliente.contraSenha,
	tab, erro := db.Query(`
		SELECT			
			
			dispositivo.SenhaVerbal,
			dispositivo.ContraSenhaVerbal,
			cliente.Telefone1,
			cliente.telefone2,
			
			usuarios.Nome,
			usuarios.Telefone1,
			usuarios.Telefone2

		FROM dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
		ON  cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN usuarios
		ON  franqueado.ID_UsuarioMaster = usuarios.ID_Usuario

		WHERE dispositivo.ID_Dispositivo = ?
	`, obj.IdDispositivo)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	if tab.Next() {
		var (
			supervisor   sql.NullString
			supTelefone1 sql.NullString
			supTelefone2 sql.NullString
		)
		if erro := tab.Scan(

			&obj.SenhaVerbal,
			&obj.ContraSenha,
			&obj.CliTelefone1,
			&obj.CliTelefone2,
			&supervisor,
			&supTelefone1,
			&supTelefone2,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		obj.Supervisor = supervisor.String

		obj.SupTelefone1 = supTelefone1.String

		obj.SupTelefone2 = supTelefone2.String
		aux.RespostaJsonDados(w, http.StatusOK, obj)
		return
	}

	aux.RespostaJsonVazio(w)
}
