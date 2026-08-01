package contactidV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type ContactId struct {
	ID_Contactid string `json:"idContactid"`
	Codigo       string `json:"codigo"`
	Grupo        string `json:"grupo"`
	Descricao    string `json:"descricao"`
	Nivel        string `json:"nivel"`
	ID_Viculo    string `json:"idVinculo"`
}

type SContactId struct {
	ID_Contactid sql.NullString
	Codigo       sql.NullString
	Grupo        sql.NullString
	Descricao    sql.NullString
	Nivel        sql.NullString
	ID_Viculo    sql.NullString
}

type ListaGrupos struct {
	Grupo string `json:"grupo"`
}

func (ci *ContactId) Insere() error {
	if ci.ID_Contactid == "" {
		ci.ID_Contactid = auxiliar.GeradorDeId()
	}

	if ci.Codigo == "" {
		return errors.New("um codigo de contactid deve ser informado")
	}

	if ci.Grupo == "" {
		return errors.New("um grupo de contactid deve ser informado")
	}

	if ci.Descricao == "" {
		return errors.New("uma descrição para contactid deve ser informado")
	}

	if ci.Nivel == "" {
		return errors.New("um nivel de contact id deve ser informado")
	}

	if ci.ID_Viculo == "" {
		ci.ID_Viculo = "1"
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO contactId(
			contactId.ID_ContactId, 
			contactId.Codigo, 
			contactId.Grupo, 
			contactId.Descricao, 
			contactId.Nivel,
			contactId.ID_Vinculo
		) VALUES ( ?, ?, ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(ci.ID_Contactid),
		strings.ToUpper(ci.Codigo),
		strings.ToUpper(ci.Grupo),
		strings.ToUpper(ci.Descricao),
		strings.ToUpper(ci.Nivel),
		strings.ToUpper(ci.ID_Viculo),
	); err != nil {
		return err
	}

	return nil
}

func (ci *ContactId) GetDadosByCodigo() error {

	if ci.Codigo == "" {
		return errors.New(`um código de contactid deve ser informado`)
	}

	if ci.ID_Viculo == "" {
		ci.ID_Viculo = "1"
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Pesquisa personalizado =================================================
	filtro := fmt.Sprintf(`
		WHERE contactId.Codigo = '%s'
		AND contactId.ID_Vinculo = '%s'
	`, ci.Codigo, ci.ID_Viculo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, ci); err != nil {
			return err
		}
		return nil

	}

	// Casao não tenha personalizado pesquisa geral ===========================
	filtro = fmt.Sprintf(`
		WHERE contactId.Codigo = '%s'
		AND contactId.ID_Vinculo = '1'
	`, ci.Codigo)

	tab, err = db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, ci); err != nil {
			return err
		}
		return nil

	}

	return msgNaoEncontrado()
}

func (ci *ContactId) GetDadosById() error {

	if ci.ID_Contactid == "" {
		return errors.New(`um id de contactid deve ser informado`)
	}

	if ci.ID_Viculo == "" {
		ci.ID_Viculo = "1"
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE contactId.ID_Contactid = '%s'
	`, ci.ID_Contactid)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, ci); err != nil {
			return err
		}
		return nil

	}

	return msgNaoEncontrado()
}

func (ci *ContactId) AlteraById() error {
	if ci.ID_Contactid == "" {
		return errors.New("um id de contactid deve ser informado")
	}

	if ci.Grupo == "" {
		return errors.New("um grupo de contactid deve ser informado")
	}

	if ci.Descricao == "" {
		return errors.New("uma descrição para contactid deve ser informado")
	}

	if ci.Nivel == "" {
		return errors.New("um nivel de contact id deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE contactId 
		SET 
			contactId.Grupo = ?,
			contactId.Descricao = ?,
			contactId.Nivel = ?
		WHERE contactId.ID_Contactid = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(ci.Grupo),
		strings.ToUpper(ci.Descricao),
		strings.ToUpper(ci.Nivel),
		ci.ID_Contactid,
	); err != nil {
		return err
	}

	return nil
}

func (ci *ContactId) DeletaById() error {
	if ci.ID_Contactid == "" {
		return errors.New("um id de contactid deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM contactId WHERE contactId.ID_Contactid = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(ci.ID_Contactid); err != nil {
		return err
	}

	return nil
}

func (ci *ContactId) DeletaAllByIdVinculo() error {
	if ci.ID_Viculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM contactId WHERE contactId.ID_Vinculo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(ci.ID_Viculo); err != nil {
		return err
	}

	return nil
}

func (ci *ContactId) ListaPadrao(lista *[]ContactId) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := `
		WHERE contactId.ID_Vinculo = 'CENTRAL'
		ORDER BY contactId.Codigo
	`
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item ContactId
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (ci *ContactId) ListaByIdVinculo(lista *[]ContactId) error {
	if ci.ID_Viculo == "" {
		return errors.New("um id vinculo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE contactId.ID_Vinculo = '%s'
		ORDER BY contactId.Codigo
	`, ci.ID_Viculo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item ContactId
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (ci *ContactId) ListaByGrupo(lista *[]ContactId) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE contactId.Grupo = '%s'`, ci.Grupo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item ContactId
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// Lista os grupos disponiveis
func (c *ContactId) ListaGrupos(lista *[]ListaGrupos) error {
	// Abre um canal de conexao
	db, erro := connV4.Conectar()
	if erro != nil {
		return erro
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT 
			contactId.Grupo
		FROM  contactId 
		GROUP BY contactId.Grupo
		ORDER BY contactId.Grupo
	`)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	for tab.Next() {

		var item ListaGrupos

		if erro := tab.Scan(&item.Grupo); erro != nil {
			fmt.Println(erro)
			return erro
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (ci *ContactId) GetNivelById() (int, error) {
	return 0, nil
}

// Funcoes internas ===========================================================
func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			contactId.ID_Contactid,
			contactId.Codigo, 
			contactId.Grupo, 
			contactId.Descricao, 
			contactId.Nivel, 
			contactId.ID_Vinculo
		FROM contactId
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, ci *ContactId) error {
	var tmp SContactId
	if err := tab.Scan(
		&tmp.ID_Contactid,
		&tmp.Codigo,
		&tmp.Grupo,
		&tmp.Descricao,
		&tmp.Nivel,
		&tmp.ID_Viculo,
	); err != nil {
		return err
	}

	*ci = sContactIdToContactId(tmp)
	return nil
}

func sContactIdToContactId(sci SContactId) (ci ContactId) {
	ci.Codigo = sci.Codigo.String
	ci.ID_Contactid = sci.ID_Contactid.String
	ci.Grupo = sci.Grupo.String
	ci.Descricao = sci.Descricao.String
	ci.Nivel = sci.Nivel.String
	ci.ID_Viculo = sci.ID_Viculo.String
	return
}

func msgNaoEncontrado() error {
	return errors.New("contactid não encontrado na base de dados")
}
