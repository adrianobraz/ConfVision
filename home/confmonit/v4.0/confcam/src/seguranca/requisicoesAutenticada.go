package seguranca

import (
	"io"
	"net/http"
)

// RequisiacaoAutenticada efetua uma requisçao com atutenticação pelo token
// e retorna a resposta para quem a chamou
func RequisiacaoAutenticada(r *http.Request, metodo, url string, dados io.Reader) (*http.Response, error) {
	request, erro := http.NewRequest(metodo, url, dados)
	if erro != nil {
		return nil, erro
	}

	token := TokenDaRequisicao(r)
	request.Header.Add("Authorization", "Bearer "+token)
	if dados != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	cliente := &http.Client{}
	return cliente.Do(request)
}
