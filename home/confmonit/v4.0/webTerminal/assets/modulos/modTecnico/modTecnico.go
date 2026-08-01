package modTecnico

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
		Uri:    "/modTecnico/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdFranqueado string `json:"idFranqueado"`
		IdTecnico    string `json:"idTecnico"`
		Nome         string `json:"nome"`
		Celular      string `json:"celular"`
		Telefone     string `json:"telefone"`
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
		 tecnico.ID_Tecnico,
		 tecnico.Nome,
		 tecnico.Telefone1,
		 tecnico.Telefone2
			
		FROM tecnico
 
		WHERE tecnico.ID_Vinculo = ?
	`, obj.IdFranqueado)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto

	for tab.Next() {
		var (
			item      objeto
			idTecnico sql.NullString
			nome      sql.NullString
			celular   sql.NullString
			telefone  sql.NullString
		)

		if erro := tab.Scan(
			&idTecnico,
			&nome,
			&celular,
			&telefone,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.IdTecnico = idTecnico.String
		item.Nome = nome.String
		item.Celular = celular.String
		item.Telefone = telefone.String

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}

}
