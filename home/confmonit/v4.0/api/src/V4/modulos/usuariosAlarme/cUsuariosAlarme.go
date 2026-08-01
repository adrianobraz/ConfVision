package usuariosAlarmeV4

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
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, ua.ID_Usuario)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, ua)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.AlteraById(); err != nil {
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
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func deletaAllByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.DeletaAllByIdDispositivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func listaByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []UsuarioAlarme
	// Executa o comendo
	if err := ua.ListaByIdDispositivo(&lista); err != nil {
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
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.GetAtivoByid(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, ua.Ativo)
}

func setAtivoByid(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.SetAtivoByid(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, ua.Ativo)
}

func inverteAtivoByid(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var ua UsuarioAlarme

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &ua); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := ua.InverteAtivoByid(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, ua.Ativo)
}
