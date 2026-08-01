package admfinanceiroV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/V4/seguranca"
	"database/sql"
	"errors"
	"strings"
)

type admLogin struct {
	Token              string `json:"token"`
	IDUsuario          string `json:"idUsuario"`
	IDVinculo          string `json:"idVinculo"`
	IDCentralUUID      string `json:"idCentralUUID"`
	IDCentralCatalogo  string `json:"idCentralCatalogo"`
	UserTipo           string `json:"userTipo"`
	Nome               string `json:"nome"`
	Nick               string `json:"nick"`
	Email1             string `json:"email1"`
	Master             string `json:"master"`
	AdmFinanceiro      string `json:"admFinanceiro"`
	Ativo              string `json:"ativo"`
	Senha              string `json:"senha,omitempty"`
}

func (a *admLogin) logar(email, senha string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	senha = strings.TrimSpace(senha)

	if email == "" {
		return errors.New("um email deve ser informado")
	}
	if senha == "" {
		return errors.New("uma senha deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
			usuarios.ID_Usuario,
			usuarios.ID_Vinculo,
			COALESCE(NULLIF(usuarios.IDCentralUUID, ''), '') AS IDCentralUUID,
			COALESCE(NULLIF(rep.IDCentralUUID, ''), '') AS RepIDCentralUUID,
			usuarios.Nome,
			usuarios.Nick,
			usuarios.Email1,
			usuarios.Senha,
			usuarios.Master,
			COALESCE(usuarios.AdmFinanceiro, 'N') AS AdmFinanceiro,
			rep.ID_Representante AS tipoRep,
			blUser.ID_Alvo AS blocUser,
			blRepId.ID_Alvo AS blocRepId
		FROM usuarios
		LEFT JOIN representante AS rep
			ON usuarios.ID_Vinculo = rep.ID_Representante
		LEFT JOIN listaBloqueio AS blUser
			ON usuarios.ID_Usuario = blUser.ID_Alvo
		LEFT JOIN listaBloqueio AS blRepId
			ON rep.ID_Representante = blRepId.ID_Alvo
		WHERE usuarios.Email1 = ?
		LIMIT 1
	`, email)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "admfinanceiro") {
			return errors.New("coluna AdmFinanceiro ausente — rode o SQL 20260718_usuarios_adm_financeiro.sql")
		}
		return err
	}
	defer tab.Close()

	if !tab.Next() {
		return errors.New("usuario ou senha invalidos")
	}

	var (
		master            sql.NullString
		admFin            sql.NullString
		tipoRep           sql.NullString
		blocUser          sql.NullString
		blocRepId         sql.NullString
		idCentralUUID     sql.NullString
		repIDCentralUUID  sql.NullString
		senhaHash         string
		idVinculoRaw      string
	)

	if err := tab.Scan(
		&a.IDUsuario,
		&idVinculoRaw,
		&idCentralUUID,
		&repIDCentralUUID,
		&a.Nome,
		&a.Nick,
		&a.Email1,
		&senhaHash,
		&master,
		&admFin,
		&tipoRep,
		&blocUser,
		&blocRepId,
	); err != nil {
		return err
	}

	a.IDVinculo = strings.TrimSpace(idVinculoRaw)
	a.Master = strings.ToUpper(strings.TrimSpace(master.String))
	if a.Master == "" {
		a.Master = "N"
	}
	a.AdmFinanceiro = strings.ToUpper(strings.TrimSpace(admFin.String))
	if a.AdmFinanceiro == "" {
		a.AdmFinanceiro = "N"
	}

	if err := seguranca.VerificarSenha(senhaHash, senha); err != nil {
		return errors.New("usuario ou senha invalidos")
	}

	if blocUser.Valid || blocRepId.Valid {
		a.Ativo = "N"
		return errors.New("usuario bloqueado")
	}
	a.Ativo = "S"

	a.IDCentralUUID = strings.TrimSpace(idCentralUUID.String)

	if a.IDVinculo == "CENTRAL" {
		a.UserTipo = "CEN"
		// Masters de qualquer Central usam ID_Vinculo="CENTRAL"; o escopo do catalogo
		// é o ID_Central real (matriz="CENTRAL"; novas centrais=mesmo id do UUID).
		idCatalogo := resolverIDCentralCatalogo(db, a.IDCentralUUID)
		if idCatalogo == "" {
			return errors.New("usuario Central sem IDCentralUUID — nao foi possivel determinar a Central")
		}
		a.IDVinculo = idCatalogo
		a.IDCentralCatalogo = idCatalogo
	} else if tipoRep.Valid {
		a.UserTipo = "REP"
		// idVinculo permanece = ID_Representante; catalogo usa a Central do representante.
		uuidCentral := strings.TrimSpace(repIDCentralUUID.String)
		if uuidCentral == "" {
			uuidCentral = a.IDCentralUUID
		}
		if uuidCentral == "" {
			return errors.New("representante sem Central vinculada (IDCentralUUID)")
		}
		idCatalogo := resolverIDCentralCatalogo(db, uuidCentral)
		if idCatalogo == "" {
			return errors.New("representante sem Central vinculada (IDCentralUUID)")
		}
		a.IDCentralUUID = uuidCentral
		a.IDCentralCatalogo = idCatalogo
	} else {
		return errors.New("acesso admConfmonit apenas para Central ou Representante")
	}

	if a.Master != "S" {
		return errors.New("acesso admConfmonit exige usuario Master")
	}
	if a.AdmFinanceiro != "S" {
		return errors.New("acesso admConfmonit exige flag AdmFinanceiro = S")
	}

	a.Token, err = seguranca.CriarToken(a.IDUsuario)
	if err != nil {
		return err
	}

	a.Senha = ""
	return nil
}

// resolverIDCentralCatalogo devolve o ID_Central usado no admConfmonit/Xano.
// Preferência: linha em `central` por UUID ou ID; fallback UUID.
// Nao inventa "CENTRAL" quando o UUID esta vazio — isso misturava Centrais distintas.
func resolverIDCentralCatalogo(db *sql.DB, idCentralUUID string) string {
	idCentralUUID = strings.TrimSpace(idCentralUUID)
	if idCentralUUID == "" {
		return ""
	}

	var idCentral sql.NullString
	err := db.QueryRow(`
		SELECT ID_Central
		FROM central
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

func setAdmFinanceiro(idUsuario, flag string) error {
	idUsuario = strings.TrimSpace(idUsuario)
	flag = strings.ToUpper(strings.TrimSpace(flag))
	if idUsuario == "" {
		return errors.New("idUsuario obrigatorio")
	}
	if flag != "S" && flag != "N" {
		return errors.New("admFinanceiro deve ser S ou N")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`UPDATE usuarios SET AdmFinanceiro = ? WHERE ID_Usuario = ?`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "admfinanceiro") {
			return errors.New("coluna AdmFinanceiro ausente — rode o SQL 20260718_usuarios_adm_financeiro.sql")
		}
		return err
	}
	defer stm.Close()

	res, err := stm.Exec(flag, idUsuario)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("usuario nao encontrado")
	}
	return nil
}
