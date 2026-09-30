package pgatendimento

import (
	_ "embed"
	"fmt"

	"apifunction/pgcredito"
)

//go:embed schema_parceiro.sql
var schemaParceiroSQL string

func Migrate() error {
	if !Configurado() {
		return nil
	}
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	if _, err := db.Exec(schemaParceiroSQL); err != nil {
		return fmt.Errorf("aplicar schema parceiro atendimento: %w", err)
	}
	return nil
}
