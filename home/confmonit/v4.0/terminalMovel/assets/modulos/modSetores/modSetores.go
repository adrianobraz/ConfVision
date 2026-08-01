package modSetores

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
		Uri:    "/modSetores/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		Particao      string `json:"particao"`
		IdDispositivo string `json:"idDispositivo"`
		Camera        string `json:"camera"`
		Setor         string `json:"setor"`
		Descricao     string `json:"descricao"`
		Nome          string `json:"nome"`
		Tipo          string `json:"tipo"`
		Conta         string `json:"conta"`
		CodFranq      string `json:"codFranq"`
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
			setorAlarme.Particao,
			setorAlarme.Camera,
			setorAlarme.Numero,
			setorAlarme.Descricao,
			setorAlarme.Nome,
			setorAlarme.Tipo,

			dispositivo.Conta,
			
			franqueado.CodBenuvem

		FROM setorAlarme
		
		LEFT JOIN dispositivo
		ON   setorAlarme.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
 		ON  cliente.ID_Franqueado = franqueado.ID_Franqueado

		WHERE setorAlarme.ID_Dispositivo = ?

	`, obj.IdDispositivo)

	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto

	for tab.Next() {
		var (
			item     objeto
			conta    sql.NullString
			codFranq sql.NullString
		)

		if erro := tab.Scan(
			&item.Particao,
			&item.Camera,
			&item.Setor,
			&item.Descricao,
			&item.Nome,
			&item.Tipo,
			&conta,
			&codFranq,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}
		item.Conta = conta.String
		item.CodFranq = codFranq.String
		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}
