package proEventoDetalhe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/proEventoDetalhe/carregarTabela",
		Metodo: http.MethodPost,
		Funcao: carregarTabela,
	},
}

type objeto struct {
	Quantidade string `json:"quantidade"`
	IdProcesso string `json:"idProcesso"`
	Codigo     string `json:"codigo"`
	Particao   string `json:"particao"`
	ZonaUser   string `json:"zonaUser"`
	Descricao  string `json:"descricao"`
	Nivel      string `json:"nivel"`
}

func carregarTabela(w http.ResponseWriter, r *http.Request) {

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
			COUNT(evento.ID_Evento) AS qtd, 
			evento.ID_Processo, 
			evento.Codigo, 
			evento.Particao, 
			evento.ZonaUser, 
			evento.Nivel, 
			processo.ID_Dispositivo 
		FROM evento 
		LEFT JOIN processo 
		ON evento.ID_Processo = processo.ID_Processo 
		WHERE evento.ID_Processo = ? 
		GROUP BY evento.Codigo, evento.ZonaUser 
		ORDER BY evento.Nivel DESC	
	`, obj.IdProcesso)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto
	for tab.Next() {
		var (
			item   objeto
			idDisp sql.NullString
		)

		if erro := tab.Scan(
			&item.Quantidade,
			&item.IdProcesso,
			&item.Codigo,
			&item.Particao,
			&item.ZonaUser,
			&item.Nivel,
			&idDisp,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		// Busca o id do franqueado ===========================================
		var idFranq string
		if err := aux.GetIdFranqByIdDisp(idDisp.String, &idFranq); err != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		// Busca a descricao do contacid ======================================
		if err := aux.GetCtiDecricao(
			idFranq, item.Codigo, &item.Descricao,
		); err != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		lista = append(lista, item)
	}

	// contactId.Codigo,
	// contactId.Descricao,

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}
