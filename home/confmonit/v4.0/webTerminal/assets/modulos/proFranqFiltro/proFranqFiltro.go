package proFranqFiltro

import (
	"database/sql"
	"fmt"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/proFranqFiltro/carregarTabela",
		Metodo: http.MethodPost,
		Funcao: carregarTabela,
	},
}

func carregarTabela(w http.ResponseWriter, r *http.Request) {
	type pro struct {
		IdFranqueado   string `json:"idFranqueado"`
		NomeFranqueado string `json:"nomeFranqueado"`
		Quantidade     string `json:"quantidade"`
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT 
			franqueado.ID_Franqueado, 
			franqueado.NomeFantasia, 
			COUNT(processo.ID_Processo) AS quantidade 

		FROM processo 	

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado 
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado 

		WHERE processo.DataAtenFim IS NULL
		AND processo.Nivel > 0
		GROUP BY franqueado.ID_Franqueado		
	`)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []pro
	for tab.Next() {
		var (
			item pro
			nome sql.NullString
		)

		if erro := tab.Scan(
			&item.IdFranqueado,
			&nome,
			&item.Quantidade,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.NomeFranqueado = nome.String

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}

}
