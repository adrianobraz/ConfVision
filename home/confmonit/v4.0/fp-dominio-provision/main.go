package main

import (
	"fmt"
	"log"
	"net/http"

	"fp-dominio-provision/config"
	"fp-dominio-provision/handler"
)

func main() {
	config.Carregar()
	h := handler.Novo()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/provisionar", h.Provisionar)
	mux.HandleFunc("/remover", h.Remover)
	mux.HandleFunc("/retentar-ssl", h.RetentarSSL)

	addr := fmt.Sprintf("127.0.0.1:%s", config.Porta)
	log.Printf("fp-dominio-provision escutando em %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
