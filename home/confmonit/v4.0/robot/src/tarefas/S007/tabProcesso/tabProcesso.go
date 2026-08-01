package tabProcesso

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	"robot/src/tarefas/S007/tabEvento"
	"time"
)

type processo struct {
	idProcesso     string
	idDispositivo  string
	idAtendente    string
	nivel          string
	dataCriacao    time.Time
	dataAtenInicio time.Time
	dataAtenFim    time.Time
	descricao      string
	msgAtendente   string
}

func BackupTabelaProcesso(qtdMes int) error {

	tabela := fmt.Sprintf(
		"processo_%s", time.Now().AddDate(0, -qtdMes, 0).Format("Jan_2006"),
	)

	db, erro := auxiliar.Conectar()
	if erro != nil {
		return erro
	}
	defer db.Close()

	var tIni, tFim time.Time
	if erro := auxiliar.MontaData(qtdMes, &tIni, &tFim); erro != nil {
		return erro
	}

	sqltxt := fmt.Sprintf(`
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
		WHERE processo.DataCriacao >= '%s'
		AND processo.DataCriacao <= '%s'
		LIMIT 50
	`, tIni.Format("2006-01-02 15:04:05"), tFim.Format("2006-01-02 15:04:05"))

	// Carrega todos os processos com data entre tIni e tFim
	tab, erro := db.Query(sqltxt)
	if erro != nil {
		return erro
	}
	defer tab.Close()
	// Processar os processos carregados
	for tab.Next() {
		if erro := tabelaExiste(db, tabela); erro != nil {
			return erro
		}

		var p processo

		if erro := tab.Scan(
			&p.idProcesso,
			&p.idDispositivo,
			&p.idAtendente,
			&p.nivel,
			&p.dataCriacao,
			&p.dataAtenInicio,
			&p.dataAtenFim,
			&p.descricao,
			&p.msgAtendente,
		); erro != nil {
			return erro
		}

		// Gravar o processo na tabela backup
		if erro := gravarTabBackup(db, p, tabela); erro != nil {
			return erro
		}

		// // Efetuar o backup dos eventos desse processo
		if erro := tabEvento.BackupTabelaEventos(db, qtdMes, p.idProcesso); erro != nil {
			return erro
		}

		// // Apagar o processo na tabela principal
		if erro := deletarTabPrincipal(db, p.idProcesso); erro != nil {
			return erro
		}
	}

	return nil
}

func gravarTabBackup(db *sql.DB, p processo, tabela string) error {
	fmt.Printf("Gravando processo: %s\n", p.idProcesso)
	sqlTxt := fmt.Sprintf(`
		INSERT INTO %s(
			ID_Processo, 
			ID_Dispositivo, 
			ID_Atendente, 
			Nivel, 
			DataCriacao, 
			DataAtenInicio, 
			DataAtenFim, 
			Descricao, 
			MsgAtendente
		) VALUES (?,?,?,?,?,?,?,?,?)
	`, tabela)

	stm, erro := db.Prepare(sqlTxt)
	if erro != nil {
		return erro
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		p.idProcesso,
		p.idDispositivo,
		p.idAtendente,
		p.nivel,
		p.dataCriacao,
		p.dataAtenInicio,
		p.dataAtenFim,
		p.descricao,
		p.msgAtendente,
	); erro != nil {
		return erro
	}
	return nil
}

func deletarTabPrincipal(db *sql.DB, idProcesso string) error {
	stm, erro := db.Prepare(`DELETE FROM processo WHERE processo.ID_Processo = ?`)
	if erro != nil {
		return erro
	}
	defer stm.Close()

	if _, erro := stm.Exec(idProcesso); erro != nil {
		return erro
	}
	return nil
}

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
func criarTabela(db *sql.DB, tabela string) error {

	sqlTxt := fmt.Sprintf(`	
		CREATE TABLE %s (
			ID_Processo varchar(25) NOT NULL,
			ID_Dispositivo varchar(25) NOT NULL DEFAULT '0',
			ID_Atendente varchar(25) NOT NULL DEFAULT '0',
			Nivel char(1) NOT NULL DEFAULT '0',
			DataCriacao timestamp NOT NULL DEFAULT current_timestamp(),
			DataAtenInicio timestamp NULL DEFAULT NULL,
			DataAtenFim timestamp NULL DEFAULT NULL,
			Descricao longtext NOT NULL DEFAULT '',
			MsgAtendente text NOT NULL DEFAULT '',

			PRIMARY KEY (ID_Processo)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_general_ci;
	`, tabela)

	tab, erro := db.Query(sqlTxt)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	return nil
}
