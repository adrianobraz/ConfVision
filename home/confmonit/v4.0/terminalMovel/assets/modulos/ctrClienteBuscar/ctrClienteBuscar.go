package ctrClienteBuscar

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/ctrClienteBuscar/filtrar",
		Metodo: http.MethodPost,
		Funcao: filtrar,
	},
}

func filtrar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		Filtro        string `json:"filtro"`
		Valor         string `json:"valor"`
		IdDispositivo string `json:"idDispositivo"`
		IdCliente     string `json:"idCliente"`
		Nome          string `json:"nome"`
		Telefone1     string `json:"telefone1"`
		Conta         string `json:"conta"`
		NomeFranq     string `json:"nomeFranq"`
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
	
	var filtro string
	if obj.Filtro == "nome" {
		filtro = fmt.Sprintf(`WHERE cliente.Nome LIKE '%s' ORDER BY cliente.Nome`, "%"+obj.Valor+"%")
	} else if obj.Filtro == "celular" {
		filtro = fmt.Sprintf(`WHERE cliente.Telefone1 like '%s' ORDER BY cliente.Nome`, "%"+obj.Valor+"%")
	} else if obj.Filtro == "conta" {
		filtro = fmt.Sprintf(`WHERE dispositivo.Conta like '%s' ORDER BY cliente.Nome`, "%"+obj.Valor+"%")
	} else {
		erro := errors.New("erro no filtro")
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	txtSql := fmt.Sprintf(`
		SELECT 
				dispositivo.ID_Dispositivo,
				dispositivo.Conta,

				cliente.ID_Cliente,
				cliente.Nome,
				cliente.Telefone1,
				
				franqueado.razaoSocial
			FROM dispositivo			

			LEFT JOIN cliente
			ON cliente.ID_Cliente = dispositivo.ID_Cliente

			LEFT JOIN franqueado
			ON franqueado.ID_Franqueado = cliente.ID_Franqueado
			%s		
	`, filtro)

	tab, erro := db.Query(txtSql)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto
	for tab.Next() {
		var (
			item          objeto
			idCliente     sql.NullString
			idDispositivo sql.NullString
			nomeCli       sql.NullString
			telefone1     sql.NullString
			conta         sql.NullString
			nomeFranq     sql.NullString
		)

		if erro := tab.Scan(
			&idDispositivo,
			&conta,
			&idCliente,
			&nomeCli,
			&telefone1,
			&nomeFranq,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.IdDispositivo = idDispositivo.String
		item.IdCliente = idCliente.String
		item.Nome = nomeCli.String
		item.Telefone1 = telefone1.String
		item.Conta = conta.String
		item.NomeFranq = nomeFranq.String

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}
