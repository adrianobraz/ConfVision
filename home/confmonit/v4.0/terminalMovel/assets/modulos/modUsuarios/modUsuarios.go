package modUsuarios

import (
	"encoding/json"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/modUsuarios/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdDispositivo string `json:"idDispositivo"`
		Codigo        string `json:"codigo"`
		Nome          string `json:"nome"`
		Celular       string `json:"celular"`
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
			usuariosAlarme.codigo,
			usuariosAlarme.nome,
			usuariosAlarme.celular

		FROM  usuariosAlarme

		WHERE  usuariosAlarme.ID_Dispositivo = ?

	`, obj.IdDispositivo)

	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto

	for tab.Next() {
		var item objeto
		if erro := tab.Scan(
			&item.Codigo,
			&item.Nome,
			&item.Celular,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}
