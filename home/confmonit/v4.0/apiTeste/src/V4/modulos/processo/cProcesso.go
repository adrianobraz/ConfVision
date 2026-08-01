package processoV4

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
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, proc.ID_Processo)
}

func getIdProcessoByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.GetIdProcessoByIdDispositivo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, proc.ID_Processo)
}

func getIdAtendentelById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.GetIdAtendentelById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, proc.ID_Atendente)
}

func setIdAtendenteById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

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

func getNivelById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.GetNivelById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, proc.Nivel)
}

func setNivelById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.SetNivelById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func atualizaNivelById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.AtualizaNivelById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, proc.Nivel)
}

func getMsgAtendenteById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.GetMsgAtendenteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, proc.MsgAtendente)
}

func setMsgAtendenteById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := proc.SetMsgAtendenteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func ListarToOpen(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Processo
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

func ListarToOpenByIdDispositivo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var proc Processo

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &proc); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Processo
	// Executa o comendo
	if err := proc.ListarToOpenByIdDispositivo(&lista); err != nil {
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

func ListarToOpenByIdCliente(w http.ResponseWriter, r *http.Request) {
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

	var proc Processo
	var lista []Processo
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

func listarByFiltro(w http.ResponseWriter, r *http.Request) {
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

	var proc Processo
	var lista []Processo
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

func listarEventosByFiltro(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var filtro struct {
		Filtro string `json:"filtro"`
	}

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &filtro); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var proc Processo
	var lista []ProcEvt
	// Executa o comendo
	if err := proc.ListarEventosByFiltro(filtro.Filtro, &lista); err != nil {
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
