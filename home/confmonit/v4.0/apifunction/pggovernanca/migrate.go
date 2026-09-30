package pggovernanca

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
		return fmt.Errorf("aplicar schema governanca: %w", err)
	}
	return nil
}
