package auxiliar

import (
	"database/sql"
	"strings"
)

// ResolverIDCentralPorReferencia obtem central.ID_Central a partir de ID_Central
// ou central.IDCentralUUID (hex legado da sessao ou id gerado).
func ResolverIDCentralPorReferencia(db *sql.DB, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return IDCentralPadrao
	}

	var idCentral sql.NullString
	err := db.QueryRow(`
		SELECT ID_Central FROM central
		WHERE ID_Central = ? OR IDCentralUUID = ?
		LIMIT 1
	`, ref, ref).Scan(&idCentral)
	if err == nil {
		if s := strings.TrimSpace(idCentral.String); s != "" {
			return s
		}
	}

	return ref
}
