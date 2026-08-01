package seguranca

import (
	"receptorWeb/src/config"
	"fmt"
	"io"
	"net/http"
)

// RequisiacaoAutenticada efetua uma requisçao com atutenticação pelo token
// e retorna a resposta para quem a chamou
func RequisiacaoAutenticada(metodo, url string, dados io.Reader) (*http.Response, error) {
	request, erro := http.NewRequest(metodo, url, dados)
	if erro != nil {
		return nil, erro
	}

	request.Header.Add("Authorization", "Bearer "+config.API.Token)

	cliente := &http.Client{}

	response, erro := cliente.Do(request)
	if erro != nil {
		fmt.Println(erro)
		return nil, erro
	}

	return response, nil
}
