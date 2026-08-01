package confserviceparceiroV4

import (
	connV4 "api/src/V4/conexao"
	"database/sql"
	"errors"
	"strings"
)

type admEscopo struct {
	UserTipo          string
	IDUsuario         string
	IDVinculo         string
	IDCentralCatalogo string
}

func carregarEscopoAdm(idUsuario string) (*admEscopo, error) {
	idUsuario = strings.TrimSpace(idUsuario)
	if idUsuario == "" {
		return nil, errors.New("usuario nao autenticado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var (
		idVinculoRaw     string
		idCentralUUID    sql.NullString
		repIDCentralUUID sql.NullString
		master           sql.NullString
		admFin           sql.NullString
		tipoRep          sql.NullString
	)
	err = db.QueryRow(`
		SELECT
			usuarios.ID_Vinculo,
			COALESCE(NULLIF(usuarios.IDCentralUUID, ''), '') AS IDCentralUUID,
			COALESCE(NULLIF(rep.IDCentralUUID, ''), '') AS RepIDCentralUUID,
			usuarios.Master,
			COALESCE(usuarios.AdmFinanceiro, 'N') AS AdmFinanceiro,
			rep.ID_Representante AS tipoRep
		FROM usuarios
		LEFT JOIN representante AS rep ON usuarios.ID_Vinculo = rep.ID_Representante
		WHERE usuarios.ID_Usuario = ?
		LIMIT 1
	`, idUsuario).Scan(&idVinculoRaw, &idCentralUUID, &repIDCentralUUID, &master, &admFin, &tipoRep)
	if err == sql.ErrNoRows {
		return nil, errors.New("usuario nao encontrado")
	}
	if err != nil {
		return nil, err
	}

	if strings.ToUpper(strings.TrimSpace(master.String)) != "S" {
		return nil, errors.New("acesso exige usuario Master")
	}
	if strings.ToUpper(strings.TrimSpace(admFin.String)) != "S" {
		return nil, errors.New("acesso exige AdmFinanceiro = S")
	}

	esc := &admEscopo{IDUsuario: idUsuario, IDVinculo: strings.TrimSpace(idVinculoRaw)}

	if esc.IDVinculo == "CENTRAL" {
		esc.UserTipo = "CEN"
		idCatalogo := resolverIDCentralCatalogo(db, strings.TrimSpace(idCentralUUID.String))
		if idCatalogo == "" {
			return nil, errors.New("central sem IDCentralUUID")
		}
		esc.IDVinculo = idCatalogo
		esc.IDCentralCatalogo = idCatalogo
		return esc, nil
	}

	if tipoRep.Valid {
		esc.UserTipo = "REP"
		uuidCentral := strings.TrimSpace(repIDCentralUUID.String)
		if uuidCentral == "" {
			uuidCentral = strings.TrimSpace(idCentralUUID.String)
		}
		idCatalogo := resolverIDCentralCatalogo(db, uuidCentral)
		if idCatalogo == "" {
			return nil, errors.New("representante sem central vinculada")
		}
		esc.IDCentralCatalogo = idCatalogo
		return esc, nil
	}

	return nil, errors.New("acesso apenas para Central ou Representante")
}

func resolverIDCentralCatalogo(db *sql.DB, idCentralUUID string) string {
	idCentralUUID = strings.TrimSpace(idCentralUUID)
	if idCentralUUID == "" {
		return ""
	}
	var idCentral sql.NullString
	err := db.QueryRow(`
		SELECT ID_Central FROM central
		WHERE IDCentralUUID = ? OR ID_Central = ?
		LIMIT 1
	`, idCentralUUID, idCentralUUID).Scan(&idCentral)
	if err == nil {
		out := strings.TrimSpace(idCentral.String)
		if out != "" {
			return out
		}
	}
	return idCentralUUID
}

func repPertenceCentral(idCentral, idRepresentante string) (bool, error) {
	idCentral = strings.TrimSpace(idCentral)
	idRepresentante = strings.TrimSpace(idRepresentante)
	if idCentral == "" || idRepresentante == "" {
		return false, errors.New("idCentral e idRepresentante obrigatorios")
	}
	db, err := connV4.Conectar()
	if err != nil {
		return false, err
	}
	defer db.Close()
	var n int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM representante
		WHERE ID_Representante = ?
		  AND (IDCentralUUID = ? OR IDCentralUUID = (
		    SELECT COALESCE(NULLIF(TRIM(IDCentralUUID), ''), ID_Central)
		    FROM central WHERE ID_Central = ? OR IDCentralUUID = ? LIMIT 1
		  ))
	`, idRepresentante, idCentral, idCentral, idCentral).Scan(&n)
	return n > 0, err
}

func assertRepDaCentral(idCentral, idRepresentante string) error {
	ok, err := repPertenceCentral(idCentral, idRepresentante)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("representante nao pertence a central")
	}
	return nil
}

func assertFranqueadoDoRep(idRepresentante, idFranqueado string) error {
	ok, err := franqueadoPertenceRep(idRepresentante, idFranqueado)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("franqueado nao pertence ao representante")
	}
	return nil
}

func franqueadoPertenceRep(idRepresentante, idFranqueado string) (bool, error) {
	idRepresentante = strings.TrimSpace(idRepresentante)
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idRepresentante == "" || idFranqueado == "" {
		return false, errors.New("idRepresentante e idFranqueado obrigatorios")
	}
	db, err := connV4.Conectar()
	if err != nil {
		return false, err
	}
	defer db.Close()
	var n int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM franqueado
		WHERE ID_Franqueado = ? AND ID_Representante = ?
	`, idFranqueado, idRepresentante).Scan(&n)
	return n > 0, err
}
