package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"apifunction/config"
	"apifunction/db"
	"apifunction/handler"
	"apifunction/pgfinmirror"
	"apifunction/pgcatalogo"
	"apifunction/pgcobranca"
	"apifunction/pgcredito"
	"apifunction/pggovernanca"
	"apifunction/pgatendimento"
	"apifunction/pgreceptordns"
	"apifunction/pgreceptordiag"
	"apifunction/pgcentraldominio"
	"apifunction/pgcentralwhitelabel"
	"apifunction/service/tarifa"
)

func main() {
	config.Carregar()
	db.Abrir(config.MySQLDSN)
	if err := tarifa.Migrate(); err != nil {
		log.Fatalf("mysql tarifa migrate: %v", err)
	}
	if config.PostgresDSN != "" {
		if err := pgcredito.Abrir(config.PostgresDSN); err != nil {
			log.Fatalf("postgres credito: %v", err)
		}
		if err := pggovernanca.Migrate(); err != nil {
			log.Fatalf("postgres governanca: %v", err)
		}
		if err := pgfinmirror.Migrate(); err != nil {
			log.Fatalf("postgres fin mirror: %v", err)
		}
		if err := pgcobranca.Migrate(); err != nil {
			log.Fatalf("postgres cobranca: %v", err)
		}
		if err := pgcatalogo.Migrate(); err != nil {
			log.Fatalf("postgres catalogo: %v", err)
		}
		if err := pgatendimento.Migrate(); err != nil {
			log.Fatalf("postgres atendimento parceiro: %v", err)
		}
		if err := pgreceptordns.Migrate(); err != nil {
			log.Fatalf("postgres receptor dns: %v", err)
		}
		if err := pgreceptordiag.Migrate(); err != nil {
			log.Fatalf("postgres receptor diag: %v", err)
		}
		if err := pgcentraldominio.Migrate(); err != nil {
			log.Fatalf("postgres central dominio marca: %v", err)
		}
		if err := pgcentralwhitelabel.Migrate(); err != nil {
			log.Fatalf("postgres central whitelabel: %v", err)
		}
	} else {
		log.Println("aviso: POSTGRES_URL vazio — credito/debito usara fallback HTTP ops (legado)")
	}

	h := handler.Novo()
	addr := fmt.Sprintf(":%s", config.Porta)
	srv := &http.Server{
		Addr:              addr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("apifunction escutando em %s", addr)
	log.Fatal(srv.ListenAndServe())
}
