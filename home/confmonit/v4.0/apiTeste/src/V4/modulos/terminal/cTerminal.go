package terminalV4

import (
	connV4 "api/src/V4/conexao"
	processoV4 "api/src/V4/modulos/processo"
	"api/src/V4/respApp"
	"fmt"
	"strings"
	"time"

	"encoding/json"
	"io"
	"net/http"
)

/*
Parametro:

	ID_Processo string
	ID_Atendente string

Retorno: respApp.OK
*/
func bloqueiaProcessoById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc processoV4.Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.SetIdAtendenteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

/*
Parametro:

	ID_Processo string

Retorno: respApp.OK
*/
func liberaProcessoById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc processoV4.Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	proc.ID_Atendente = "0"

	// Executa o comendo
	if err := proc.SetIdAtendenteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

/*
Parametro:

	Deve passar um json vazio

Retorno: lista de processo
*/
func listarProcessos(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc processoV4.Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []processoV4.Processo

	// Executa o comendo
	if err := proc.ListarToOpen(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

/*
Parametro:

	ID_Cliente string

Retorno: lista de processo
*/
func listarProcessosByIdCliente(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var cli struct {
		ID_Cliente string `json:"idCliente"`
	}

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var proc processoV4.Processo
	var lista []processoV4.Processo
	// Executa o comendo
	if err := proc.ListarToOpenByIdCliente(cli.ID_Cliente, &lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

// ////////////////////////////////////

// Verificado OK
/*
Parametro:

	ID_Cliente string

Retorno: lista de processo abertos
*/
func listarProcessoToOpenByCliente(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var cli struct {
		ID_Cliente string `json:"idCliente"`
	}

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var proc processoV4.Processo
	var lista []processoV4.Processo
	// Executa o comendo
	if err := proc.ListarToOpenByCliente(cli.ID_Cliente, &lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func getDadosProcessoById(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var proc processoV4.Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, proc)

}

func listarEventosByProcesso(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var ter evtDetalhe

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ter); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []evtDetalhe
	// Executa o comendo
	if err := ter.listarEvtDetalheByProc(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}

}

//////////////////////////////////////

func listarEventosAgrupadosByProcesso(w http.ResponseWriter, r *http.Request) {}

func listarEventosDesagrupadoByProcesso(w http.ResponseWriter, r *http.Request) {}

func finalizarProcesso(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso string `json:"idProcesso"`
		IdCliente  string `json:"idCliente"`
		Descricao  string `json:"descricao"`
		Nome       string `json:"nome"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := connV4.Conectar()
	if erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	stm, erro := db.Prepare(`
		UPDATE processo
		SET 
			processo.Descricao = ?,
			processo.ID_Atendente = ?,
			processo.DataAtenFim = ?

		WHERE processo.ID_Processo = ?
	`)

	if erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		fmt.Sprintf("Finalizado por %s: %s", obj.Nome, strings.ToUpper(obj.Descricao)),
		obj.IdCliente,
		time.Now().Format("2006-01-02 15:04:05"),
		obj.IdProcesso,
	); erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	respApp.Dados(w, http.StatusOK, obj)
}
