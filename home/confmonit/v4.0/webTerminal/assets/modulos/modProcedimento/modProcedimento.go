package modProcedimento

import (
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
		Uri:    "/modProcedimento/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

type objeto struct {
	IdFranqueado string `json:"idFranqueado"`
	IdCliente    string `json:"idCliente"`
	Grupo        string `json:"grupo"`
	Procedimento string `json:"procedimento"`
}

func buscarDados(w http.ResponseWriter, r *http.Request) {

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

	/*
	   ALARME
	   ARME
	   DESARME
	   EMERGENCIA
	   FALHAS
	   GERAL
	   MEDICO
	   PANICO
	   RESTAURE
	   SETUP
	   TESTE
	*/
	var item objeto
	var lista []objeto

	item.IdFranqueado = obj.IdFranqueado
	item.IdCliente = obj.IdCliente

	///////////////////// ALARME /////////////////////
	obj.Grupo = "ALARME"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "ALARME"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	/////////////////////// ARME /////////////////////
	obj.Grupo = "ARME"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "ARME"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	//////////////////// DESARME /////////////////////
	obj.Grupo = "DESARME"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "DESARME"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	/////////////////// EMERGENCIA ///////////////////
	obj.Grupo = "EMERGENCIA"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "EMERGENCIA"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	///////////////////// FALHAS /////////////////////
	obj.Grupo = "FALHAS"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "FALHAS"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	///////////////////// GERAL //////////////////////
	obj.Grupo = "GERAL"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "GERAL"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	///////////////////// MEDICO /////////////////////
	obj.Grupo = "MEDICO"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "MEDICO"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	///////////////////// PANICO /////////////////////
	obj.Grupo = "PANICO"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "PANICO"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	/////////////////// RESTAURE /////////////////////
	obj.Grupo = "RESTAURE"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "RESTAURE"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	///////////////////// SETUP //////////////////////
	obj.Grupo = "SETUP"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "SETUP"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	////////////////////// TESTE /////////////////////
	obj.Grupo = "TESTE"
	if erro := obj.bucarProcedimento(); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	item.Grupo = "TESTE"
	item.Procedimento = obj.Procedimento
	lista = append(lista, item)

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}

func (p *objeto) bucarProcedimento() error {

	// Valida a existencia do id de franqueado
	if p.IdFranqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	// Valida a existencia do id do cliente
	if p.IdCliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	// Abre um canal de conexao
	db, erro := aux.Conectar()
	if erro != nil {
		return erro
	}
	defer db.Close()

	////////////////////////////////////////////////////////////
	//        Pequisa procedimentos cliente especifico        //
	////////////////////////////////////////////////////////////
	tab, erro := db.Query(`
		SELECT procedimentos.Descricao
		FROM procedimentos
		WHERE procedimentos.ID_Franqueado = ?
		AND procedimentos.ID_Cliente = ?
		AND procedimentos.Grupo = ?
		AND procedimentos.Ativo = 'S'
	
	`, p.IdFranqueado, p.IdCliente, p.Grupo)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	if tab.Next() {
		if erro := tab.Scan(&p.Procedimento); erro != nil {
			return erro
		}
		return nil
	}

	////////////////////////////////////////////////////////////
	//        Pequisa procedimentos geral do franqueado       //
	////////////////////////////////////////////////////////////
	tab, erro = db.Query(`
		SELECT procedimentos.Descricao
		FROM procedimentos
		WHERE procedimentos.ID_Franqueado = ?
		AND procedimentos.ID_Cliente = "TODOS"
		AND procedimentos.Grupo = ?
		AND procedimentos.Ativo = 'S'
	`, p.IdFranqueado, p.Grupo)
	if erro != nil {
		return erro
	}

	if tab.Next() {
		if erro := tab.Scan(&p.Procedimento); erro != nil {
			return erro
		}
		return nil
	}

	////////////////////////////////////////////////////////////
	//              Pequisa procedimentos padrao              //
	////////////////////////////////////////////////////////////
	txtSql := fmt.Sprintf(`
		SELECT procedimentos.Descricao
		FROM procedimentos
		WHERE procedimentos.ID_Procedimento = 'PADRAO_%s'
	`, p.Grupo)

	tab, erro = db.Query(txtSql)
	if erro != nil {
		return erro
	}

	if tab.Next() {
		if erro := tab.Scan(&p.Procedimento); erro != nil {
			return erro
		}
		return nil
	}
	p.Procedimento = "Sem procedimento para esse grupo de evento"
	return nil
}
