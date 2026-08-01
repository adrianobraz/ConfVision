package fatCliente

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	g004 "robot/src/tarefas/S004/global"
	"strconv"
)

type cliente struct {
	g004.Dados
}

const GRAVA_ID_FATURA bool = true

func Fechar(idCli string) error {
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	var cli cliente
	cli.FaturaId = auxiliar.GeradorDeId()
	cli.CliId.String = idCli

	if err := cli.buscaDadosPacote(db); err != nil {
		return err
	}
	if !cli.CliId.Valid || cli.CliId.String == "" || !cli.FraId.Valid || cli.FraId.String == "" {
		return fmt.Errorf("cliente sem dados de pacote: %s", idCli)
	}

	fmt.Println("Fechando ->", cli.CliNome.String)
	if err := g004.CriaFatura(db, cli.FaturaId, cli.FraId.String, cli.CliId.String); err != nil {
		return err
	}

	if err := g004.InserePacote(cli.PacValor.String, cli.PacNome.String, &cli.Itens); err != nil {
		return err
	}

	if err := cli.buscarCustoCliente(db); err != nil {
		return err
	}

	listaCli := []string{cli.CliId.String}
	if cli.PacGrade.String == "S" {
		fmt.Println("    grade ->", cli.CliNome.String)
		if err := g004.CarregaGrade(db, listaCli, cli.PacGradeValor.String, &cli.Itens); err != nil {
			return err
		}
	}

	if err := cli.calculaExedentes(); err != nil {
		return err
	}

	if err := g004.AdicionaItem(db, cli.FaturaId, cli.Itens); err != nil {
		return err
	}

	return nil
}

// Busca os dados de pacote do nucleo para aplicar no franqueado
func (fc *cliente) buscaDadosPacote(db *sql.DB) error {
	tab, err := db.Query(`
		SELECT
			cliente.ID_Franqueado, 
			
			cliente.ID_Cliente, 
			cliente.Nome,

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

		FROM cliente		
				
		LEFT JOIN pacotes
		ON cliente.ID_Pacote = pacotes.ID_Pacote

		WHERE cliente.ID_Cliente  = ? 
	`, fc.CliId.String)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(
			&fc.FraId,
			&fc.CliId,
			&fc.CliNome,
			&fc.PacId,
			&fc.PacNome,
			&fc.PacValor,
			&fc.PacGrade,
			&fc.PacGradeValor,
			&fc.PacAteQtd,
			&fc.PacAteExedente,
			&fc.PacEmaQtd,
			&fc.PacEmaExedente,
			&fc.PacSmsQtd,
			&fc.PacSmsExedente,
			&fc.PacLigQtd,
			&fc.PacLigExedente,
		); err != nil {
			return err
		}
	}
	return nil
}

// Busca os custos na tabela tarifação para o nucleo
func (fc *cliente) buscarCustoCliente(db *sql.DB) error {
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
		AND tarifacao.ID_Fatura = '0'
	`, fc.CliId.String)
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
			fc.QtdAtendimento++
		case "LIGAÇÃO":
			fc.QtdLigacao++
		case "EMAIL":
			fc.QtdEmail++
		case "SMS":
			fc.QtdSms++
		default:
			if err := g004.InsereAdicionais(DadoOperacao, Credito, Debito, &fc.Itens); err != nil {
				return err
			}
		}

		if GRAVA_ID_FATURA {
			if err := g004.GravaIdFatura(db, fc.FaturaId, ID_Tarifacao); err != nil {
				return err
			}
		}
	}
	return nil
}

// Calcula os sms exedentes
func (fc *cliente) calculaExedentes() error {
	// Atendimento
	if fc.PacAteQtd.String != "-1" {

		pacQtdAte, err := strconv.Atoi(fc.PacAteQtd.String)
		if err != nil {
			return err
		}
		exedente := fc.QtdAtendimento - pacQtdAte
		//fmt.Println("Atendimeto", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fc.PacAteExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "SMS Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fc.Itens = append(fc.Itens, i)
		}
	}

	// Ligaçao
	if fc.PacLigQtd.String != "-1" {
		pacQtdLig, err := strconv.Atoi(fc.PacLigQtd.String)
		if err != nil {
			return err
		}
		exedente := fc.QtdLigacao - pacQtdLig
		//fmt.Println("Ligação", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fc.PacLigExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "Ligação Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fc.Itens = append(fc.Itens, i)
		}
	}

	// Email
	if fc.PacEmaQtd.String != "-1" {
		pacQtdEma, err := strconv.Atoi(fc.PacEmaQtd.String)
		if err != nil {
			return err
		}
		exedente := fc.QtdEmail - pacQtdEma
		//fmt.Println("Email", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fc.PacEmaExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "Email Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fc.Itens = append(fc.Itens, i)
		}
	}

	// Sms
	if fc.PacSmsQtd.String != "-1" {
		pacQtdSms, err := strconv.Atoi(fc.PacSmsQtd.String)
		if err != nil {
			return err
		}

		exedente := fc.QtdSms - pacQtdSms
		//fmt.Println("Sms", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(fc.PacSmsExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "SMS Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			fc.Itens = append(fc.Itens, i)
		}
	}

	return nil
}
