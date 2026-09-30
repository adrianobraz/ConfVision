package tarifa

import (
	"apifunction/db"
	"strings"
)

func Migrate() error {
	if db.Conn == nil {
		return nil
	}
	stmts := []string{
		`ALTER TABLE fp_tarifa_operacional ADD COLUMN PisoTentativa DECIMAL(12,4) NOT NULL DEFAULT 0`,
		`ALTER TABLE fp_tarifa_operacional ADD COLUMN PisoMinuto DECIMAL(12,4) NOT NULL DEFAULT 0`,
		`ALTER TABLE fp_tarifa_operacional ADD COLUMN PisoUnidade DECIMAL(12,4) NOT NULL DEFAULT 0`,
	}
	for _, s := range stmts {
		if _, err := db.Conn.Exec(s); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
				return err
			}
		}
	}
	_, err := db.Conn.Exec(`
UPDATE fp_tarifa_operacional
SET PisoTentativa = ValorTentativa,
    PisoMinuto = ValorMinuto,
    PisoUnidade = ValorUnidade
WHERE PisoTentativa = 0 AND PisoMinuto = 0 AND PisoUnidade = 0
  AND (ValorTentativa > 0 OR ValorMinuto > 0 OR ValorUnidade > 0)`)
	return err
}

func SyncPisoBreakglass(idCentral string) error {
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" || db.Conn == nil {
		return nil
	}
	_, err := db.Conn.Exec(`
UPDATE fp_tarifa_operacional
SET PisoTentativa = ValorTentativa,
    PisoMinuto = ValorMinuto,
    PisoUnidade = ValorUnidade
WHERE ID_Central = ?`, idCentral)
	return err
}
