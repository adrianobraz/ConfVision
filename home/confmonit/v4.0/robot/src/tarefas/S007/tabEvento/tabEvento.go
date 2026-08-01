package tabEvento

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type evento struct {
	idEvento    string
	idProcesso  string
	codigo      string
	particao    string
	zonaUser    string
	nivel       string
	img         string
	DataEntrada time.Time
}

func BackupTabelaEventos(db *sql.DB, qtdMes int, idProcesso string) error {

	tabela := fmt.Sprintf(
		"evento_%s",
		time.Now().AddDate(0, -qtdMes, 0).Format("Jan_2006"),
	)

	if idProcesso == "" {
		return errors.New("um id de processo deve ser informado")
	}

	// Busca os evento referente ao processo
	tab, erro := db.Query(`
		SELECT
			evento.ID_Evento, 
			evento.ID_Processo, 
			evento.Codigo, 
			evento.Particao, 
			evento.ZonaUser, 
			evento.Nivel, 
			evento.Img, 
			evento.DataEntrada 
		FROM evento
		WHERE evento.ID_Processo = ?
	`, idProcesso)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	// Proecessos os eventos retornados
	for tab.Next() {
		// Verifica se a tabela de backup existe e se não cria
		if erro := tabelaExiste(db, tabela); erro != nil {
			return erro
		}

		var e evento

		if erro := tab.Scan(
			&e.idEvento,
			&e.idProcesso,
			&e.codigo,
			&e.particao,
			&e.zonaUser,
			&e.nivel,
			&e.img,
			&e.DataEntrada,
		); erro != nil {
			return erro
		}

		// Grava o evento na tabela backup
		if erro := gravarTabBackup(db, e, tabela); erro != nil {
			return erro
		}

		// // Exclui o evento da tabela evento
		if erro := deletarTabPrincipal(db, e.idEvento); erro != nil {
			return erro
		}

	}

	return nil
}

/////////////////////////// FUNÇOES INTERNAS ///////////////////////////

// gravarTabBackup grava os eventos de um determinado processo na tabela
// backup corrente
func gravarTabBackup(db *sql.DB, dados evento, tabela string) error {
	fmt.Println("	Gravando Evento", dados.idEvento)
	sqlTxt := fmt.Sprintf(`
		INSERT INTO %s(
			ID_Evento, 
			ID_Processo, 
			Codigo, 
			Particao, 
			ZonaUser, 
			Nivel, 
			Img, 
			DataEntrada
	) VALUES (?,?,?,?,?,?,?,?)`, tabela)

	stm, erro := db.Prepare(sqlTxt)
	if erro != nil {
		return erro
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		dados.idEvento,
		dados.idProcesso,
		dados.codigo,
		dados.particao,
		dados.zonaUser,
		dados.nivel,
		dados.img,
		dados.DataEntrada,
	); erro != nil {
		return erro
	}

	return nil
}

// deletarTabPrincipal deleta os eventod de um determinado processo
// da tabela principal.
func deletarTabPrincipal(db *sql.DB, idEvento string) error {

	stm, erro := db.Prepare(`DELETE FROM evento WHERE evento.ID_Evento = ?`)
	if erro != nil {
		return erro
	}
	defer stm.Close()

	if _, erro := stm.Exec(idEvento); erro != nil {
		return erro
	}

	return nil
}

// tabelaExiste verifica se a tabela do backup do mes existe e caso não
// cria uma tabela para o mes
func tabelaExiste(db *sql.DB, tabela string) error {
	sqlTxt := fmt.Sprintf(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'confmonit'
		AND TABLE_TYPE = 'BASE TABLE'
		AND table_name = '%s';
	`, tabela)

	tab, erro := db.Query(sqlTxt)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	if tab.Next() {
		return nil
	} else {
		if erro := criarTabela(db, tabela); erro != nil {
			return erro
		}
		return nil
	}
}

// criarTabela cria uma nova tabela para o mes de backup
func criarTabela(db *sql.DB, nome string) error {
	sqlTxt := fmt.Sprintf(`	
		CREATE TABLE %s (
  			ID_Evento varchar(25) NOT NULL,
  			ID_Processo varchar(25) NOT NULL,
  			Codigo varchar(4) NOT NULL DEFAULT '0000',
  			Particao varchar(2) NOT NULL DEFAULT '00',
  			ZonaUser varchar(3) NOT NULL DEFAULT '000',
  			Nivel char(1) NOT NULL DEFAULT '0',
  			Img char(1) NOT NULL DEFAULT '0',
  			DataEntrada timestamp NOT NULL DEFAULT current_timestamp()
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_general_ci;
	`, nome)

	tab, erro := db.Query(sqlTxt)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	return nil
}
