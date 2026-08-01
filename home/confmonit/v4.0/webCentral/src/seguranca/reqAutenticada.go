package seguranca

import (
	"fmt"
	"io"
	"net/http"
)

// ReqAutenticada efetua uma requisçao com atutenticação pelo token
// e retorna a resposta para quem a chamou
func ReqAutenticada(r *http.Request, metodo, url string, dados io.Reader) (*http.Response, error) {
	request, erro := http.NewRequest(metodo, url, dados)
	if erro != nil {
		return nil, erro
	}

	cookie, _ := LerCookies(r)

	request.Header.Add("Authorization", "Bearer "+cookie["token"])

	cliente := &http.Client{}

	response, erro := cliente.Do(request)
	if erro != nil {
		fmt.Println(erro)
		return nil, erro
	}

	return response, nil
}
