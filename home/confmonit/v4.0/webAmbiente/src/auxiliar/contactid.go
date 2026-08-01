package auxiliar

import (
	"database/sql"
)

func GetCtiGrupoAndDescricao(vinculo, codigo string, grupo, descricao *string) error {
	db, err := Conectar()
	if err != nil {
		return err
	}
	defer db.Close()
	return GetCtiGrupoAndDescricaoDB(db, vinculo, codigo, grupo, descricao)
}

func GetCtiGrupoAndDescricaoDB(db *sql.DB, vinculo, codigo string, grupo, descricao *string) error {
	var sqlGrupo, sqlDescricao sql.NullString

	tabPer, err := db.Query(`
		SELECT contactId.Grupo, contactId.Descricao
		FROM contactId
		WHERE contactId.Codigo = ? AND contactId.ID_Vinculo = ?
	`, codigo, vinculo)
	if err != nil {
		return err
	}
	defer tabPer.Close()

	if tabPer.Next() {
		if err := tabPer.Scan(&sqlGrupo, &sqlDescricao); err != nil {
			return err
		}
		*grupo = sqlGrupo.String
		*descricao = sqlDescricao.String
		return nil
	}

	tabPad, err := db.Query(`
		SELECT contactId.Grupo, contactId.Descricao
		FROM contactId
		WHERE contactId.Codigo = ? AND contactId.ID_Vinculo = 'CENTRAL'
	`, codigo)
	if err != nil {
		return err
	}
	defer tabPad.Close()

	if tabPad.Next() {
		if err := tabPad.Scan(&sqlGrupo, &sqlDescricao); err != nil {
			return err
		}
		*grupo = sqlGrupo.String
		*descricao = sqlDescricao.String
		return nil
	}

	*grupo = "NÃO CADASTRADO"
	*descricao = "NÃO CADASTRADO"
	return nil
}
