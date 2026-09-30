package main

import (
	"log"
	"os"

	"apifunction/config"
	"apifunction/pgcentraldominio"
	"apifunction/pgcentralwhitelabel"
	"apifunction/pgcredito"
)

func main() {
	config.Carregar()
	dsn := config.PostgresDSN
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_URL")
	}
	if dsn == "" {
		log.Fatal("POSTGRES_URL obrigatorio")
	}
	if err := pgcredito.Abrir(dsn); err != nil {
		log.Fatalf("conectar postgres: %v", err)
	}
	if err := pgcentraldominio.Migrate(); err != nil {
		log.Fatalf("migration fp_central_dominio_marca: %v", err)
	}
	if err := pgcentralwhitelabel.Migrate(); err != nil {
		log.Fatalf("migration fp_central_whitelabel: %v", err)
	}
	log.Println("migration fp_central_dominio_marca + fp_central_whitelabel ok")
}
