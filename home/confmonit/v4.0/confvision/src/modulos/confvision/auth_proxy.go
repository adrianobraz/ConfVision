package confvision

import (
	"bytes"
	"confvision/src/seguranca"
	"io"
	"net/http"
)

func requisicaoAutenticada(r *http.Request, metodo, url string, dados io.Reader) (*http.Response, error) {
	request, erro := http.NewRequest(metodo, url, dados)
	if erro != nil {
		return nil, erro
	}

	token := seguranca.TokenDaRequisicao(r)
	request.Header.Add("Authorization", "Bearer "+token)
	if dados != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	cliente := &http.Client{}
	return cliente.Do(request)
}

func proxyConfmonitGet(w http.ResponseWriter, r *http.Request, url string) {
	resp, erro := requisicaoAutenticada(r, http.MethodGet, url, nil)
	if erro != nil {
		http.Error(w, erro.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	corpo, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(corpo)
}

func proxyConfmonitPostRaw(w http.ResponseWriter, r *http.Request, url string) {
	body, _ := io.ReadAll(r.Body)
	resp, erro := requisicaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		http.Error(w, erro.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	corpo, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(corpo)
}
