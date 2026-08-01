//=====================================================\\
// Status possiveis:
//			"NOVO"
//			"AGUARDANDO RESPOSTA"
//			"RESPONDIDO"
//			"FECHADO"
//=====================================================\\

package ticketV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Ticket struct {
	ID_Ticket    string `json:"idTicket"`
	ID_Master    string `json:"idMaster"`
	ID_Slave     string `json:"idSlave"`
	TipoSlave    string `json:"tipoSlave"`
	DataCadastro string `json:"dataCadastro"`
	Assunto      string `json:"assunto"`
	Descricao    string `json:"descricao"`
	Status       string `json:"status"`
	Nome         string `json:"nome"`
}

type STicket struct {
	ID_Ticket    sql.NullString
	ID_Master    sql.NullString
	ID_Slave     sql.NullString
	TipoSlave    sql.NullString
	DataCadastro sql.NullTime
	Assunto      sql.NullString
	Descricao    sql.NullString
	Status       sql.NullString
}

func (t *Ticket) Insere() error {
	// Cria um id para o ticket caso não receba um
	if t.ID_Ticket == "" {
		t.ID_Ticket = auxiliar.GeradorDeId()
	}

	// Valida se um id de master foi informado
	if t.ID_Master == "" {
		return errors.New("um id de master deve ser informado")
	}

	// Valida se um id de slave foi informado
	if t.ID_Slave == "" {
		return errors.New("um id de slave deve ser informado")
	}

	// Abre uma conexao com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepara o comando para ser executado
	stm, err := db.Prepare(`
		INSERT INTO ticket(
			ticket.ID_Ticket, 
			ticket.ID_Master, 
			ticket.ID_Slave,  
			ticket.Assunto, 
			ticket.Descricao, 
			ticket.Status
		) VALUES (?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	// Forca o status em novo
	t.Status = "NOVO"

	// Executa o comando de inserção
	if _, err := stm.Exec(
		t.ID_Ticket,
		t.ID_Master,
		t.ID_Slave,
		strings.ToUpper(t.Assunto),
		strings.ToUpper(t.Descricao),
		strings.ToUpper(t.Status),
	); err != nil {
		return err
	}

	return nil
}

// GetDadosById busca os dado de um ticket pelo seu id
func (t *Ticket) GetDadosById() error {
	if t.ID_Ticket == "" {
		return errors.New("um id de ticket deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE ticket.ID_Ticket = '%s'", t.ID_Ticket)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, t); err != nil {
			return err
		}
		return nil
	}
	return errors.New("ticket não encontrado na base de dados")
}

// ok
// AlteraById altera os dados do ticket pelo seu id
func (t *Ticket) AlteraById() error {
	if t.ID_Ticket == "" {
		return errors.New("um id de ticket deve ser infomado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE ticket 
		SET 
			ticket.Assunto = ?, 
			ticket.Descricao = ?, 
			ticket.Status = ?
		WHERE ticket.ID_Ticket = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(t.Assunto),
		strings.ToUpper(t.Descricao),
		strings.ToUpper(t.Status),
		t.ID_Ticket,
	); err != nil {
		return err
	}
	return nil
}

// DeletaById apaga o ticket vinculado ao id ticket fornecido
func (t *Ticket) DeletaById() error {
	if t.ID_Ticket == "" {
		return errors.New("um id de ticket deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM ticket WHERE ticket.ID_Ticket = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(t.ID_Ticket); err != nil {
		return err
	}

	return nil
}

// DeletaAllByMster apaga todos os tickets vinculado ao id master fornecido
func (t *Ticket) DeletaAllByMster() error {
	if t.ID_Master == "" {
		return errors.New("um id master deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`DELETE FROM ticket WHERE ticket.ID_Master = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(t.ID_Master); err != nil {
		return err
	}

	return nil
}

// DeletaAllBySlave apaga todos os tickets vinculado ao id slave fornecido
func (t *Ticket) DeletaAllBySlave() error {
	if t.ID_Slave == "" {
		return errors.New("um id slave deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`DELETE FROM ticket WHERE ticket.ID_Slave = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(t.ID_Slave); err != nil {
		return err
	}

	return nil
}

// ListarByIdMaster lista os tickets pelo id Master
func (t *Ticket) ListarByIdMaster(lista *[]Ticket) error {
	if t.ID_Master == "" {
		return errors.New("um id de master deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE ticket.ID_Master = '%s'", t.ID_Master)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	var novo, aguardandoResposta, respondido, fechado []Ticket

	for tab.Next() {
		var item Ticket
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		switch item.Status {
		case "NOVO":
			novo = append(novo, item)
		case "AGUARDANDO RESPOSTA":
			aguardandoResposta = append(aguardandoResposta, item)
		case "RESPONDIDO":
			respondido = append(respondido, item)
		case "FECHADO":
			fechado = append(fechado, item)
		}
	}
	*lista = append(*lista, novo...)
	*lista = append(*lista, aguardandoResposta...)
	*lista = append(*lista, respondido...)
	*lista = append(*lista, fechado...)

	return nil
}

// ListarByIdSlave lista os tickets pelo id slave
func (t *Ticket) ListarByIdSlave(lista *[]Ticket) error {
	if t.ID_Slave == "" {
		return errors.New("um id de slave deve ser informado")
	}

	// Abre uma conexao com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE ticket.ID_Slave = '%s'", t.ID_Slave)
	// Consulta o banco

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	// Variaveis auxiliares para ordenar os tickets
	var novo, aguardandoResposta, respondido, fechado []Ticket

	for tab.Next() {
		var item Ticket
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		// Separa o item baseado no status
		switch item.Status {
		case "NOVO":
			novo = append(novo, item)
		case "AGUARDANDO RESPOSTA":
			aguardandoResposta = append(aguardandoResposta, item)
		case "RESPONDIDO":
			respondido = append(respondido, item)
		case "FECHADO":
			fechado = append(fechado, item)
		}
	}
	// Ordena os tickets
	*lista = append(*lista, novo...)
	*lista = append(*lista, aguardandoResposta...)
	*lista = append(*lista, respondido...)
	*lista = append(*lista, fechado...)

	return nil
}

// Funcoes internas ===========================================================
func sTicketToTicket(st STicket) (t Ticket) {
	t.ID_Ticket = st.ID_Ticket.String
	t.ID_Master = st.ID_Master.String
	t.ID_Slave = st.ID_Slave.String
	t.DataCadastro = st.DataCadastro.Time.Format("02/01/2006 15:04:05")
	t.Assunto = st.Assunto.String
	t.Descricao = st.Descricao.String
	t.Status = st.Status.String
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			ticket.ID_Ticket,
			ticket.ID_Master,
			ticket.ID_Slave,
			ticket.DataCadastro,
			ticket.Assunto,
			ticket.Descricao,
			ticket.Status,

			representante.RazaoSocial,
			franqueado.RazaoSocial,
			cliente.Nome

		FROM ticket

		LEFT JOIN representante
		ON ticket.ID_Slave = representante.ID_Representante

		LEFT JOIN franqueado
		ON ticket.ID_Slave = franqueado.ID_Franqueado

		LEFT JOIN cliente
		ON ticket.ID_Slave = cliente.ID_Cliente

		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, t *Ticket) error {
	var (
		tmp     STicket
		nomeRep sql.NullString
		nomeFra sql.NullString
		nomeCli sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Ticket,
		&tmp.ID_Master,
		&tmp.ID_Slave,
		&tmp.DataCadastro,
		&tmp.Assunto,
		&tmp.Descricao,
		&tmp.Status,
		&nomeRep,
		&nomeFra,
		&nomeCli,
	); err != nil {
		return err
	}

	*t = sTicketToTicket(tmp)

	if nomeRep.Valid {
		t.Nome = nomeRep.String
	} else if nomeFra.Valid {
		t.Nome = nomeFra.String
	} else if nomeCli.Valid {
		t.Nome = nomeCli.String
	} else {
		t.Nome = ""
	}

	return nil
}
