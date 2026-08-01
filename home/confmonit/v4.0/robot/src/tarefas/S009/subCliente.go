package S009

import (
	"database/sql"
	"strconv"
)

func cliente(db *sql.DB, email *tEmail) error {
	/*
		email.idEmail		Recebe do modulo anterior
		email.idVinculo		Recebe do modulo anterior = idCliente
		email.tipoVinculo	Recebe do modulo anterior = CLI
		email.destinatario	Recebe do modulo anterior
		email.menssagem		Recebe do modulo anterior
		email.chave			Recebe do modulo anterior = N
		email.valor			Recebe do modulo anterior = 0,00
	*/

	tab, err := db.Query(`
		SELECT 
			pctRep.EmailQtd AS repQtd,
			pctRep.EmailBloquear AS repBloq,

			pctFra.EmailQtd AS fraQtd,
			pctFra.EmailBloquear AS fraBloq,

			pctCli.EmailQtd AS cliQtd,
			pctCli.EmailExedente As cliValor,
			pctCli.EmailBloquear AS cliBloq ,

			franqueado.ID_Representante,
			franqueado.ID_Franqueado,
			cliente.Nome
		
		FROM cliente 

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado		

		LEFT JOIN  representante
		ON franqueado.ID_Representante = representante.ID_Representante		

		LEFT JOIN pacotes AS pctRep
		ON representante.ID_Pacote = pctRep.ID_Pacote	

		LEFT JOIN pacotes AS pctFra
		ON franqueado.ID_Pacote = pctFra.ID_Pacote	
		
		LEFT JOIN pacotes AS pctCli
		ON cliente.ID_Pacote = pctCli.ID_Pacote		

		WHERE cliente.ID_Cliente = ?
	`, email.idVinculo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var (
			qtdRep  sql.NullString
			bloqRep sql.NullString

			qtdFra  sql.NullString
			bloqFra sql.NullString

			qtdCli  sql.NullString
			valor   sql.NullString
			bloqCli sql.NullString

			idRep sql.NullString
			idFra sql.NullString
		)

		if err := tab.Scan(
			&qtdRep,
			&bloqRep,

			&qtdFra,
			&bloqFra,

			&qtdCli,
			&valor,
			&bloqCli,

			&idRep,
			&idFra,
			&email.nomeVinculo,
		); err != nil {
			return err
		}

		if valor.Valid {
			email.valor = valor.String
		} else {
			email.valor = "0,00"
		}

		// Verifica se o envio de email e ilimitado -1
		if qtdCli.String == "-1" {
			// Seta a chave para envio livre N -> L
			email.chave = "L"
		} else {
			// Seta a chave para envio com custo N -> S
			email.chave = "S"
		}

		var usoRep int
		var usoFra int
		var usoCli int

		// Verifica se o bloqueio de exedente do cliente esta ativo
		// e se o envio livre esta desabilitadp
		if bloqCli.String == "S" && qtdCli.String != "-1" {

			cliPct, err := strconv.Atoi(bloqCli.String)
			if err != nil {
				return err
			}

			if err := cliByCli(db, email.idVinculo, &usoCli); err != nil {
				return err
			}

			if usoCli >= cliPct {
				email.chave = "N"
				return nil
			}
		}

		// se chegar até aqui email.chave = "S" ou "L"

		if bloqFra.String == "S" && qtdFra.String != "-1" {

			usoPct, err := strconv.Atoi(qtdFra.String)
			if err != nil {
				return err
			}

			if err := fraByFra(db, idFra.String, &usoFra); err != nil {
				return err
			}

			if usoCli+usoFra >= usoPct {
				email.chave = "N"
				return nil
			}

		}

		// se chegar até aqui email.chave = "S" ou "L"

		if bloqRep.String == "S" && qtdRep.String != "-1" {

			usoPct, err := strconv.Atoi(qtdRep.String)
			if err != nil {
				return err
			}

			// busca qtdRep
			if err := repByRep(db, idRep.String, &usoRep); err != nil {
				return err
			}

			if usoCli+usoFra+usoRep >= usoPct {
				email.chave = "N"
				return nil
			}

		}

		// se chegar até aqui email.chave = "S" ou "L"
	}

	return nil
}

// busca a quantidade e email enviado pelo cliente pelo id do represenate
func cliByRep(db *sql.DB, idRep string, qtdOut *int) error {

	// busca qtdCli
	tab, err := db.Query(`
			SELECT COUNT(tarifacao.ID_Tarifacao)
			FROM tarifacao
			WHERE tarifacao.ID_Vinculo IN(
				SELECT cliente.ID_Cliente
				FROM cliente
				WHERE cliente.ID_Franqueado IN(
					SELECT franqueado.ID_Franqueado
					FROM franqueado
					WHERE franqueado.ID_Representante = ?
				)
			)
			AND tarifacao.TipoOperacao = 'EMAIL'
			
		`, idRep)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var qtd sql.NullString

		if err := tab.Scan(&qtd); err != nil {
			return err
		}

		tmp, err := strconv.Atoi(qtd.String)
		if err != nil {
			return err
		}

		*qtdOut = tmp
		return nil

	}

	*qtdOut = 0
	return nil
}

// busca a quantidade e email enviado pelo cliente pelo id franqueado
func cliByFra(db *sql.DB, idFra string, qtdOut *int) error {

	// busca qtdCli
	tab, err := db.Query(`
			SELECT COUNT(tarifacao.ID_Tarifacao)
			FROM tarifacao
			WHERE tarifacao.ID_Vinculo IN(
				SELECT cliente.ID_Cliente
				FROM cliente
				WHERE cliente.ID_Franqueado = ?				
			)
			AND tarifacao.TipoOperacao = 'EMAIL'
			
		`, idFra)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var qtd sql.NullString

		if err := tab.Scan(&qtd); err != nil {
			return err
		}

		tmp, err := strconv.Atoi(qtd.String)
		if err != nil {
			return err
		}

		*qtdOut = tmp
		return nil

	}

	*qtdOut = 0
	return nil
}

// busca a quantidade e email enviado pelo cliente pelo seu id
func cliByCli(db *sql.DB, idCli string, qtdOut *int) error {

	// busca qtdCli
	tab, err := db.Query(`
			SELECT COUNT(tarifacao.ID_Tarifacao)
			FROM tarifacao
			WHERE tarifacao.ID_Vinculo  = ?				
			AND tarifacao.TipoOperacao = 'EMAIL'
			
		`, idCli)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var qtd sql.NullString

		if err := tab.Scan(&qtd); err != nil {
			return err
		}

		tmp, err := strconv.Atoi(qtd.String)
		if err != nil {
			return err
		}

		*qtdOut = tmp
		return nil

	}

	*qtdOut = 0
	return nil
}
