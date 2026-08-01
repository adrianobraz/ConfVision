package processoV4

import (
	connV4 "api/src/V4/conexao"
	paginacaoV4 "api/src/V4/paginacaoV4"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type Processo struct {
	ID_Processo    string `json:"idProcesso"`
	DataCriacao    string `json:"dataCriacao"`
	ID_Dispositivo string `json:"idDispositivo"`
	ID_Atendente   string `json:"idAtendente"`
	Nivel          string `json:"nivel"`
	DataAtenInicio string `json:"dataAtenInicio"`
	DataAtenFim    string `json:"dataAtenFim"`
	Descricao      string `json:"descricao"`
	MsgAtendente   string `json:"msgAtendente"`
	DispId         string `json:"dispId"`
	DispNome       string `json:"dispNome"`
	CliNome        string `json:"cliNome"`
}

type SProcesso struct {
	ID_Processo    sql.NullString
	DataCriacao    sql.NullTime
	ID_Dispositivo sql.NullString
	ID_Atendente   sql.NullString
	Nivel          sql.NullString
	DataAtenInicio sql.NullTime
	DataAtenFim    sql.NullTime
	Descricao      sql.NullString
	MsgAtendente   sql.NullString
}

type ProcEvt struct {
	ID_Processo    string `json:"idProcesso"`
	DataCriacao    string `json:"dataCriacao"`
	DataAtenInicio string `json:"dataAtenInicio"`
	DataAtenFim    string `json:"dataAtenFim"`
	Descricao      string `json:"descricao"`
	MsgAtendente   string `json:"msgAtendente"`

	ID_Evento   string `json:"idEvento"`
	Codigo      string `json:"codigo,omitempty"`
	Particao    string `json:"particao,omitempty"`
	ZonaUser    string `json:"zonaUser,omitempty"`
	Nivel       string `json:"nivel"`
	Img         string `json:"img"`
	DataEntrada string `json:"dataEntrada"`

	Nick string `json:"nick"`

	DispNome string `json:"dispNome"`

	CliNome string `json:"cliNome"`
}

type SProcEvt struct {
	ID_Processo    sql.NullString
	DataCriacao    sql.NullString
	DataAtenInicio sql.NullTime
	DataAtenFim    sql.NullTime
	Descricao      sql.NullString
	MsgAtendente   sql.NullString

	ID_Evento   sql.NullString
	Codigo      sql.NullString
	Particao    sql.NullString
	ZonaUser    sql.NullString
	Nivel       sql.NullString
	Img         sql.NullString
	DataEntrada sql.NullTime

	Nick sql.NullString

	DispNome sql.NullString

	CliNome sql.NullString
}

func (p *Processo) Insere() error {

	if p.ID_Processo == "" {
		p.ID_Processo = auxiliar.GeradorDeId()
	}

	if p.ID_Dispositivo == "" {
		return errors.New("um id de despositivo deve ser informado")
	}

	if p.Nivel == "" {
		p.Nivel = "0"
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO processo(
		processo.ID_Processo, 
		processo.ID_Dispositivo, 
		processo.Nivel
		) VALUES ( ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		p.ID_Processo,
		p.ID_Dispositivo,
		p.Nivel,
	); err != nil {
		return err
	}

	return nil
}

func (p *Processo) GetIdProcessoByIdDispositivo() error {
	if p.ID_Dispositivo == "" {
		return errors.New("um id de despositivo deve ser informado")
	}

	// Verifica se existe um processo aberto ==================================
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT processo.ID_Processo 
		FROM processo
		WHERE processo.ID_Dispositivo = ?
		AND processo.DataAtenFim IS NULL
	`, p.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var id sql.NullString
		if err := tab.Scan(&id); err != nil {
			return err
		}

		p.ID_Processo = id.String
		return nil
	}

	// Caso nao exista gera um novo processo ==================================
	p.ID_Processo = auxiliar.GeradorDeId()
	p.Nivel = "0"

	if err := p.Insere(); err != nil {
		return err
	}

	return nil
}

// Controla campo ID_Atendente ================================================
func (p *Processo) GetIdAtendentelById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT processo.ID_Atendente
		FROM processo
		WHERE processo.ID_Processo
	`, p.ID_Processo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&p.ID_Atendente); err != nil {
			return err
		}
	}

	return errors.New("processo não encontrado no banco de dados")
}

func (p *Processo) SetIdAtendenteById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE processo 
		SET processo.ID_Atendente = ?
		WHERE processo.ID_Processo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.ID_Atendente, p.ID_Processo); err != nil {
		return err
	}
	return nil
}

// Controla campo Nivel =======================================================
func (p *Processo) GetNivelById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT processo.Nivel
		FROM processo
		WHERE processo.ID_Processo = ?
	`, p.ID_Processo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&p.Nivel); err != nil {
			return err
		}
		return nil
	}

	return errors.New("processo não encontrado no banco de dados")
}

func (p *Processo) SetNivelById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE processo 
		SET processo.Nivel = ?
		WHERE processo.ID_Processo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.Nivel, p.ID_Processo); err != nil {
		return err
	}
	return nil
}

func (p *Processo) AtualizaNivelById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	// Converte o nivel recebido em inteiro para calculo
	nivelIn, err := strconv.Atoi(p.Nivel)
	if err != nil {
		return err
	}

	if err := p.GetNivelById(); err != nil {
		return err
	}

	// Coverte o nivel consultado no banco para inteiro para calculo
	nivelBanco, err := strconv.Atoi(p.Nivel)
	if err != nil {
		return err
	}

	if nivelIn > nivelBanco {
		p.Nivel = strconv.Itoa(nivelIn)
		if err := p.SetNivelById(); err != nil {
			return err
		}
	}
	return nil
}

// Controla campo MsgAtendente ================================================
func (p *Processo) GetMsgAtendenteById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT processo.MsgAtendente
		FROM processo
		WHERE processo.ID_Processo
	`, p.ID_Processo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&p.MsgAtendente); err != nil {
			return err
		}
	}

	return errors.New("processo não encontrado no banco de dados")
}

func (p *Processo) SetMsgAtendenteById() error {
	if p.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE processo 
		SET processo.MsgAtendente = ?
		WHERE processo.ID_Processo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(p.MsgAtendente, p.ID_Processo); err != nil {
		return err
	}
	return nil
}

func (p *Processo) ListarToOpen(lista *[]Processo) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := `
		WHERE processo.DataAtenFim IS NULL
		ORDER BY processo.Nivel DESC
	`

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Processo

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (p *Processo) ListarToOpenByIdDispositivo(lista *[]Processo) error {
	if p.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE processo.ID_Dispositivo = '%s'  
		AND processo.DataAtenFim IS NULL 
		ORDER BY processo.Nivel DESC
	`, p.ID_Dispositivo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Processo

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (p *Processo) ListarToOpenByIdCliente(idCliente string, lista *[]Processo) error {
	if idCliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE processo.ID_Dispositivo IN (
			SELECT dispositivo.ID_Dispositivo 
			FROM dispositivo
			WHERE dispositivo.ID_Cliente = '%s'
		)
		AND processo.DataAtenFim IS NULL
		ORDER BY processo.Nivel DESC 
	`, idCliente)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Processo

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// 1.1
func (e *Processo) ListarToOpenByCliente(idCliente string, lista *[]Processo) error {
	if idCliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			processo.ID_Processo, 
			processo.ID_Dispositivo, 
			processo.ID_Atendente, 
			processo.Nivel, 
			processo.DataCriacao, 
			processo.DataAtenInicio, 
			processo.DataAtenFim, 
			processo.Descricao, 
			processo.MsgAtendente,
			
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,

			cliente.Nome
		
		FROM processo

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		WHERE cliente.ID_Cliente = ?

		AND processo.DataAtenFim IS NULL

		ORDER BY processo.Nivel DESC 
	`, idCliente)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Processo
		var sItem SProcesso
		var dispNome, dispId, cliNome sql.NullString

		if err := tab.Scan(
			&sItem.ID_Processo,
			&sItem.ID_Dispositivo,
			&sItem.ID_Atendente,
			&sItem.Nivel,
			&sItem.DataCriacao,
			&sItem.DataAtenInicio,
			&sItem.DataAtenFim,
			&sItem.Descricao,
			&sItem.MsgAtendente,
			&dispId,
			&dispNome,
			&cliNome,
		); err != nil {
			return err
		}
		item.ID_Processo = sItem.ID_Processo.String
		item.ID_Dispositivo = sItem.ID_Dispositivo.String
		item.ID_Atendente = sItem.ID_Atendente.String
		item.Nivel = sItem.Nivel.String
		item.DataCriacao = sItem.DataCriacao.Time.Format("02/01/2006 15:04:05")
		item.DataAtenInicio = sItem.DataAtenInicio.Time.Format("02/01/2006 15:04:05")
		item.DataAtenFim = sItem.DataAtenFim.Time.Format("02/01/2006 15:04:05")
		item.Descricao = sItem.Descricao.String
		item.MsgAtendente = sItem.MsgAtendente.String
		item.DispId = dispId.String
		item.DispNome = dispNome.String
		item.CliNome = cliNome.String

		*lista = append(*lista, item)
	}

	return nil
}

func (e *Processo) GetDadosById() error {
	if e.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			processo.ID_Processo, 
			processo.ID_Dispositivo, 
			processo.ID_Atendente, 
			processo.Nivel, 
			processo.DataCriacao, 
			processo.DataAtenInicio, 
			processo.DataAtenFim, 
			processo.Descricao, 
			processo.MsgAtendente,
			
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,

			cliente.Nome
		
		FROM processo

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		WHERE processo.ID_Processo = ?
	`, e.ID_Processo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		var sItem SProcesso
		var dispNome, dispId, cliNome sql.NullString

		if err := tab.Scan(
			&sItem.ID_Processo,
			&sItem.ID_Dispositivo,
			&sItem.ID_Atendente,
			&sItem.Nivel,
			&sItem.DataCriacao,
			&sItem.DataAtenInicio,
			&sItem.DataAtenFim,
			&sItem.Descricao,
			&sItem.MsgAtendente,
			&dispId,
			&dispNome,
			&cliNome,
		); err != nil {
			return err
		}

		e.ID_Processo = sItem.ID_Processo.String
		e.ID_Dispositivo = sItem.ID_Dispositivo.String
		e.ID_Atendente = sItem.ID_Atendente.String
		e.Nivel = sItem.Nivel.String
		e.DataCriacao = sItem.DataCriacao.Time.Format("02/01/2006 15:04:05")
		e.DataAtenInicio = sItem.DataAtenInicio.Time.Format("02/01/2006 15:04:05")
		e.DataAtenFim = sItem.DataAtenFim.Time.Format("02/01/2006 15:04:05")
		e.Descricao = sItem.Descricao.String
		e.MsgAtendente = sItem.MsgAtendente.String
		e.DispId = dispId.String
		e.DispNome = dispNome.String
		e.CliNome = cliNome.String

	}

	return nil
}

func (p *Processo) ListarByFiltro(idCliente string, lista *[]Processo) error {
	if idCliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE processo.ID_Dispositivo IN (
			SELECT dispositivo.ID_Dispositivo 
			FROM dispositivo
			WHERE dispositivo.ID_Cliente = '%s'
		)
		AND processo.DataAtenFim IS NULL
		ORDER BY processo.Nivel DESC 
	`, idCliente)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Processo

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (e *Processo) ListarEventosByFiltro(filtro string, limit, offset int, lista *[]ProcEvt) (int, error) {
	if filtro == "" {
		return 0, errors.New("o filtro informado é invalido")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	baseFrom := `
		FROM processo	
		LEFT JOIN usuarios ON processo.ID_Atendente = usuarios.ID_Usuario
		LEFT JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
		LEFT JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
	`

	total := 0
	if limit > 0 {
		sqlCount := `SELECT COUNT(*) ` + baseFrom + filtro
		var qtd sql.NullInt64
		if err := db.QueryRow(sqlCount).Scan(&qtd); err != nil {
			return 0, err
		}
		total = int(qtd.Int64)
	}

	txtSql := fmt.Sprintf(`
		SELECT 
			processo.ID_Processo, 
			processo.Descricao,
			processo.DataAtenInicio, 
			processo.DataAtenFim, 
			processo.Descricao, 
			processo.MsgAtendente, 
			usuarios.Nick,
			dispositivo.Nome,
			cliente.Nome
		%s
		%s
		ORDER BY processo.DataAtenFim DESC
		%s
	`, baseFrom, filtro, paginacaoV4.Clausula(limit, offset))

	tab, err := db.Query(txtSql)
	if err != nil {
		return 0, err
	}
	defer tab.Close()

	for tab.Next() {
		var sItem SProcEvt
		var item ProcEvt

		if err := tab.Scan(
			&sItem.ID_Processo,
			&sItem.Descricao,
			&sItem.DataAtenInicio,
			&sItem.DataAtenFim,
			&sItem.Descricao,
			&sItem.MsgAtendente,
			&sItem.Nick,
			&sItem.DispNome,
			&sItem.CliNome,
		); err != nil {
			return 0, err
		}

		item.ID_Processo = sItem.ID_Processo.String
		item.Descricao = sItem.Descricao.String
		item.DataAtenInicio = sItem.DataAtenInicio.Time.Format("02/01/2006 15:04:05")
		item.DataAtenFim = sItem.DataAtenFim.Time.Format("02/01/2006 15:04:05")
		item.Descricao = sItem.Descricao.String
		item.MsgAtendente = sItem.MsgAtendente.String
		item.Nick = sItem.Nick.String
		item.DispNome = sItem.DispNome.String
		item.CliNome = sItem.CliNome.String

		*lista = append(*lista, item)
	}
	if limit <= 0 {
		total = len(*lista)
	}
	return total, nil
}

func (e *Processo) FinalizarSistema() error {
	if e.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	if e.Descricao == "" {
		return errors.New("uma descrição deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, erro := db.Prepare(`
		UPDATE processo
		SET 
			processo.Descricao = ?,
			processo.ID_Atendente = ?,
			processo.DataAtenFim = ?

		WHERE processo.ID_Processo = ?
	`)

	if erro != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		e.Descricao,
		"CONFBOT",
		time.Now().Format("2006-01-02 15:04:05"),
		e.ID_Processo,
	); err != nil {
		return err
	}

	return nil
}

// Funcoes internas ===========================================================
func sProcessoToProcesso(proc *Processo, sProc SProcesso) {
	proc.ID_Processo = sProc.ID_Processo.String
	proc.ID_Dispositivo = sProc.ID_Dispositivo.String
	proc.ID_Atendente = sProc.ID_Atendente.String
	proc.Nivel = sProc.Nivel.String

	if sProc.DataCriacao.Valid {
		proc.DataCriacao = sProc.DataCriacao.Time.Format("02/01/2006 15:04:05")
	} else {
		proc.DataCriacao = ""
	}

	if sProc.DataAtenInicio.Valid {
		proc.DataAtenInicio = sProc.DataAtenInicio.Time.Format("02/01/2006 15:04:05")
	} else {
		proc.DataAtenInicio = ""
	}

	if sProc.DataAtenFim.Valid {
		proc.DataAtenFim = sProc.DataAtenFim.Time.Format("02/01/2006 15:04:05")
	} else {
		proc.DataAtenFim = ""
	}

	proc.Descricao = sProc.Descricao.String
	proc.MsgAtendente = sProc.MsgAtendente.String
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			processo.ID_Processo, 
			processo.ID_Dispositivo, 
			processo.ID_Atendente, 
			processo.Nivel, 
			processo.DataCriacao, 
			processo.DataAtenInicio, 
			processo.DataAtenFim, 
			processo.Descricao, 
			processo.MsgAtendente 
		
		FROM processo
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, item *Processo) error {

	var tmp SProcesso
	if err := tab.Scan(
		&tmp.ID_Processo,
		&tmp.ID_Dispositivo,
		&tmp.ID_Atendente,
		&tmp.Nivel,
		&tmp.DataCriacao,
		&tmp.DataAtenInicio,
		&tmp.DataAtenFim,
		&tmp.Descricao,
		&tmp.MsgAtendente,
	); err != nil {
		return err
	}

	sProcessoToProcesso(item, tmp)
	return nil
}
