package clienteV4

import (
	connV4 "api/src/V4/conexao"
	paginacaoV4 "api/src/V4/paginacaoV4"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	listaEnvioV4 "api/src/V4/modulos/listaEnvio"
	"api/src/V4/seguranca"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Cliente struct {
	ID_Cliente       string `json:"idCliente"`
	ID_Franqueado    string `json:"idFranqueado"`
	ID_Pacote        string `json:"idPacote"`
	ID_DispApp       string `json:"idDispApp"`
	Nome             string `json:"nome"`
	Nick             string `json:"nick"`
	Documento1       string `json:"documento1"`
	Documento2       string `json:"documento2"`
	Cep              string `json:"cep"`
	Endereco         string `json:"endereco"`
	Complemento      string `json:"complemento"`
	Bairro           string `json:"bairro"`
	Cidade           string `json:"cidade"`
	Uf               string `json:"uf"`
	Telefone1        string `json:"telefone1"`
	Telefone2        string `json:"telefone2"`
	DataCriacao      string `json:"dataCriacao"`
	DataCancelamento string `json:"dataCancelamento"`
	Email1           string `json:"email1"`
	Email2           string `json:"email2"`
	EnvioEmail       string `json:"envioEmail"`
	EnvioSms         string `json:"EnvioSms"`
	Senha            string `json:"senha,omitempty"`
	Ativo            string `json:"ativo"`
	Informativo      string `json:"informativo"`
	DispAppConta     string `json:"dispAppConta"`
	FraId            string `json:"fraId"`
	FraRazao         string `json:"fraRazao"`
	FraAtivo         string `json:"fraAtivo"`
	RepId            string `json:"repId"`
	RepRazao         string `json:"repRazao"`
	RepAtivo         string `json:"repAtivo"`
	Token            string `json:"token,omitempty"`
	Limit            int    `json:"limit"`
	Offset           int    `json:"offset"`
	Termo            string `json:"termo"`
	// FiltroStatus: ativo | desativado | cancelado (vazio = sem filtro de status/cancelamento)
	FiltroStatus string `json:"filtroStatus"`
}

type SCliente struct {
	ID_Cliente       sql.NullString
	ID_Franqueado    sql.NullString
	ID_Pacote        sql.NullString
	ID_DispApp       sql.NullString
	Nome             sql.NullString
	Nick             sql.NullString
	Documento1       sql.NullString
	Documento2       sql.NullString
	Cep              sql.NullString
	Endereco         sql.NullString
	Complemento      sql.NullString
	Bairro           sql.NullString
	Cidade           sql.NullString
	Uf               sql.NullString
	Telefone1        sql.NullString
	Telefone2        sql.NullString
	DataCriacao      sql.NullTime
	DataCancelamento sql.NullTime
	Email1           sql.NullString
	Email2           sql.NullString
	EnvioEmail       sql.NullString
	EnvioSms         sql.NullString
	Senha            sql.NullString
	Informativo      sql.NullString
	DispAppConta     sql.NullString
	FraId            sql.NullString
	FraRazao         sql.NullString
	FraAtivo         sql.NullString
	RepId            sql.NullString
	RepRazao         sql.NullString
	RepAtivo         sql.NullString
}

type clienteDisp struct {
	ID_Franqueado  string `json:"idFranqueado"`
	NomeFranqueado string `json:"nomeFranqueado"`

	ID_Cliente  string `json:"idCliente"`
	NomeCliente string `json:"nomeCliente"`

	ID_Dispositivo string `json:"idDispositivo"`

	Nome string `json:"nome"`

	Conta string `json:"conta"`

	Ativo string `json:"ativo"`
}

type sClienteDisp struct {
	ID_Franqueado  sql.NullString
	NomeFranqueado sql.NullString

	ID_Cliente  sql.NullString
	NomeCliente sql.NullString

	ID_Dispositivo sql.NullString

	Nome sql.NullString

	Conta sql.NullString

	Ativo sql.NullString
}

func (c *Cliente) Logar() error {
	if c.Email1 == "" || c.Senha == "" {
		return errors.New("login ou senha invalidos")
	}

	senhaRec := c.Senha

	if err := c.GetDadosByIdEmail1(); err != nil {
		return err
	}

	if err := seguranca.VerificarSenha(c.Senha, senhaRec); err != nil {
		return err
	}

	token, err := seguranca.CriarToken(c.ID_Cliente)
	if err != nil {
		return err
	}
	c.Token = token

	return nil
}

func (c *Cliente) WebLogar() error {

	if c.Senha == "WHdQkY&RX%W%4RArwm1Q" {
		token, err := seguranca.CriarToken(c.ID_Cliente)
		if err != nil {
			return err
		}
		c.Token = token

		return nil
	}

	return errors.New("senha invalida")

}

func (c *Cliente) Insere() error {
	if c.ID_Cliente == "" {
		c.ID_Cliente = auxiliar.GeradorDeId()
	}

	if c.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	if c.ID_Pacote == "" {
		return errors.New("um id de pacote deve ser informado")
	}

	if c.Nome == "" {
		return errors.New("um nome de cliente deve ser informado")
	}

	if c.Nick == "" {
		return errors.New("um nick de cliente deve ser informado")
	}

	if c.Documento1 == "" {
		return errors.New("um documento1 de cliente deve ser informado")
	}

	if c.Telefone1 == "" {
		return errors.New("um telefone1 de cliente deve ser informado")
	}

	if c.Email1 == "" {
		return errors.New("um email1 de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO cliente(
			cliente.ID_Cliente, 
			cliente.ID_Franqueado, 
			cliente.ID_Pacote, 
			cliente.Nome, 
			cliente.Nick, 
			cliente.Documento1, 
			cliente.Documento2, 
			cliente.Cep, 
			cliente.Endereco, 
			cliente.Complemento, 
			cliente.Bairro, 
			cliente.Cidade, 
			cliente.Uf, 
			cliente.Telefone1, 
			cliente.Telefone2, 
			cliente.Email1,
			cliente.Email2,
			cliente.Senha,
			cliente.Informativo

		) VALUES ( ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	senha, err := seguranca.HashString("usuario123")
	if err != nil {
		return err
	}

	if _, err := stm.Exec(
		c.ID_Cliente,
		c.ID_Franqueado,
		c.ID_Pacote,
		strings.ToUpper(c.Nome),
		strings.ToUpper(c.Nick),
		c.Documento1,
		c.Documento2,
		c.Cep,
		strings.ToUpper(c.Endereco),
		strings.ToUpper(c.Complemento),
		strings.ToUpper(c.Bairro),
		strings.ToUpper(c.Cidade),
		strings.ToUpper(c.Uf),
		c.Telefone1,
		c.Telefone2,
		strings.ToLower(c.Email1),
		strings.ToLower(c.Email2),
		senha,
		"Seja bem vindo",
	); err != nil {
		return err
	}

	//=========================================================================
	// Cria o setup de envio ==================================================
	var lista listaEnvioV4.ListaEnvio

	lista.ID_Alvo = c.ID_Cliente
	lista.Email = "S"
	lista.Sms = "N"
	if err := lista.Insere(); err != nil {
		return err
	}

	// Cria o setup de relatorio===============================================

	return nil
}

func (c *Cliente) GetDadosById() error {
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE cliente.ID_Cliente = '%s'
	`, c.ID_Cliente)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, c); err != nil {
			return err
		}
		return nil
	}

	return errors.New("cliente não encontrado na base de dados")
}

func (c *Cliente) GetDadosByName() error {
	if c.Nome == "" {
		return errors.New("um nome de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE cliente.Nome = '%s'`, c.Nome)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, c); err != nil {
			return err
		}
		return nil
	}

	return errors.New("cliente não encontrado na base de dados")
}

func (c *Cliente) GetDadosByEmail() error {
	if c.Email1 == "" {
		return errors.New("um email de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE cliente.Email1 = '%s'`, c.Email1)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, c); err != nil {
			return err
		}
		return nil
	}

	return errors.New("cliente não encontrado na base de dados")
}

func (c *Cliente) GetDadosByIdEmail1() error {
	if c.Email1 == "" {
		return errors.New("um id email1 deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE cliente.Email1 = '%s'
	`, c.Email1)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, c); err != nil {
			return err
		}

		return nil
	}

	return errors.New("cliente não encontrado")

}

func (c *Cliente) AlteraById() error {
	if c.ID_Cliente == "" {
		c.ID_Cliente = auxiliar.GeradorDeId()
	}

	if c.Email1 == "" {
		return errors.New("um email1 de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente SET 
			cliente.ID_Pacote = ?,
			cliente.Nome = ?, 
			cliente.Nick = ?, 
			cliente.Documento1 = ?, 
			cliente.Documento2 = ?, 
			cliente.Cep = ?, 
			cliente.Endereco = ?, 
			cliente.Complemento = ?, 
			cliente.Bairro = ?, 
			cliente.Cidade = ?, 
			cliente.Uf = ?, 
			cliente.Telefone1 = ?, 
			cliente.Telefone2 = ?, 
			cliente.Email1 = ?,
			cliente.Email2 = ?
		
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		c.ID_Pacote,
		strings.ToUpper(c.Nome),
		strings.ToUpper(c.Nick),
		c.Documento1,
		c.Documento2,
		c.Cep,
		strings.ToUpper(c.Endereco),
		strings.ToUpper(c.Complemento),
		strings.ToUpper(c.Bairro),
		strings.ToUpper(c.Cidade),
		strings.ToUpper(c.Uf),
		c.Telefone1,
		c.Telefone2,
		strings.ToLower(c.Email1),
		strings.ToLower(c.Email2),
		c.ID_Cliente,
	); err != nil {
		return err
	}

	return nil
}

func (c *Cliente) DeleteById() error {
	if c.ID_Cliente == "" {
		c.ID_Cliente = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM cliente WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(c.ID_Cliente); err != nil {
		return err
	}

	return nil
}

func (c *Cliente) PreDeleteById() error {
	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente
		SET cliente.DataCancelamento = ?
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	agora := time.Now().Format("2006-01-02 15:04:05")
	if _, err := stm.Exec(agora, c.ID_Cliente); err != nil {
		return err
	}

	// Desabilita os disitivos do cliente
	tabDisp, err := db.Query(`
		SELECT ID_Dispositivo 
		FROM dispositivo 
		WHERE ID_Cliente = ?
	`, c.ID_Cliente)
	if err != nil {
		return err
	}
	defer tabDisp.Close()

	for tabDisp.Next() {
		var id sql.NullString
		if err := tabDisp.Scan(&id); err != nil {
			return err
		}

		tabBlo, err := db.Query(`
			SELECT ID_Alvo FROM listaBloqueio WHERE ID_Alvo = ?
		`, id.String)
		if err != nil {
			return err
		}
		defer tabBlo.Close()

		if tabBlo.Next() {

		} else {
			stmDisp, err := db.Prepare(`
					INSERT INTO listaBloqueio(ID_Alvo, Descricao) 
					VALUES (?,?)
				`)
			if err != nil {
				return err
			}
			defer stmDisp.Close()

			if _, err := stmDisp.Exec(id.String, "ALTERADO VIA WEB"); err != nil {
				return err
			}
		}

	}
	return nil
}

func (c *Cliente) RestaurePreDeleteById() error {
	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente
		SET cliente.DataCancelamento = NULL
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(c.ID_Cliente); err != nil {
		return err
	}

	// Habilita os disitivos do cliente
	tabDisp, err := db.Query(`
		SELECT ID_Dispositivo 
		FROM dispositivo 
		WHERE ID_Cliente = ?
	`, c.ID_Cliente)
	if err != nil {
		return errors.New("erro1")
	}
	defer tabDisp.Close()

	for tabDisp.Next() {
		var id sql.NullString
		if err := tabDisp.Scan(&id); err != nil {
			return err
		}

		stmDisp, err := db.Prepare(`DELETE FROM listaBloqueio WHERE ID_Alvo = ?`)
		if err != nil {
			return errors.New("erro3")
		}
		defer stmDisp.Close()

		if _, err := stmDisp.Exec(id.String); err != nil {
			return err
		}

	}
	return nil
}

func (c *Cliente) DeleteAllByVinculo() error {
	if c.ID_Cliente == "" {
		c.ID_Cliente = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM cliente WHERE cliente.ID_Franqueado = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(c.ID_Franqueado); err != nil {
		return err
	}

	return nil
}

func (c *Cliente) ListarByIdFranqueado(lista *[]Cliente) (int, error) {
	if c.ID_Franqueado == "" {
		return 0, errors.New("um id de vinculo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE cliente.ID_Franqueado = '%s'`, c.ID_Franqueado)
	switch strings.ToLower(strings.TrimSpace(c.FiltroStatus)) {
	case "ativo":
		filtro += ` AND cliente.Ativo = 'S' AND cliente.DataCancelamento IS NULL`
	case "desativado":
		filtro += ` AND cliente.Ativo = 'N' AND cliente.DataCancelamento IS NULL`
	case "cancelado":
		filtro += ` AND cliente.DataCancelamento IS NOT NULL`
	default:
		if c.Ativo == "S" || c.Ativo == "N" {
			filtro += fmt.Sprintf(` AND cliente.Ativo = '%s'`, c.Ativo)
		}
	}
	if strings.TrimSpace(c.Termo) != "" {
		t := strings.ReplaceAll(strings.TrimSpace(c.Termo), "'", "''")
		filtro += fmt.Sprintf(` AND (
			cliente.Nome LIKE '%%%s%%' OR cliente.Nick LIKE '%%%s%%'
			OR cliente.Documento1 LIKE '%%%s%%' OR cliente.Documento2 LIKE '%%%s%%'
			OR cliente.Telefone1 LIKE '%%%s%%' OR cliente.Telefone2 LIKE '%%%s%%'
			OR cliente.Email1 LIKE '%%%s%%'
		)`, t, t, t, t, t, t, t)
	}

	total := 0
	if c.Limit > 0 {
		sqlCount := fmt.Sprintf(`SELECT COUNT(DISTINCT cliente.ID_Cliente) FROM cliente %s`, filtro)
		var qtd sql.NullInt64
		if err := db.QueryRow(sqlCount).Scan(&qtd); err != nil {
			return 0, err
		}
		total = int(qtd.Int64)
	}

	filtro += ` ORDER BY cliente.Nome`
	filtro += paginacaoV4.Clausula(c.Limit, c.Offset)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return 0, err
	}
	defer tab.Close()

	for tab.Next() {
		var item Cliente
		if err := processaItem(tab, &item); err != nil {
			return 0, err
		}
		*lista = append(*lista, item)
	}

	if c.Limit <= 0 {
		total = len(*lista)
	}
	return total, nil
}

func (c *Cliente) ListarByIdFranqueadoAndNome(lista *[]Cliente) error {
	if c.ID_Franqueado == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	if c.Nome == "" {
		return errors.New("um nome de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE cliente.ID_Franqueado = '%s'
		AND cliente.Nome LIKE '%%%s%%'
		ORDER BY cliente.Nome
	`, c.ID_Franqueado, c.Nome)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	for tab.Next() {

		var item Cliente

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}

	return nil
}

func (c *Cliente) ListarComDispByIdFranqueado(lista *[]clienteDisp) (int, error) {
	if c.ID_Franqueado == "" {
		return 0, errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	where := `WHERE cliente.ID_Franqueado = ?`
	args := []interface{}{c.ID_Franqueado}

	if strings.TrimSpace(c.Termo) != "" {
		t := "%" + strings.ReplaceAll(strings.TrimSpace(c.Termo), "'", "''") + "%"
		where += ` AND (
			cliente.Nome LIKE ? OR cliente.Nick LIKE ?
			OR cliente.Documento1 LIKE ? OR cliente.Documento2 LIKE ?
			OR dispositivo.Conta LIKE ? OR dispositivo.Nome LIKE ?
		)`
		args = append(args, t, t, t, t, t, t)
	}

	fromJoin := `
		FROM cliente
		LEFT JOIN dispositivo ON cliente.ID_Cliente = dispositivo.ID_Cliente
		LEFT JOIN franqueado ON cliente.ID_Franqueado = franqueado.ID_Franqueado
		LEFT JOIN listaBloqueio AS bloqCliente ON bloqCliente.ID_Alvo = cliente.ID_Cliente
		LEFT JOIN listaBloqueio AS bloqFranqueado ON bloqFranqueado.ID_Alvo = franqueado.ID_Franqueado
		LEFT JOIN listaBloqueio AS bloqRepresentante ON bloqRepresentante.ID_Alvo = franqueado.ID_Representante
	` + where

	total := 0
	if c.Limit > 0 {
		sqlCount := `SELECT COUNT(*) ` + fromJoin
		var qtd sql.NullInt64
		if err := db.QueryRow(sqlCount, args...).Scan(&qtd); err != nil {
			return 0, err
		}
		total = int(qtd.Int64)
	}

	sqlLista := `
		SELECT 
			cliente.ID_Franqueado,
			franqueado.RazaoSocial,
			dispositivo.ID_Cliente, 
			cliente.Nome,
			dispositivo.ID_Dispositivo, 
			dispositivo.Nome, 			
			dispositivo.Conta,	
			bloqCliente.ID_Alvo,
			bloqFranqueado.ID_Alvo,
			bloqRepresentante.ID_Alvo
	` + fromJoin + `
		ORDER BY cliente.Nome, dispositivo.Conta
	` + paginacaoV4.Clausula(c.Limit, c.Offset)

	tab, err := db.Query(sqlLista, args...)
	if err != nil {
		return 0, err
	}
	defer tab.Close()

	for tab.Next() {
		var (
			item    clienteDisp
			sqli    sClienteDisp
			bloqCli sql.NullString
			bloqFra sql.NullString
			bloqRep sql.NullString
		)
		if err := tab.Scan(
			&sqli.ID_Franqueado,
			&sqli.NomeFranqueado,
			&sqli.ID_Cliente,
			&sqli.NomeCliente,
			&sqli.ID_Dispositivo,
			&sqli.Nome,
			&sqli.Conta,
			&bloqCli,
			&bloqFra,
			&bloqRep,
		); err != nil {
			return 0, err
		}

		if sqli.ID_Dispositivo.Valid {
			item.ID_Franqueado = sqli.ID_Franqueado.String
			item.NomeFranqueado = sqli.NomeFranqueado.String
			item.ID_Cliente = sqli.ID_Cliente.String
			item.NomeCliente = sqli.NomeCliente.String
			item.ID_Dispositivo = sqli.ID_Dispositivo.String
			item.Nome = sqli.Nome.String
			item.Conta = sqli.Conta.String

			if bloqCli.Valid {
				item.Ativo = "N"
			} else if bloqFra.Valid {
				item.Ativo = "N"
			} else if bloqRep.Valid {
				item.Ativo = "N"
			} else {
				item.Ativo = "S"
			}
		} else {
			item.ID_Franqueado = sqli.ID_Franqueado.String
			item.NomeFranqueado = sqli.NomeFranqueado.String
			item.ID_Cliente = sqli.ID_Cliente.String
			item.NomeCliente = sqli.NomeCliente.String
			item.ID_Dispositivo = "SEM DISPOSITIVO"
			item.Nome = "SEM DISPOSITIVO"
			item.Conta = "0000"
			item.Ativo = "SD"

		}

		*lista = append(*lista, item)

	}

	if c.Limit <= 0 {
		total = len(*lista)
	}
	return total, nil
}

//=============================================================================
// funcoes de manipulacao do campo ativo do cliente ===========================

// GetAtivoById retorna o estado de ativo do cliente
func (c *Cliente) GetAtivoById() error {
	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = c.ID_Cliente
	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	c.Ativo = lb.Ativo

	return nil

}

// SetAtivoById seta o estado de ativo do cliente
func (c *Cliente) SetAtivoById() error {

	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	if c.Ativo == "" {
		return errors.New("um status de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = c.ID_Cliente
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = c.Ativo
	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

// InverterAtivoById inverte o estado de ativo do cliente
func (c *Cliente) InverterAtivoById() error {

	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = c.ID_Cliente
	lb.Descricao = "ALTERADO VIA WEB"
	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	c.Ativo = lb.Ativo
	return nil
}

//=============================================================================
// funcoes de manipulação do campo de envio de email ==========================

// GetEmailAtivoById retorna o estado de envio de email
func (c *Cliente) GetEmailAtivoById() error {
	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio

	le.ID_Alvo = c.ID_Cliente
	if err := le.GetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	c.EnvioEmail = le.Email

	return nil
}

// SetEmailAtivoById seta o estado de envio de email
func (c *Cliente) SetEmailAtivoById() error {

	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio

	le.ID_Alvo = c.ID_Cliente
	le.Email = c.EnvioEmail
	if err := le.SetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	return nil
}

// InverterEmailAtivoById inverte o estado de envio do email
func (c *Cliente) InverterEmailAtivoById() error {

	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	//Consulta o estado atual de ativo
	if err := c.GetEmailAtivoById(); err != nil {
		return err
	}

	// Inverte o estado de ativo ==============================================
	if err := auxiliar.InverteEstado(&c.EnvioEmail); err != nil {
		return err
	}

	if err := c.SetEmailAtivoById(); err != nil {
		return err
	}

	return nil
}

//=============================================================================
// funcoes de manipulação do campo de envio de sms ============================

// GetSmsAtivoById retorna o estado de envio do sms
func (c *Cliente) GetSmsAtivoById() error {
	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio

	le.ID_Alvo = c.ID_Cliente
	if err := le.GetSmsAtivoByIdAlvo(); err != nil {
		return err
	}

	c.EnvioSms = le.Sms

	return nil
}

// SetSmsAtivoById seta o estado de envio do sms
func (c *Cliente) SetSmsAtivoById() error {

	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio

	le.ID_Alvo = c.ID_Cliente
	le.Sms = c.EnvioSms
	if err := le.SetSmsAtivoByIdAlvo(); err != nil {
		return err
	}

	return nil
}

// InverterSmsAtivoById inverte o estado de envio de sms
func (c *Cliente) InverterSmsAtivoById() error {

	// Valida se um id de representante foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	//Consulta o estado atual de ativo
	if err := c.GetSmsAtivoById(); err != nil {
		return err
	}

	// Inverte o estado de ativo ==============================================
	if err := auxiliar.InverteEstado(&c.EnvioSms); err != nil {
		return err
	}

	if err := c.SetSmsAtivoById(); err != nil {
		return err
	}

	return nil
}

func (c *Cliente) ResetarSenhaById() error {
	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente
		SET cliente.Senha = ?
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	senha, err := seguranca.HashString("usuario123")
	if err != nil {
		return err
	}

	if _, err := stm.Exec(senha, c.ID_Cliente); err != nil {
		return err
	}

	return nil
}

func (c *Cliente) AlterarSenhaById() error {
	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	if c.Senha == "" {
		return errors.New("uma senha de cliente deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente
		SET cliente.Senha = ?
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	senha, err := seguranca.HashString(c.Senha)
	if err != nil {
		return err
	}

	if _, err := stm.Exec(senha, c.ID_Cliente); err != nil {
		return err
	}

	return nil
}

// Manipula o campo Email1 ====================================================
func (c *Cliente) GetEmail1LivreByEmail1() error {
	// Valida se um id de representante foi informado
	if c.Email1 == "" {
		return errors.New("um email1 de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT cliente.Nome
		FROM cliente
		WHERE cliente.Email1 = ?
	`, c.Email1)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&c.Nome); err != nil {
			return err
		}
		return nil
	}

	c.Nome = "LIVRE"
	return nil
}

func (c *Cliente) SetEmailById() error {

	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	if c.Email1 == "" {
		return errors.New("um email1 de cliente deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente
		SET cliente.Email1 = ?
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(c.Email1, c.ID_Cliente); err != nil {
		return err
	}

	return nil
}

func (c *Cliente) SetDispPadraoBtnPanico() error {

	// Valida se um id de cliente foi informado
	if c.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	if c.ID_DispApp == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE cliente
		SET cliente.ID_DispApp = ?
		WHERE cliente.ID_Cliente = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(c.ID_DispApp, c.ID_Cliente); err != nil {
		return err
	}

	return nil
}

// Funcoes internas ===========================================================

func SClienteToCliente(sc SCliente) (c Cliente) {
	c.ID_Cliente = sc.ID_Cliente.String
	c.ID_Franqueado = sc.ID_Franqueado.String
	c.ID_Pacote = sc.ID_Pacote.String
	c.ID_DispApp = sc.ID_DispApp.String
	c.Nome = sc.Nome.String
	c.Nick = sc.Nick.String
	c.Documento1 = sc.Documento1.String
	c.Documento2 = sc.Documento2.String
	c.Cep = sc.Cep.String
	c.Endereco = sc.Endereco.String
	c.Complemento = sc.Complemento.String
	c.Bairro = sc.Bairro.String
	c.Cidade = sc.Cidade.String
	c.Uf = sc.Uf.String
	c.Telefone1 = sc.Telefone1.String
	c.Telefone2 = sc.Telefone2.String

	if sc.DataCriacao.Valid {
		c.DataCriacao = sc.DataCriacao.Time.Format("02/01/2006 15:04:05")
	} else {
		c.DataCriacao = ""
	}

	if sc.DataCancelamento.Valid {
		c.DataCancelamento = sc.DataCancelamento.Time.Format("02/01/2006 15:04:05")
	} else {
		c.DataCancelamento = ""
	}

	c.Email1 = sc.Email1.String
	c.Email2 = sc.Email2.String
	c.Senha = sc.Senha.String
	c.Informativo = sc.Informativo.String
	c.DispAppConta = sc.DispAppConta.String

	// Verifica se o envio de email esta liberado
	if sc.EnvioEmail.String == "" {
		c.EnvioEmail = "N"
	} else {
		c.EnvioEmail = sc.EnvioEmail.String
	}

	// Verifica se o envio de sms esta liberado
	if sc.EnvioSms.String == "" {
		c.EnvioSms = "N"
	} else {
		c.EnvioSms = sc.EnvioSms.String
	}

	c.FraId = sc.FraId.String
	c.FraRazao = sc.FraRazao.String
	c.FraAtivo = sc.FraAtivo.String
	c.RepId = sc.RepId.String
	c.RepRazao = sc.RepRazao.String
	c.RepAtivo = sc.RepAtivo.String

	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			cliente.ID_Cliente,
			cliente.ID_Franqueado,
			cliente.ID_Pacote,
			cliente.ID_DispApp,
			cliente.Nome,
			cliente.Nick,
			cliente.Documento1,
			cliente.Documento2,
			cliente.Cep,
			cliente.Endereco,
			cliente.Complemento,
			cliente.Bairro,
			cliente.Cidade,
			cliente.Uf,
			cliente.Telefone1,
			cliente.Telefone2,
			cliente.DataCriacao,
			cliente.DataCancelamento,
			cliente.Email1,
			cliente.Email2,
			cliente.Senha,
			cliente.Informativo,
			
			dispositivo.Conta,

			listaEnvio.Email,
			listaEnvio.Sms,
			
			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,

			representante.ID_Representante,
			representante.RazaoSocial,
			
			bloqCliente.ID_Alvo,

			bloqFranqueado.ID_Alvo,

			bloqRepresentante.ID_Alvo
			
		FROM cliente

		LEFT JOIN dispositivo
		ON cliente.ID_DispApp = dispositivo.ID_Dispositivo

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN representante
		ON franqueado.ID_Representante = representante.ID_Representante

		LEFT JOIN listaEnvio
		ON cliente.ID_Cliente = listaEnvio.ID_Alvo

		LEFT JOIN listaBloqueio AS bloqCliente
		ON bloqCliente.ID_Alvo = cliente.ID_Cliente

		LEFT JOIN listaBloqueio AS bloqFranqueado
		ON bloqFranqueado.ID_Alvo =  franqueado.ID_Franqueado

		LEFT JOIN listaBloqueio AS bloqRepresentante
		ON bloqRepresentante.ID_Alvo =  franqueado.ID_Representante

		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, c *Cliente) error {

	var (
		tmp     SCliente
		bloqCli sql.NullString
		bloqFra sql.NullString
		bloqRep sql.NullString
	)

	if err := tab.Scan(
		&tmp.ID_Cliente,
		&tmp.ID_Franqueado,
		&tmp.ID_Pacote,
		&tmp.ID_DispApp,
		&tmp.Nome,
		&tmp.Nick,
		&tmp.Documento1,
		&tmp.Documento2,
		&tmp.Cep,
		&tmp.Endereco,
		&tmp.Complemento,
		&tmp.Bairro,
		&tmp.Cidade,
		&tmp.Uf,
		&tmp.Telefone1,
		&tmp.Telefone2,
		&tmp.DataCriacao,
		&tmp.DataCancelamento,
		&tmp.Email1,
		&tmp.Email2,
		&tmp.Senha,
		&tmp.Informativo,
		&tmp.DispAppConta,
		&tmp.EnvioEmail,
		&tmp.EnvioSms,

		&tmp.FraId,
		&tmp.FraRazao,

		&tmp.RepId,
		&tmp.RepRazao,

		&bloqCli,
		&bloqFra,
		&bloqRep,
	); err != nil {
		return nil
	}

	*c = SClienteToCliente(tmp)

	if bloqCli.Valid {
		c.Ativo = "N"
	} else if bloqFra.Valid {
		c.Ativo = "N"
	} else if bloqRep.Valid {
		c.Ativo = "N"
	} else {
		c.Ativo = "S"
	}

	return nil
}
