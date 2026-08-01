package respApp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// JSON retorna uma resposta em JSON para a requisição
func JSON(w http.ResponseWriter, statusCode int, dados interface{}) {
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if dados != nil {
		if erro := json.NewEncoder(w).Encode(dados); erro != nil {
			log.Fatal(erro)
		}
	}

}

func Dados(w http.ResponseWriter, statusCode int, dados interface{}) {
	JSON(w, statusCode, struct {
		Status string      `json:"status"`
		Dados  interface{} `json:"dados"`
	}{
		Status: "OK",
		Dados:  dados,
	})
}

// Erro retorna um erro em formato JSON
func Erro(w http.ResponseWriter, statusCode int, erro error) {
	JSON(w, statusCode, struct {
		Erro string `json:"status"`
	}{
		fmt.Sprintf("Erro: %s", erro.Error()),
	})
}

func OK(w http.ResponseWriter) {
	JSON(w, http.StatusOK, struct {
		Status string `json:"status,omitempty"`
	}{
		Status: "OK",
	})
}

func Vazio(w http.ResponseWriter) {
	JSON(w, http.StatusOK, struct {
		Status string `json:"status,omitempty"`
	}{
		Status: "Vazio",
	})
}

func Email(w http.ResponseWriter, resposta string) {
	JSON(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{
		Status: resposta,
	})
}
