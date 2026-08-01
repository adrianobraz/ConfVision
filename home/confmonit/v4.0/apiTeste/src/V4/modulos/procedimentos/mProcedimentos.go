package procedimentosV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Procedimentos struct {
	ID_Procedimento string `json:"idProcedimento"`
	ID_Franqueado   string `json:"idFranqueado"`
	FraNome         string `json:"fraNome"`
	ID_Cliente      string `json:"idCliente"`
	CliNome         string `json:"cliNome"`
	Grupo           string `json:"grupo"`
	Descricao       string `json:"descricao"`
	DataCadastro    string `json:"dataCadastro"`
	Ativo           string `json:"ativo"`
}

type SProcedimentos struct {
	ID_Procedimento sql.NullString
	ID_Franqueado   sql.NullString
	FraNome         sql.NullString
	ID_Cliente      sql.NullString
	CliNome         sql.NullString
	Grupo           sql.NullString
	Descricao       sql.NullString
	DataCadastro    sql.NullTime
	Ativo           sql.NullString

	// ID_Procedimento
	// ID_Franqueado
	// ID_Cliente
	// Grupo
	// Descricao
	// DataCadastro
	// Ativo
}

func (p *Procedimentos) Insere() error {
	if p.ID_Procedimento == "" {
		p.ID_Procedimento = auxiliar.GeradorDeId()
	}

	if p.ID_Franqueado == "" {
		errors.New("um id de franqueado deve ser informado")
	}

	if p.ID_Cliente == "" {
		errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO procedimentos(
			procedimentos.ID_Procedimento, 
			procedimentos.ID_Franqueado, 
			procedimentos.ID_Cliente, 
			procedimentos.Grupo, 
			procedimentos.Descricao
		) VALUES ( ?, ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(p.ID_Procedimento),
		strings.ToUpper(p.ID_Franqueado),
		strings.ToUpper(p.ID_Cliente),
		strings.ToUpper(p.Grupo),
		strings.ToUpper(p.Descricao),
	); err != nil {

		return err
	}

	return nil
}

func (p *Procedimentos) GetDadosById() error {
	if p.ID_Procedimento == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE procedimentos.ID_Procedimento = '%s'
	`, p.ID_Procedimento)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, p); err != nil {
			return err
		}
		return err
	}
	return errors.New("procedimento não encontrado na base de dados")
}

func (p *Procedimentos) AlteraById() error {
	if p.ID_Procedimento == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE procedimentos SET 
			procedimentos.Grupo = ?,
			procedimentos.Descricao = ?
		WHERE procedimentos.ID_Procedimento = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		p.Grupo,
		p.Descricao,
		p.ID_Procedimento,
	); err != nil {

		return err
	}
	return nil
}

func (p *Procedimentos) DeletaById() error {
	if p.ID_Procedimento == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM procedimentos WHERE procedimentos.ID_Procedimento = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.ID_Procedimento); err != nil {

		return err
	}
	return nil
}

func (p *Procedimentos) DeletaAllByIdFranqueado() error {
	if p.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM procedimentos WHERE procedimentos.ID_Franqueado = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.ID_Franqueado); err != nil {

		return err
	}
	return nil
}

func (p *Procedimentos) DeletaAllByIdCliente() error {
	if p.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM procedimentos WHERE procedimentos.ID_Cliente = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.ID_Cliente); err != nil {

		return err
	}
	return nil
}

func (p *Procedimentos) ListaByAllIdFranqueado(lista *[]Procedimentos) error {
	if p.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE procedimentos.ID_Franqueado = '%s'
	`, p.ID_Franqueado)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Procedimentos

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}

	return nil
}

// Manipula o campo ativo =====================================================
func (p *Procedimentos) GetAtivoByid() error {
	if p.ID_Procedimento == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(` 
		SELECT procedimentos.Ativo 
		FROM procedimentos 
		WHERE procedimentos.ID_Procedimento = ?
	`, p.ID_Procedimento)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&p.Ativo); err != nil {
			return err
		}
		return nil
	}

	return errors.New("procedimento não encontrado na base de dados")
}

func (p *Procedimentos) SetAtivoByid() error {
	if p.ID_Procedimento == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	if p.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE procedimentos 
		SET procedimentos.Ativo = ?
		WHERE procedimentos.ID_Procedimento = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		p.Ativo,
		p.ID_Procedimento,
	); err != nil {

		return err
	}
	return nil
}

func (p *Procedimentos) InverteAtivoByid() error {
	if p.ID_Procedimento == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	if err := p.GetAtivoByid(); err != nil {
		return err
	}

	if err := auxiliar.InverteEstado(&p.Ativo); err != nil {
		return err
	}

	if err := p.SetAtivoByid(); err != nil {
		return err
	}
	return nil
}

// funcoes internas ===========================================================

func sProcedimentosToProcedimentos(sp SProcedimentos) (p Procedimentos) {
	p.ID_Procedimento = sp.ID_Procedimento.String
	p.ID_Franqueado = sp.ID_Franqueado.String
	p.FraNome = sp.FraNome.String
	p.ID_Cliente = sp.ID_Cliente.String

	if sp.ID_Cliente.String == "TODOS" {
		p.CliNome = "TODOS"
	} else {
		p.CliNome = sp.CliNome.String
	}

	p.Grupo = sp.Grupo.String
	p.Descricao = sp.Descricao.String

	if sp.DataCadastro.Valid {
		p.DataCadastro = sp.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		p.DataCadastro = ""
	}

	p.Ativo = sp.Ativo.String
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			procedimentos.ID_Procedimento,
			procedimentos.ID_Franqueado,
			franqueado.RazaoSocial,
			procedimentos.ID_Cliente,
			cliente.Nome,
			procedimentos.Grupo,
			procedimentos.Descricao,
			procedimentos.DataCadastro,
			procedimentos.Ativo
		FROM procedimentos

		LEFT JOIN franqueado
		ON procedimentos.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN cliente
		ON procedimentos.ID_Cliente = cliente.ID_Cliente
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, p *Procedimentos) error {
	var tmp SProcedimentos

	if err := tab.Scan(
		&tmp.ID_Procedimento,
		&tmp.ID_Franqueado,
		&tmp.FraNome,
		&tmp.ID_Cliente,
		&tmp.CliNome,
		&tmp.Grupo,
		&tmp.Descricao,
		&tmp.DataCadastro,
		&tmp.Ativo,
	); err != nil {
		return err
	}

	*p = sProcedimentosToProcedimentos(tmp)

	return nil
}
