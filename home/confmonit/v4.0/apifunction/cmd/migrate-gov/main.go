package main

import (
	"fmt"
	"log"
	"os"

	"apifunction/pgcredito"
	"apifunction/pggovernanca"
)

func main() {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		log.Fatal("POSTGRES_URL obrigatorio")
	}
	if err := pgcredito.Abrir(dsn); err != nil {
		log.Fatalf("conectar postgres: %v", err)
	}
	if len(os.Args) > 1 && os.Args[1] == "--verify" {
		verify()
		return
	}
	if err := pggovernanca.Migrate(); err != nil {
		log.Fatalf("migration governanca: %v", err)
	}
	log.Println("migration 004_fp_governanca_repasse ok")
	verify()
}

func verify() {
	db, err := pgcredito.DB()
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name LIKE 'fp_gov_%'
		ORDER BY 1
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("Tabelas fp_gov_*:")
	for rows.Next() {
		var name string
		_ = rows.Scan(&name)
		fmt.Println(" -", name)
	}
}
