package respApp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// JSON retorna uma resposta em JSON para a requisição
func RespJSON(w http.ResponseWriter, statusCode int, dados interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if dados != nil {
		if erro := json.NewEncoder(w).Encode(dados); erro != nil {
			log.Fatal(erro)
		}
	}

}

func RespDados(w http.ResponseWriter, statusCode int, dados interface{}) {
	RespJSON(w, statusCode, struct {
		Status string      `json:"status"`
		Dados  interface{} `json:"dados"`
	}{
		Status: "OK",
		Dados:  dados,
	})
}

// Erro retorna um erro em formato JSON
func RespErro(w http.ResponseWriter, statusCode int, erro error) {
	RespJSON(w, statusCode, struct {
		Erro string `json:"status"`
	}{
		fmt.Sprintf("Erro: %s", erro.Error()),
	})
}

func RespOK(w http.ResponseWriter) {
	RespJSON(w, http.StatusOK, struct {
		Status string `json:"status,omitempty"`
	}{
		Status: "OK",
	})
}

func RespVazio(w http.ResponseWriter) {
	RespJSON(w, http.StatusOK, struct {
		Status string `json:"status,omitempty"`
	}{
		Status: "Vazio",
	})
}

func TratarStatusCodeDeErro(id string, r *http.Response) error {
	var erro struct {
		Status string `json:"status"`
	}
	corpo, _ := io.ReadAll(r.Body)
	json.Unmarshal(corpo, &erro)
	return errors.New(id + " Benuvem -> " + erro.Status)
}
