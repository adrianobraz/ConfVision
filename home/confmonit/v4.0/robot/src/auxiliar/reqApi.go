package auxiliar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"robot/src/config"
)

func Requisitar(url string, dados io.Reader) (*http.Response, error) {

	request, erro := http.NewRequest(http.MethodPost, url, dados)
	if erro != nil {
		return nil, erro
	}

	// Pega token de autorização
	token, erro := logarApi()
	if erro != nil {
		return nil, erro
	}

	request.Header.Add("Authorization", "Bearer "+token)

	cliente := &http.Client{}

	response, erro := cliente.Do(request)
	if erro != nil {
		fmt.Println(erro)
		return nil, erro
	}

	return response, nil
}

func logarApi() (string, error) {

	// Cria o json pra requeisição
	body, _ := json.Marshal(map[string]string{
		"key": string(config.SecretKey),
	})

	payload := bytes.NewBuffer(body)

	// Efetua a requisição
	resp, erro := http.Post(config.Url+"/ReceptorLogar", "application/json", payload)
	if erro != nil {
		return "", erro
	}

	body, erro = io.ReadAll(resp.Body)
	if erro != nil {
		return "", erro
	}
	defer resp.Body.Close()

	// Extrutura pra receber a respota da api
	var d struct {
		Status string `json:"status"`
		Token  string `json:"dados"`
	}

	if erro := json.Unmarshal(body, &d); erro != nil {
		return "", erro
	}

	if d.Status == "OK" {
		return d.Token, nil
	} else {
		return "", errors.New("erro ao logar")
	}

}

func Token() {
	token, erro := logarApi()
	if erro != nil {
		fmt.Println(erro)
	} else {
		fmt.Println(token)
	}

}
