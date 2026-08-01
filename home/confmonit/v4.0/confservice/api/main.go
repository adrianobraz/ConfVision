package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"confservice/config"
	"confservice/db"
	"confservice/handler"
)

func main() {
	config.Carregar()
	db.Abrir(config.MySQLDSN)

	h := handler.Novo()
	addr := fmt.Sprintf(":%s", config.Porta)
	srv := &http.Server{
		Addr:              addr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Worker simples de retry a cada 60s
	go func() {
		t := time.NewTicker(60 * time.Second)
		defer t.Stop()
		for range t.C {
			h.ProcessarPendentes(20)
		}
	}()

	log.Printf("confservice api em %s", addr)
	log.Fatal(srv.ListenAndServe())
}
