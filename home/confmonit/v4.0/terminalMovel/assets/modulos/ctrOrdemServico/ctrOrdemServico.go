package ctrOrdemServico

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/ctrOrdemServico/gravar",
		Metodo: http.MethodPost,
		Funcao: gravar,
	},
	{
		Uri:    "/ctrOrdemServico/buscarFranqueado",
		Metodo: http.MethodPost,
		Funcao: buscarFranqueado,
	},
	{
		Uri:    "/ctrOrdemServico/buscarCliente",
		Metodo: http.MethodPost,
		Funcao: buscarCliente,
	},
}

func gravar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdTicket  string `json:"idTicket"`
		IdMaster  string `json:"idMaster"`
		IdSlave   string `json:"idSlave"`
		Asunto    string `json:"assunto"`
		Descricao string `json:"descricao"`
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
		INSERT INTO ticket(
			ticket.ID_Ticket, 
			ticket.ID_Master, 
			ticket.ID_Slave, 
			ticket.Assunto, 
			ticket.Descricao, 
			ticket.Status

		) VALUES (?,?,?,?,?,?)
	`)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer stm.Close()

	obj.IdTicket = aux.GeradorDeId()
	fmt.Println(obj)
	if _, erro := stm.Exec(
		obj.IdTicket,
		obj.IdMaster,
		obj.IdSlave,
		strings.ToUpper(obj.Asunto),
		strings.ToUpper(obj.Descricao),
		"NOVO",
	); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	aux.RespostaJsonOK(w)
}

func buscarFranqueado(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdFranqueado string `json:"idFranqueado"`
		RazaoSocial  string `json:"razaoSocial"`
		NomeFantasia string `json:"nomeFantasia"`
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
			franqueado.RazaoSocial,
			franqueado.NomeFantasia
		FROM franqueado
	`)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto
	for tab.Next() {
		var item objeto

		if erro := tab.Scan(
			&item.IdFranqueado,
			&item.RazaoSocial,
			&item.NomeFantasia,
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

func buscarCliente(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdFranqueado string `json:"idFranqueado,omitempty"`
		IdCliente    string `json:"idCliente"`
		Nome         string `json:"nome"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	var cli objeto

	if err := json.Unmarshal(body, &cli); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
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
			cliente.ID_Cliente,
			cliente.Nome
		FROM cliente			
		WHERE cliente.ID_Franqueado = ?
	`, cli.IdFranqueado)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto
	for tab.Next() {
		var item objeto

		if erro := tab.Scan(
			&item.IdCliente,
			&item.Nome,
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
