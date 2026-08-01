package fatFranqueado

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	g004 "robot/src/tarefas/S004/global"
	"strconv"
)

type franqueado struct {
	g004.Dados
}

const GRAVA_ID_FATURA bool = true

func Fechar(idFra string) error {
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	var fra franqueado
	fra.FaturaId = auxiliar.GeradorDeId()
	fra.FraId.String = idFra

	if err := fra.buscaDadosPacote(db); err != nil {
		return err
	}
	if !fra.FraId.Valid || fra.FraId.String == "" || !fra.NucId.Valid || fra.NucId.String == "" {
		return fmt.Errorf("franqueado sem dados de pacote: %s", idFra)
	}

	fmt.Println("Fechando ->", fra.FraNome.String)
	if err := g004.CriaFatura(db, fra.FaturaId, fra.NucId.String, fra.FraId.String); err != nil {
		return err
	}

	if err := g004.InserePacote(fra.PacValor.String, fra.PacNome.String, &fra.Itens); err != nil {
		return err
	}

	if err := fra.buscarCustoFranqueado(db); err != nil {
		return err
	}
	fmt.Println(fra.FraNome.String, fra.QtdAtendimento, fra.QtdLigacao, fra.QtdEmail, fra.QtdSms)

	listaFra := []string{fra.FraId.String}
	var listaCli []string
	if err := g004.BuscarClientes(db, listaFra, &listaCli); err != nil {
		return err
	}

	if err := fra.carregaCusto(db, listaCli); err != nil {
		return err
	}

	if fra.PacGrade.String == "S" {
		fmt.Println("    grade ->", fra.FraNome.String)
		if err := g004.CarregaGrade(db, listaCli, fra.PacGradeValor.String, &fra.Itens); err != nil {
			return err
		}
	}

	if err := fra.calculaExedentes(); err != nil {
		return err
	}

	if err := g004.AdicionaItem(db, fra.FaturaId, fra.Itens); err != nil {
		return err
	}

	fmt.Println(fra.FraNome.String, fra.QtdAtendimento, fra.QtdLigacao, fra.QtdEmail, fra.QtdSms)
	return nil
}

// Busca os dados de pacote do nucleo para aplicar no franqueado
func (ff *franqueado) buscaDadosPacote(db *sql.DB) error {
	tab, err := db.Query(`
		SELECT
			representante.ID_Representante,					

			franqueado.ID_Franqueado, 
			franqueado.RazaoSocial,

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

		FROM franqueado		
				
		LEFT JOIN pacotes
		ON franqueado.ID_Pacote = pacotes.ID_Pacote

		WHERE franqueado.ID_Franqueado = ? 
	`, ff.FraId.String)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(
			&ff.NucId,
			&ff.FraId,
			&ff.FraNome,
			&ff.PacId,
			&ff.PacNome,
			&ff.PacValor,
			&ff.PacGrade,
			&ff.PacGradeValor,
			&ff.PacAteQtd,
			&ff.PacAteExedente,
			&ff.PacEmaQtd,
			&ff.PacEmaExedente,
			&ff.PacSmsQtd,
			&ff.PacSmsExedente,
			&ff.PacLigQtd,
			&ff.PacLigExedente,
		); err != nil {
			return err
		}
	}
	return nil
}

// Busca os custos na tabela tarifação para o nucleo
func (ff *franqueado) buscarCustoFranqueado(db *sql.DB) error {
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
		AND tarifacao.ID_FaturaFra = '0'
	`, ff.FraId.String)
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
			ff.QtdAtendimento++
		case "LIGAÇÃO":
			ff.QtdLigacao++
		case "EMAIL":
			ff.QtdEmail++
		case "SMS":
			ff.QtdSms++
		default:
			if err := g004.InsereAdicionais(DadoOperacao, Credito, Debito, &ff.Itens); err != nil {
				return err
			}
		}

		if GRAVA_ID_FATURA {
			if err := g004.GravaIdFatura(db, ff.FaturaId, ID_Tarifacao); err != nil {
				return err
			}
		}
	}
	return nil
}

// Busca os custos na tabela tarifação para o nucleo
func (ff *franqueado) carregaCusto(db *sql.DB, lista []string) error {
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
			AND tarifacao.ID_FaturaFra = '0'
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
				ff.QtdAtendimento++
			case "LIGAÇÃO":
				ff.QtdLigacao++
			case "EMAIL":
				ff.QtdEmail++
			case "SMS":
				ff.QtdSms++
			}

			if GRAVA_ID_FATURA {
				if err := g004.GravaIdFatura(db, ff.FaturaId, ID_Tarifacao); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Calcula os sms exedentes
func (ff *franqueado) calculaExedentes() error {
	// Atendimento
	if ff.PacAteQtd.String != "-1" {

		pacQtdAte, err := strconv.Atoi(ff.PacAteQtd.String)
		if err != nil {
			return err
		}
		exedente := ff.QtdAtendimento - pacQtdAte
		//fmt.Println("Atendimeto", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(ff.PacAteExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "SMS Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			ff.Itens = append(ff.Itens, i)
		}
	}

	// Ligaçao
	if ff.PacLigQtd.String != "-1" {
		pacQtdLig, err := strconv.Atoi(ff.PacLigQtd.String)
		if err != nil {
			return err
		}
		exedente := ff.QtdLigacao - pacQtdLig
		//fmt.Println("Ligação", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(ff.PacLigExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "Ligação Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			ff.Itens = append(ff.Itens, i)
		}
	}

	// Email
	if ff.PacEmaQtd.String != "-1" {
		pacQtdEma, err := strconv.Atoi(ff.PacEmaQtd.String)
		if err != nil {
			return err
		}
		exedente := ff.QtdEmail - pacQtdEma
		//fmt.Println("Email", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(ff.PacEmaExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "Email Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			ff.Itens = append(ff.Itens, i)
		}
	}

	// Sms
	if ff.PacSmsQtd.String != "-1" {
		pacQtdSms, err := strconv.Atoi(ff.PacSmsQtd.String)
		if err != nil {
			return err
		}

		exedente := ff.QtdSms - pacQtdSms
		//fmt.Println("Sms", exedente)
		if exedente > 0 {
			var valor float64
			if err := g004.StrigToFloat(ff.PacSmsExedente.String, &valor); err != nil {
				return err
			}

			// Cria o objeto item e carrega seu conteudo
			var i g004.Item
			i.Quantidade = strconv.Itoa(exedente)
			i.Descricao = "SMS Execedentes"
			i.Credito = 0.00
			i.Debito = float64(exedente) * valor

			// Adiciona o item na lista de itens da fatura
			ff.Itens = append(ff.Itens, i)
		}
	}

	return nil
}
