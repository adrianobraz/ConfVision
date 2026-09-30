// Aplica schema PostgreSQL fp_receptor_presenca / fp_receptor_diag_log.
// Uso: go run ./cmd/migrate-receptordiag (a partir da pasta apifunction, com .env carregado)
package main

import (
	"log"

	"apifunction/config"
	"apifunction/pgcredito"
	"apifunction/pgreceptordiag"
)

func main() {
	config.Carregar()
	if config.PostgresDSN == "" {
		log.Fatal("POSTGRES_URL vazio no .env")
	}
	if err := pgcredito.Abrir(config.PostgresDSN); err != nil {
		log.Fatalf("postgres: %v", err)
	}
	if err := pgreceptordiag.Migrate(); err != nil {
		log.Fatalf("migrate receptor diag: %v", err)
	}
	log.Println("ok: schema receptor diag aplicado")
}
