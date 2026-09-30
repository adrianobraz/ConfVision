package auth

import (
	"apifunction/db"
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

// SessaoAdm sessão admConfmonit com escopo CEN/REP.
type SessaoAdm struct {
	UsuarioAdm
	UserTipo          string
	IDRepresentante   string
	IDCentralCatalogo string
	IDCentralUUID     string
}

func RequireEscopo(r *http.Request) (SessaoAdm, error) {
	if strings.ToUpper(strings.TrimSpace(r.Header.Get("X-Breakglass"))) == "S" {
		idCentral := strings.TrimSpace(r.Header.Get("X-Adm-Central"))
		if idCentral == "" {
			return SessaoAdm{}, errors.New("break-glass exige X-Adm-Central")
		}
		if strings.TrimSpace(r.Header.Get("X-Adm-Token")) == "" {
			return SessaoAdm{}, errors.New("break-glass exige X-Adm-Token")
		}
		cat := resolverIDCentralCatalogo(idCentral, "CENTRAL")
		return SessaoAdm{
			UsuarioAdm: UsuarioAdm{
				IDUsuario:     "BREAKGLASS",
				IDVinculo:     cat,
				Master:        "S",
				AdmFinanceiro: "S",
			},
			UserTipo:          "CEN",
			IDCentralUUID:     idCentral,
			IDCentralCatalogo: cat,
		}, nil
	}

	admToken := strings.TrimSpace(r.Header.Get("X-Adm-Token"))
	idCentral := strings.TrimSpace(r.Header.Get("X-Adm-Central"))
	if admToken != "" && idCentral != "" {
		userTipo := strings.ToUpper(strings.TrimSpace(r.Header.Get("X-Adm-Tipo")))
		if userTipo != "CEN" && userTipo != "REP" {
			userTipo = "CEN"
		}
		s := SessaoAdm{
			UsuarioAdm: UsuarioAdm{
				IDUsuario:     "ADMCONF",
				IDVinculo:     idCentral,
				Master:        "S",
				AdmFinanceiro: "S",
			},
			UserTipo: userTipo,
		}
		s.IDCentralCatalogo = resolverIDCentralCatalogo(idCentral, idCentral)
		if s.IDCentralCatalogo == "" {
			s.IDCentralCatalogo = idCentral
		}
		if isCentralUUID(idCentral) {
			s.IDCentralUUID = idCentral
		} else if uuid := resolverIDCentralUUID(idCentral); uuid != "" {
			s.IDCentralUUID = uuid
		} else {
			s.IDCentralUUID = idCentral
		}
		if userTipo == "REP" {
			s.IDRepresentante = strings.TrimSpace(r.Header.Get("X-Adm-Representante"))
			if s.IDRepresentante == "" {
				return SessaoAdm{}, errors.New("REP exige X-Adm-Representante")
			}
			s.IDVinculo = s.IDRepresentante
		} else {
			s.IDVinculo = s.IDCentralCatalogo
		}
		return s, nil
	}

	u, err := RequireAdm(r)
	if err != nil {
		return SessaoAdm{}, err
	}
	s := SessaoAdm{UsuarioAdm: u}

	var tipoRep sql.NullString
	var idCentralUUID sql.NullString
	var repCentralUUID sql.NullString
	err = db.Conn.QueryRow(`
		SELECT
			rep.ID_Representante,
			COALESCE(NULLIF(usuarios.IDCentralUUID, ''), ''),
			COALESCE(NULLIF(rep.IDCentralUUID, ''), '')
		FROM usuarios
		LEFT JOIN representante AS rep ON usuarios.ID_Vinculo = rep.ID_Representante
		WHERE usuarios.ID_Usuario = ?
		LIMIT 1
	`, u.IDUsuario).Scan(&tipoRep, &idCentralUUID, &repCentralUUID)
	if err != nil {
		return SessaoAdm{}, err
	}

	s.IDCentralUUID = strings.TrimSpace(idCentralUUID.String)

	if tipoRep.Valid && strings.TrimSpace(tipoRep.String) != "" {
		s.UserTipo = "REP"
		s.IDRepresentante = strings.TrimSpace(tipoRep.String)
		uuidCentral := strings.TrimSpace(repCentralUUID.String)
		if uuidCentral == "" {
			uuidCentral = s.IDCentralUUID
		}
		s.IDCentralCatalogo = resolverIDCentralCatalogo(uuidCentral, "")
	} else {
		s.UserTipo = "CEN"
		s.IDCentralCatalogo = resolverIDCentralCatalogo(s.IDCentralUUID, u.IDVinculo)
		if s.IDCentralUUID == "" && isCentralUUID(s.IDCentralCatalogo) {
			s.IDCentralUUID = s.IDCentralCatalogo
		}
	}

	if s.IDCentralCatalogo == "" && s.IDCentralUUID == "" {
		return SessaoAdm{}, errors.New("sessao sem idCentral")
	}
	if s.IDCentralCatalogo == "" {
		s.IDCentralCatalogo = s.IDCentralUUID
	}
	if !isCentralUUID(s.IDCentralUUID) {
		if uuid := resolverIDCentralUUID(s.IDCentralCatalogo); uuid != "" {
			s.IDCentralUUID = uuid
		} else if uuid := resolverIDCentralUUID(s.IDCentralUUID); uuid != "" {
			s.IDCentralUUID = uuid
		} else if s.UserTipo == "REP" && s.IDRepresentante != "" {
			if uuid := resolverIDCentralUUIDPorRepresentante(s.IDRepresentante); uuid != "" {
				s.IDCentralUUID = uuid
			}
		}
	}
	return s, nil
}

// resolverIDCentralUUID converte ID_Central legado (ex. "CENTRAL") em IDCentralUUID via MySQL.
func resolverIDCentralUUID(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	if isCentralUUID(ref) {
		return ref
	}
	if db.Conn == nil {
		return ""
	}
	var uuid sql.NullString
	err := db.Conn.QueryRow(`
		SELECT COALESCE(NULLIF(IDCentralUUID, ''), '')
		FROM central
		WHERE ID_Central = ? OR IDCentralUUID = ?
		LIMIT 1
	`, ref, ref).Scan(&uuid)
	if err != nil {
		return ""
	}
	out := strings.TrimSpace(uuid.String)
	if isCentralUUID(out) {
		return out
	}
	return ""
}

func resolverIDCentralUUIDPorRepresentante(idRepresentante string) string {
	idRepresentante = strings.TrimSpace(idRepresentante)
	if idRepresentante == "" || db.Conn == nil {
		return ""
	}
	// representante.IDCentralUUID guarda ID_Central (catalogo), nao o hex UUID.
	var ref sql.NullString
	err := db.Conn.QueryRow(`
		SELECT COALESCE(NULLIF(IDCentralUUID, ''), '')
		FROM representante
		WHERE ID_Representante = ?
		LIMIT 1
	`, idRepresentante).Scan(&ref)
	if err != nil {
		return ""
	}
	return resolverIDCentralUUID(ref.String)
}

func resolverIDCentralCatalogo(uuid, idVinculo string) string {
	uuid = strings.TrimSpace(uuid)
	idVinculo = strings.TrimSpace(idVinculo)
	if uuid != "" {
		if db.Conn == nil {
			return uuid
		}
		var idCentral sql.NullString
		err := db.Conn.QueryRow(`
			SELECT ID_Central FROM central
			WHERE IDCentralUUID = ? OR ID_Central = ?
			LIMIT 1
		`, uuid, uuid).Scan(&idCentral)
		if err == nil {
			if out := strings.TrimSpace(idCentral.String); out != "" {
				return out
			}
		}
		return uuid
	}
	if idVinculo != "" {
		return idVinculo
	}
	return ""
}

func isCentralUUID(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && strings.ToUpper(s) != "CENTRAL" && len(s) >= 20
}

// IDCentralFinanceiro retorna UUID para gravacao/consulta em fp_* (MySQL/Postgres).
func (s SessaoAdm) IDCentralFinanceiro() string {
	if isCentralUUID(s.IDCentralUUID) {
		return strings.TrimSpace(s.IDCentralUUID)
	}
	if isCentralUUID(s.IDCentralCatalogo) {
		return strings.TrimSpace(s.IDCentralCatalogo)
	}
	if uuid := resolverIDCentralUUID(s.IDCentralCatalogo); uuid != "" {
		return uuid
	}
	if uuid := resolverIDCentralUUID(s.IDCentralUUID); uuid != "" {
		return uuid
	}
	if s.UserTipo == "REP" && s.IDRepresentante != "" {
		if uuid := resolverIDCentralUUIDPorRepresentante(s.IDRepresentante); uuid != "" {
			return uuid
		}
	}
	return ""
}

func (s SessaoAdm) RequireIDCentralFinanceiro() (string, error) {
	if id := s.IDCentralFinanceiro(); id != "" {
		return id, nil
	}
	return "", errors.New("id_central obrigatorio (IDCentralUUID)")
}
