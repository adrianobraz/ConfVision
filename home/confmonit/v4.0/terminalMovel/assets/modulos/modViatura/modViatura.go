package modViatura

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
		Uri:    "/modViatura/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdFranqueado string `json:"idFranqueado"`
		IdViatura    string `json:"idViatura"`
		Nome         string `json:"nome"`
		Modelo       string `json:"modelo"`
		Placa        string `json:"placa"`
		Cor          string `json:"cor"`
		Telefone     string `json:"telefone"`
		Celular1     string `json:"celular1"`
		Celular2     string `json:"celular2"`
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
			viaturas.ID_Viatura,
			viaturas.Nome,
			viaturas.Modelo,
			viaturas.Placa,
			viaturas.Cor,
			viaturas.Telefone1,
			viaturas.Telefone2,
			viaturas.Telefone3
			
		FROM viaturas 

		WHERE viaturas.ID_Franqueado = ?
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
			idViatura sql.NullString
			nome      sql.NullString
			modelo    sql.NullString
			placa     sql.NullString
			cor       sql.NullString
			telefone  sql.NullString
			celular1  sql.NullString
			celular2  sql.NullString
		)

		if erro := tab.Scan(
			&idViatura,
			&nome,
			&modelo,
			&placa,
			&cor,
			&telefone,
			&celular1,
			&celular2,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.IdViatura = idViatura.String
		item.Nome = nome.String
		item.Modelo = modelo.String
		item.Placa = placa.String
		item.Cor = cor.String
		item.Telefone = telefone.String
		item.Celular1 = celular1.String
		item.Celular2 = celular2.String

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}
