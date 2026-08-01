package fatNucleo

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	g004 "robot/src/tarefas/S004/global"
	"strconv"
)

type nucleo struct {
	g004.Dados
}

const GRAVA_ID_FATURA bool = true

func Fechar(idNuc string) error {
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	var nuc nucleo
	nuc.FaturaId = auxiliar.GeradorDeId()
	nuc.NucId.String = idNuc

	if err := nuc.buscaDadosPacote(db); err != nil {
		return err
	}
	if !nuc.NucId.Valid || nuc.NucId.String == "" {
		return fmt.Errorf("nucleo sem dados de pacote: %s", idNuc)
	}

	fmt.Println("Fechando ->", nuc.NucNome.String)
	if err := g004.CriaFatura(db, nuc.FaturaId, "CENTRAL", nuc.NucId.String); err != nil {
		return err
	}

	if err := g004.InserePacote(nuc.PacValor.String, nuc.PacNome.String, &nuc.Itens); err != nil {
		return err
	}

	if err := nuc.buscarCustoNucleo(db); err != nil {
		return err
	}

	var listaFra []string
	if err := nuc.buscarFranqueados(db, idNuc, &listaFra); err != nil {
		return err
	}

	if err := nuc.carregaCusto(db, listaFra); err != nil {
		return err
	}

	var listaCli []string
	if err := g004.BuscarClientes(db, listaFra, &listaCli); err != nil {
		return err
	}

	if err := nuc.carregaCusto(db, listaCli); err != nil {
		return err
	}

	if nuc.PacGrade.String == "S" {
		fmt.Println("    grade ->", nuc.NucNome.String)
		if err := g004.CarregaGrade(db, listaCli, nuc.PacGradeValor.String, &nuc.Itens); err != nil {
			return err
		}
	}

	if err := nuc.calculaExedentes(); err != nil {
		return err
	}

	if err := g004.AdicionaItem(db, nuc.FaturaId, nuc.Itens); err != nil {
		return err
	}

	fmt.Println(nuc.NucNome.String, nuc.QtdAtendimento, nuc.QtdLigacao, nuc.QtdEmail, nuc.QtdSms)
	return nil
}

func (fn *nucleo) buscarFranqueados(db *sql.DB, idNuc string, ids *[]string) error {
	tab, err := db.Query(`
		SELECT franqueado.ID_Franqueado, franqueado.RazaoSocial
		FROM franqueado 
		WHERE franqueado.ID_Representante = ?
		AND franqueado.Ativo = 'S'
		AND (
			franqueado.DataCancelamento IS NULL
			OR franqueado.DataCancelamento > DATE_SUB(NOW(), INTERVAL 30 DAY)
		)
	`, idNuc)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var id sql.NullString
		var razao sql.NullString

		if err := tab.Scan(&id, &razao); err != nil {
			return err
		}
		fmt.Println("  franqueado ->", razao.String)
		*ids = append(*ids, id.String)
	}
	return nil
}

// Busca os dados de pacote do nucleo para aplicar no franqueado
func (fn *nucleo) buscaDadosPacote(db *sql.DB) error {
	tab, err := db.Query(`
		SELECT
			representante.ID_Representante,			
			representante.RazaoSocial,

			pacotes.ID_Pacote, 
			pacotes.Nome, 
			pacotes.Valor, 

			pacotes.Grade, 
			pacotes.GradeValor, 

			pacotes.AtendimentoQtd, 
			pacotes.AtendimentoExedente,

			pacotes.EmailQtd, 
			pacotes.EmailExedente, 
						
			pacotes.SmsQtd, 
			pacotes.SmsExedente, 			 

			pacotes.LigacoesQtd, 
			pacotes.LigacoesExedente	

		FROM representante		
				
		LEFT JOIN pacotes
		ON representante.ID_Pacote = pacotes.ID_Pacote

		WHERE representante.ID_Representante= ? 
	`, fn.NucId.String)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(
			&fn.NucId,
			&fn.NucNome,
			&fn.PacId,
			&fn.PacNome,
			&fn.PacValor,
			&fn.PacGrade,
			&fn.PacGradeValor,
			&fn.PacAteQtd,
			&fn.PacAteExedente,
			&fn.PacEmaQtd,
			&fn.PacEmaExedente,
			&fn.PacSmsQtd,
			&fn.PacSmsExedente,
			&fn.PacLigQtd,
			&fn.PacLigExedente,
		); err != nil {
			return err
		}
	}
	return nil
}

// Busca os custos na tabela tarifação para o nucleo
func (fn *nucleo) buscarCustoNucleo(db *sql.DB) error {
	tab, err := db.Query(`
		SELECT 
			tarifacao.ID_Tarifacao, 
			tarifacao.TipoOperacao, 
			tarifacao.DadoOperacao, 
			tarifacao.Credito, 
			tarifacao.Debito, 
			tarifacao.DataOperacao 
		FROM tarifacao 
		WHERE tarifacao.ID_Vinculo = ?
		AND tarifacao.ID_FaturaNuc = '0'
	`, fn.NucId.String)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var (
			ID_Tarifacao string
			TipoOperacao string
			DadoOperacao string
			Credito      string
			Debito       string
			DataOperacao string
		)

		if err := tab.Scan(
			&ID_Tarifacao,
			&TipoOperacao,
			&DadoOperacao,
			&Credito,
			&Debito,
			&DataOperacao,
		); err != nil {
			return err
		}

		switch TipoOperacao {
		case "ATENDIMENTO":
			fn.QtdAtendimento++
		case "LIGAÇÃO":
			fn.QtdLigacao++
		case "EMAIL":
			fn.QtdEmail++
		case "SMS":
			fn.QtdSms++
		default:
			if err := g004.InsereAdicionais(DadoOperacao, Credito, Debito, &fn.Itens); err != nil {
				return err
			}
		}

		if GRAVA_ID_FATURA {
			if err := g004.GravaIdFatura(db, fn.FaturaId, ID_Tarifacao); err != nil {
				return err
			}
		}
	}
	return nil
}

// Busca os custos na tabela tarifação para o nucleo
func (fn *nucleo) carregaCusto(db *sql.DB, lista []string) error {
	for _, l := range lista {

		tab, err := db.Query(`
			SELECT 
				tarifacao.ID_Tarifacao, 
				tarifacao.TipoOperacao, 
				tarifacao.DadoOperacao, 
				tarifacao.Credito, 
				tarifacao.Debito, 
				tarifacao.DataOperacao 
			FROM tarifacao 
			WHERE tarifacao.ID_Vinculo = ?
			AND tarifacao.ID_FaturaNuc = '0'
		`, l)
		if err != nil {
			return err
		}
		defer tab.Close()

		for tab.Next() {
			var (
				ID_Tarifacao string
				TipoOperacao string
				DadoOperacao string
				Credito      string
				Debito       string
				DataOperacao string
			)

			if err := tab.Scan(
				&ID_Tarifacao,
				&TipoOperacao,
				&DadoOperacao,
				&Credito,
				&Debito,
				&DataOperacao,
			); err != nil {
				return err
			}

			switch TipoOperacao {
			case "ATENDIMENTO":
				fn.QtdAtendimento++
			case "LIGAÇÃO":
				fn.QtdLigacao++
			case "EMAIL":
				fn.QtdEmail++
			case "SMS":
				fn.QtdSms++
			}

			if GRAVA_ID_FATURA {
				if err := g004.GravaIdFatura(db, fn.FaturaId, ID_Tarifacao); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Calcula os sms exedentes
func (fn *nucleo) calculaExedentes() error {
	// Atendimento
	if fn.PacAteQtd.String != "-1" {

		pacQtdAte, err := strconv.Atoi(fn.PacAteQtd.String)
		if err != nil {
			return err
		}
		exedente := fn.QtdAtendimento - pacQtdAte
		//fmt.Println("Atendimeto", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fn.PacAteExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "SMS Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fn.Itens = append(fn.Itens, i)
		}
	}

	// Ligaçao
	if fn.PacLigQtd.String != "-1" {
		pacQtdLig, err := strconv.Atoi(fn.PacLigQtd.String)
		if err != nil {
			return err
		}
		exedente := fn.QtdLigacao - pacQtdLig
		//fmt.Println("Ligação", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fn.PacLigExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "Ligação Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fn.Itens = append(fn.Itens, i)
		}
	}

	// Email
	if fn.PacEmaQtd.String != "-1" {
		pacQtdEma, err := strconv.Atoi(fn.PacEmaQtd.String)
		if err != nil {
			return err
		}
		exedente := fn.QtdEmail - pacQtdEma
		//fmt.Println("Email", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fn.PacEmaExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "Email Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fn.Itens = append(fn.Itens, i)
		}
	}

	// Sms
	if fn.PacSmsQtd.String != "-1" {
		pacQtdSms, err := strconv.Atoi(fn.PacSmsQtd.String)
		if err != nil {
			return err
		}

		exedente := fn.QtdSms - pacQtdSms
		//fmt.Println("Sms", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fn.PacSmsExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "SMS Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fn.Itens = append(fn.Itens, i)
		}
	}

	return nil
}
