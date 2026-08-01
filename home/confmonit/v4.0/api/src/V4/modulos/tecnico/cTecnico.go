package tecnicoV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func logar(w http.ResponseWriter, r *http.Request) {

	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber e manipular as informações do login
	var obj struct {
		Email string `json:"email"`
		Senha string `json:"senha"`
	}

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var login Tecnico

	// Efetua o login
	if err := login.logar(obj.Email, obj.Senha); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	login.Senha = ""
	// Envia as credenciais para o app solicitante
	respApp.Dados(w, http.StatusOK, login)
}

func insere(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, tec.ID_Tecnico)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, tec)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func deletaById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func deletaAllByIdVinculo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.DeletaAllByIdVinculo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func listaByIdVinculo(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Tecnico

	// Executar a operação
	if err := tec.ListaByIdVinculo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func resetSenhaById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.ResetSenhaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func alteraSenhaById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.AlteraSenhaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}

func getAtivoById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.GetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, tec.Ativo)
}

func setAtivoById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.SetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, tec.Ativo)
}

func inverteAtivoById(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var tec Tecnico

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &tec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := tec.InverteAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, tec.Ativo)
}
