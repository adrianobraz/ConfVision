package S009

import (
	"database/sql"
	"strconv"
)

func franqueado(db *sql.DB, email *tEmail) error {
	tab, err := db.Query(`
		SELECT 
			pctRep.EmailBloquear,
			pctRep.EmailQtd,

			pctFra.EmailQtd,
			pctFra.EmailExedente,
			pctFra.EmailBloquear,

			franqueado.ID_Representante,
			franqueado.RazaoSocial
		
		FROM franqueado 

		LEFT JOIN representante
		ON franqueado.ID_Representante = representante.ID_Representante		

		LEFT JOIN pacotes AS pctRep
		ON representante.ID_Pacote = pctRep.ID_Pacote		

		LEFT JOIN pacotes AS pctFra
		ON franqueado.ID_Pacote = pctFra.ID_Pacote		

		WHERE franqueado.ID_Franqueado = ?
		
	`, email.idVinculo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var (
			bloqRep sql.NullString
			qtdRep  sql.NullString
			bloqFra sql.NullString
			valor   sql.NullString
			qtdFra  sql.NullString
			idRep   sql.NullString
		)

		if err := tab.Scan(
			&bloqRep,
			&qtdRep,

			&qtdFra,
			&valor,
			&bloqFra,

			&idRep,
			&email.nomeVinculo,
		); err != nil {
			return err
		}

		if valor.Valid {
			email.valor = valor.String
		} else {
			email.valor = "0,00"
		}

		// Verifica email livre
		if qtdFra.String == "-1" {
			email.chave = "L"
		} else {
			email.chave = "S"
		}

		var usoRep int
		var usoFra int
		var usoCli int

		if bloqFra.String == "S" && qtdFra.String != "-1" {

			usoPct, err := strconv.Atoi(qtdFra.String)
			if err != nil {
				return err
			}

			if err := fraByFra(db, email.idVinculo, &usoFra); err != nil {
				return err
			}

			if usoFra >= usoPct {
				email.chave = "N"
				return nil
			}

			if err := cliByFra(db, email.idVinculo, &usoCli); err != nil {
				return err
			}

			if usoFra+usoCli >= usoPct {
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

			if usoRep >= usoPct {
				email.chave = "N"
				return nil
			}

			if err := fraByRep(db, idRep.String, &usoFra); err != nil {
				return err
			}

			if usoRep+usoFra >= usoPct {
				email.chave = "N"
				return nil
			}

			if err := cliByRep(db, idRep.String, &usoCli); err != nil {
				return err
			}

			if usoRep+usoFra+usoCli >= usoPct {
				email.chave = "N"
				return nil
			}
		}

		// se chegar até aqui email.chave = "S" ou "L"

	}

	return nil
}

func fraByRep(db *sql.DB, idRep string, qtdOut *int) error {
	tab, err := db.Query(`
			SELECT COUNT(tarifacao.ID_Tarifacao)
			FROM tarifacao
			WHERE  tarifacao.ID_Vinculo IN(
				SELECT franqueado.ID_Franqueado
				FROM franqueado
				WHERE franqueado.ID_Representante = ?
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

func fraByFra(db *sql.DB, idFra string, qtdOut *int) error {
	tab, err := db.Query(`
			SELECT COUNT(tarifacao.ID_Tarifacao)
			FROM tarifacao
			WHERE  tarifacao.ID_Vinculo = ?
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
