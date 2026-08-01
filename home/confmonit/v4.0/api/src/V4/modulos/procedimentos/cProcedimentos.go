package procedimentosV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func insere(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, pro.ID_Procedimento)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, pro)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func deletaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func deletaAllByIdFranqueado(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.DeletaAllByIdFranqueado(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func deletaAllByIdCliente(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.DeletaAllByIdCliente(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func listaByAllIdFranqueado(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos
	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Procedimentos

	// Executa o comendo
	if err := pro.ListaByAllIdFranqueado(&lista); err != nil {
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

func getAtivoByid(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.GetAtivoByid(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, pro.Ativo)
}

func setAtivoByid(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.SetAtivoByid(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, pro.Ativo)
}

func inverteAtivoByid(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var pro Procedimentos

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &pro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := pro.InverteAtivoByid(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, pro.Ativo)
}
