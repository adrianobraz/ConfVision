package pacotev4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Pacote struct {
	ID_Pacote           string `json:"idPacote"`
	ID_Vinculo          string `json:"idVinculo"`
	Nome                string `json:"nome"`
	Peridiocidade       string `json:"peridiocidade"`
	Valor               string `json:"valor"`
	Comissao            string `json:"comissao"`
	Terminal            string `json:"terminal"`
	Grade               string `json:"grade"`
	GradeValor          string `json:"gradeValor"`
	ContasQtd           string `json:"contasQtd"`
	ContasExedente      string `json:"contasExedente"`
	AtendimentoQtd      string `json:"atendimentoQtd"`
	AtendimentoExedente string `json:"atendimentoExedente"`
	EmailQtd            string `json:"emailQtd"`
	EmailExedente       string `json:"emailExedente"`
	EmailBloquear       string `json:"emailBloquear"`
	SmsQtd              string `json:"smsQtd"`
	SmsExedente         string `json:"smsExedente"`
	SmsBloquear         string `json:"smsBoquear"`
	LigacoesQtd         string `json:"ligacoesQtd"`
	LigacoesExedente    string `json:"ligacoesExedente"`
	LigacoesBloquear    string `json:"ligacoesBloquear"`
	DataCadastro        string `json:"dataCadastro"`
	Ativo               string `json:"ativo"`
}

type SPacote struct {
	ID_Pacote           sql.NullString
	ID_Vinculo          sql.NullString
	Nome                sql.NullString
	Peridiocidade       sql.NullString
	Valor               sql.NullString
	Comissao            sql.NullString
	Terminal            sql.NullString
	Grade               sql.NullString
	GradeValor          sql.NullString
	ContasQtd           sql.NullString
	ContasExedente      sql.NullString
	AtendimentoQtd      sql.NullString
	AtendimentoExedente sql.NullString
	EmailQtd            sql.NullString
	EmailExedente       sql.NullString
	EmailBloquear       sql.NullString
	SmsQtd              sql.NullString
	SmsExedente         sql.NullString
	SmsBloquear         sql.NullString
	LigacoesQtd         sql.NullString
	LigacoesExedente    sql.NullString
	LigacoesBloquear    sql.NullString
	DataCadastro        sql.NullTime
	Ativo               sql.NullString
}

func (p *Pacote) Insere() error {

	if p.ID_Pacote == "" {
		p.ID_Pacote = auxiliar.GeradorDeId()
	}

	if p.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	if p.Comissao == "" {
		p.Comissao = "0"
	}

	if p.Nome == "" {
		return errors.New("um nome de pacote dever ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()
	
	stm, err := db.Prepare(`
		INSERT INTO pacotes(
			pacotes.ID_Pacote, 
			pacotes.ID_Vinculo, 
			pacotes.Nome, 
			pacotes.Peridiocidade, 
			pacotes.Valor, 
			pacotes.Comissao, 
			pacotes.Terminal, 
			pacotes.Grade, 
			pacotes.GradeValor, 
			pacotes.ContasQtd, 
			pacotes.ContasExedente, 
			pacotes.AtendimentoQtd, 
			pacotes.AtendimentoExedente, 
			pacotes.EmailQtd, 
			pacotes.EmailExedente, 
			pacotes.EmailBloquear, 
			pacotes.SmsQtd, 
			pacotes.SmsExedente, 
			pacotes.SmsBloquear, 
			pacotes.LigacoesQtd, 
			pacotes.LigacoesExedente, 
			pacotes.LigacoesBloquear 
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}

	if _, err := stm.Exec(
		p.ID_Pacote,
		p.ID_Vinculo,
		strings.ToUpper(p.Nome),
		p.Peridiocidade,
		p.Valor,
		p.Comissao,

		strings.ToUpper(p.Terminal),
		strings.ToUpper(p.Grade),
		p.GradeValor,

		p.ContasQtd,
		p.ContasExedente,

		p.AtendimentoQtd,
		p.AtendimentoExedente,

		p.EmailQtd,
		p.EmailExedente,
		strings.ToUpper(p.EmailBloquear),

		p.SmsQtd,
		p.SmsExedente,
		strings.ToUpper(p.SmsBloquear),

		p.LigacoesQtd,
		p.LigacoesExedente,
		strings.ToUpper(p.LigacoesBloquear),
	); err != nil {
		return err
	}

	return nil
}

func (p *Pacote) GetDadosById() error {
	if p.ID_Pacote == "" {
		return errors.New("um id de pacote deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE pacotes.ID_Pacote = '%s'`, p.ID_Pacote)
	tab, err := db.Query(p.getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := p.processaItem(tab); err != nil {
			return err
		}
		return nil
	}
	return errors.New("pacote não encontrado na base de dados")
}

func (p *Pacote) AlteraById() error {
	if p.ID_Pacote == "" {
		return errors.New("um id de pacote deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE pacotes SET 
			pacotes.Nome = ?,
			pacotes.Peridiocidade = ?,
			pacotes.Valor = ?, 
			pacotes.Comissao = ?, 
			pacotes.Terminal = ?,
			pacotes.Grade = ?,
			pacotes.GradeValor = ?, 
			pacotes.ContasQtd = ?,
			pacotes.ContasExedente = ?,
			pacotes.AtendimentoQtd = ?,
			pacotes.AtendimentoExedente = ?, 
			pacotes.EmailQtd = ?,
			pacotes.EmailExedente = ?,
			pacotes.EmailBloquear = ?, 
			pacotes.SmsQtd = ?,
			pacotes.SmsExedente = ?,
			pacotes.SmsBloquear = ?,
			pacotes.LigacoesQtd = ?,
			pacotes.LigacoesExedente = ?,
			pacotes.LigacoesBloquear = ? 
		WHERE pacotes.ID_Pacote = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(p.Nome),
		p.Peridiocidade,
		p.Valor,
		p.Comissao,
		strings.ToUpper(p.Terminal),
		strings.ToUpper(p.Grade),
		p.GradeValor,

		p.ContasQtd,
		p.ContasExedente,

		p.AtendimentoQtd,
		p.AtendimentoExedente,

		p.EmailQtd,
		p.EmailExedente,
		strings.ToUpper(p.EmailBloquear),

		p.SmsQtd,
		p.SmsExedente,
		strings.ToUpper(p.SmsBloquear),

		p.LigacoesQtd,
		p.LigacoesExedente,
		strings.ToUpper(p.LigacoesBloquear),

		p.ID_Pacote,
	); err != nil {
		return err
	}

	return nil
}

func (p *Pacote) DeletaById() error {
	return nil
}

func (p *Pacote) DeletaAllByVinculo() error {
	if p.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve informado")
	}
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`DELETE FROM pacotes WHERE pacotes.ID_Vinculo = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.ID_Vinculo); err != nil {
		return err
	}

	return nil
}

func (p *Pacote) ListaByVinculo(lista *[]Pacote) error {
	if p.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE pacotes.ID_Vinculo = '%s'`, p.ID_Vinculo)
	tab, err := db.Query(p.getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Pacote

		if err := item.processaItem(tab); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

func (p *Pacote) GetAtivoById() error {
	if p.ID_Pacote == "" {
		return errors.New("um id de pacote deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = p.ID_Pacote
	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	p.Ativo = lb.Ativo
	return nil
}

func (p *Pacote) SetAtivoById() error {

	if p.ID_Pacote == "" {
		return errors.New("um id de pacote deve ser informado")
	}

	if p.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = p.ID_Pacote
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = p.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return nil
	}

	return nil
}

func (p *Pacote) InverteAtivoById() error {
	if p.ID_Pacote == "" {
		return errors.New("um id de pacote deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = p.ID_Pacote

	if lb.Descricao == "" {
		lb.Descricao = "ALTERADO VIA WEB"
	}

	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	p.Ativo = lb.Ativo
	return nil
}

// Funcoes internas ===========================================================

func sPacoteToPacote(sp SPacote) (p Pacote) {
	p.ID_Pacote = sp.ID_Pacote.String
	p.ID_Vinculo = sp.ID_Vinculo.String
	p.Nome = sp.Nome.String
	p.Peridiocidade = sp.Peridiocidade.String
	p.Valor = sp.Valor.String
	p.Comissao = sp.Comissao.String
	p.Terminal = sp.Terminal.String
	p.Grade = sp.Grade.String
	p.GradeValor = sp.GradeValor.String
	p.ContasQtd = sp.ContasQtd.String
	p.ContasExedente = sp.ContasExedente.String
	p.AtendimentoQtd = sp.AtendimentoQtd.String
	p.AtendimentoExedente = sp.AtendimentoExedente.String
	p.EmailQtd = sp.EmailQtd.String
	p.EmailExedente = sp.EmailExedente.String
	p.EmailBloquear = sp.EmailBloquear.String
	p.SmsQtd = sp.SmsQtd.String
	p.SmsExedente = sp.SmsExedente.String
	p.SmsBloquear = sp.SmsBloquear.String
	p.LigacoesQtd = sp.LigacoesQtd.String
	p.LigacoesExedente = sp.LigacoesExedente.String
	p.LigacoesBloquear = sp.LigacoesBloquear.String
	p.DataCadastro = sp.DataCadastro.Time.Format("02/01/2006 15:04:05")
	return
}

func (p *Pacote) getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			pacotes.ID_Pacote,
			pacotes.ID_Vinculo,
			pacotes.Nome,
			pacotes.Peridiocidade,
			pacotes.Valor,
			pacotes.Comissao,
			pacotes.Terminal,
			pacotes.Grade,
			pacotes.GradeValor,
			pacotes.ContasQtd,
			pacotes.ContasExedente,
			pacotes.AtendimentoQtd,
			pacotes.AtendimentoExedente,
			pacotes.EmailQtd,
			pacotes.EmailExedente,
			pacotes.EmailBloquear,
			pacotes.SmsQtd,
			pacotes.SmsExedente,
			pacotes.SmsBloquear,
			pacotes.LigacoesQtd,
			pacotes.LigacoesExedente,
			pacotes.LigacoesBloquear,
			pacotes.DataCadastro,

			listaBloqueio.ID_Alvo
			
		FROM pacotes

		LEFT JOIN listaBloqueio
		ON pacotes.ID_Pacote = listaBloqueio.ID_Alvo

		%s
	`, filtro)
}

func (p *Pacote) processaItem(tab *sql.Rows) error {
	var (
		tmp        SPacote
		blocPacote sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Pacote,
		&tmp.ID_Vinculo,
		&tmp.Nome,
		&tmp.Peridiocidade,
		&tmp.Valor,
		&tmp.Comissao,
		&tmp.Terminal,
		&tmp.Grade,
		&tmp.GradeValor,
		&tmp.ContasQtd,
		&tmp.ContasExedente,
		&tmp.AtendimentoQtd,
		&tmp.AtendimentoExedente,
		&tmp.EmailQtd,
		&tmp.EmailExedente,
		&tmp.EmailBloquear,
		&tmp.SmsQtd,
		&tmp.SmsExedente,
		&tmp.SmsBloquear,
		&tmp.LigacoesQtd,
		&tmp.LigacoesExedente,
		&tmp.LigacoesBloquear,
		&tmp.DataCadastro,

		&blocPacote,
	); err != nil {
		return err
	}

	*p = sPacoteToPacote(tmp)

	if blocPacote.Valid {
		p.Ativo = "N"
	} else {
		p.Ativo = "S"
	}
	return nil
}
