package visdata

import (
	"database/sql"
	"testing"
	"time"
)

func TestLicencaUsoValido(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)

	if !LicencaUsoValido(sql.NullTime{}, now) {
		t.Fatal("sem valido_ate deve ser valida")
	}
	futuro := sql.NullTime{Time: now.Add(24 * time.Hour), Valid: true}
	if !LicencaUsoValido(futuro, now) {
		t.Fatal("valido_ate futuro deve ser valida")
	}
	passado := sql.NullTime{Time: now.Add(-time.Minute), Valid: true}
	if LicencaUsoValido(passado, now) {
		t.Fatal("valido_ate passado deve ser invalida")
	}
}
