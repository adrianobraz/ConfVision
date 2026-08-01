package faturaV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func itemInsere(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var item FaturaItem

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &item); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := item.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, item.ID_FaturaItem)
}

func itemGetDadosById(w http.ResponseWriter, r *http.Request) {

	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var item FaturaItem

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &item); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := item.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, item)
}

func itemAlteraById(w http.ResponseWriter, r *http.Request) {

	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var item FaturaItem

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &item); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := item.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, item)
}

func itemDeleteById(w http.ResponseWriter, r *http.Request) {

	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var item FaturaItem

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &item); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := item.DeleteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func itemDeleteAllByIdFatura(w http.ResponseWriter, r *http.Request) {

	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var item FaturaItem

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &item); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := item.DeleteAllByIdFatura(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func itemListaByIdFatura(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var fat FaturaItem

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &fat); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []FaturaItem
	// Executar a operação
	if err := fat.ListaByIdFatura(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}
