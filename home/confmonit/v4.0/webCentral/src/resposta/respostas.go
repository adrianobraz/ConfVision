package resposta

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ErroApi struct {
	Status string `json:"status"`
}

// Erro retorna um erro em formato JSON
func Erro(w http.ResponseWriter, statusCode int, erro error) {
	JSON(w, statusCode, struct {
		Erro string `json:"status"`
	}{
		fmt.Sprintf("Erro: %s", erro.Error()),
	})
}

func JsonOK(w http.ResponseWriter) {
	ok := struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	}

	JSON(w, http.StatusOK, ok)
}

func TratarStatusCodeDeErro(w http.ResponseWriter, r *http.Response) {
	var erro ErroApi
	corpo, _ := io.ReadAll(r.Body)
	json.Unmarshal(corpo, &erro)
	JSON(w, r.StatusCode, erro)
}

func App(w http.ResponseWriter, corpo []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(corpo)
}

// JSON retorna uma resposta em JSON para a requisição
func JSON(w http.ResponseWriter, statusCode int, dados interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if dados != nil {
		if erro := json.NewEncoder(w).Encode(dados); erro != nil {
			log.Fatal(erro)
		}
	}

}
