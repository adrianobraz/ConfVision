package S009

import (
	"database/sql"
	"strconv"
)

func representante(db *sql.DB, email *tEmail) error {

	tab, err := db.Query(`
		SELECT 
			pacotes.EmailQtd,
			pacotes.EmailExedente,
			pacotes.EmailBloquear, 

			representante.ID_Representante,
			representante.RazaoSocial
		
		FROM representante 

		LEFT JOIN pacotes
		ON representante.ID_Pacote = pacotes.ID_Pacote

		WHERE representante.ID_Representante = ?
	`, email.idVinculo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var (
			bloqRep sql.NullString
			valor   sql.NullString
			qtdRep  sql.NullString
			idRep   sql.NullString
		)

		if err := tab.Scan(
			&qtdRep,
			&valor,
			&bloqRep,
			&idRep,
			&email.nomeVinculo,
		); err != nil {
			return err
		}

		if valor.Valid{
			email.valor = valor.String
		}else {
			email.valor = "0,00"
		}

		// Verifica se o uso é livre
		if qtdRep.String == "-1" {
			email.chave = "L"
		} else {
			email.chave = "S"
		}

		var usoRep int
		var usoFra int
		var usoCli int

		if bloqRep.String == "S" && qtdRep.String != "-1" {

			usoPct, err := strconv.Atoi(qtdRep.String)
			if err != nil {
				return err
			}

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

func repByRep(db *sql.DB, idRep string, qtdOut *int) error {
	tab, err := db.Query(`
			SELECT COUNT(tarifacao.ID_Tarifacao)
			FROM tarifacao
			WHERE  tarifacao.ID_Vinculo = ?
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
