package eventoV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func insere(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var evt Evento

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &evt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := evt.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.Dados(w, http.StatusOK, evt.ID_Evento)
}

func listaByIdProcesso(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var evt Evento

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &evt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Evento
	// Executar a operação
	if err := evt.ListarByIdProcesso(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listaAgrupadoByIdProcesso(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var evt Evento

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &evt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Evento
	// Executar a operação
	if err := evt.ListarAgrupadoByIdProcesso(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByDispStartEnd(w http.ResponseWriter, r *http.Request) {

	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var evt EvtFiltro

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &evt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []EvtFiltro
	// Executar a operação
	if err := evt.ListarByDispStartEnd(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByDispStartEndGrupo(w http.ResponseWriter, r *http.Request) {

	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var evt EvtFiltro

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &evt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []EvtFiltro
	// Executar a operação
	if err := evt.ListarByDispStartEndGrupo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByDispProcOpen(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var disp struct {
		IdDisp string `json:"idDisp"`
	}

	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &disp); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var evt Evento
	var lista []Evento
	// Executar a operação
	if err := evt.ListarByDispProcOpen(disp.IdDisp, &lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}
