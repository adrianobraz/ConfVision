package pgcatalogo

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
		return fmt.Errorf("aplicar schema catalogo: %w", err)
	}
	if _, err := db.Exec(`
		ALTER TABLE fp_pacote_cota
		ADD COLUMN IF NOT EXISTS valor_piso_breakglass NUMERIC(12, 2) NOT NULL DEFAULT 0;
		UPDATE fp_pacote_cota
		SET valor_piso_breakglass = valor
		WHERE valor_piso_breakglass = 0 AND valor > 0;
	`); err != nil {
		return fmt.Errorf("migrate pacote piso: %w", err)
	}
	if _, err := db.Exec(`
		ALTER TABLE vis_capacidade_config
		ADD COLUMN IF NOT EXISTS preco_piso_breakglass NUMERIC(12, 2) NOT NULL DEFAULT 0;
		UPDATE vis_capacidade_config
		SET preco_piso_breakglass = preco_base_camera
		WHERE preco_piso_breakglass = 0 AND preco_base_camera > 0;
	`); err != nil {
		return fmt.Errorf("migrate confvision capacidade piso: %w", err)
	}
	if _, err := db.Exec(`
		ALTER TABLE fp_central_preco_config
		ADD COLUMN IF NOT EXISTS valor_unitario_minimo_cota NUMERIC(12, 2) NOT NULL DEFAULT 0;
	`); err != nil {
		return fmt.Errorf("migrate central unitario minimo cota: %w", err)
	}
	return nil
}

func Configurado() bool {
	return pgcredito.Configurado()
}
