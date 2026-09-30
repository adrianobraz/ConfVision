package pgcentralwhitelabel

import (
	_ "embed"
	"fmt"

	"apifunction/pgcredito"
)

//go:embed schema.sql
var schemaSQL string

func Migrate() error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("aplicar schema central whitelabel: %w", err)
	}
	return nil
}

func Configurado() bool {
	return pgcredito.Configurado()
}
