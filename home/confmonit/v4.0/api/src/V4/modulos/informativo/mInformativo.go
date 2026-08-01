package informativo

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
)

type Informativo struct {
	ID_Informativo string `json:"idInformativo"`
	ID_Vinculo     string `json:"idVinculo"`
	Menssagem      string `json:"messagem"`
	DataCadastro   string `json:"dataCadastro"`
	Tempo          string `json:"tempo"`
	Ativo          string `json:"ativo"`
}

type SInformativo struct {
	ID_Informativo sql.NullString
	ID_Vinculo     sql.NullString
	Menssagem      sql.NullString
	DataCadastro   sql.NullString
	Tempo          sql.NullString
	Ativo          sql.NullString
}

func (i *Informativo) Insere() error {
	if i.ID_Informativo == "" {
		i.ID_Informativo = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO informativos(
			ID_Informativo, 
			ID_Vinculo, 
			Menssagem, 		
			Tempo
		) VALUES ( ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		i.ID_Informativo,
		i.ID_Vinculo,
		i.Menssagem,
		i.Tempo,
	); err != nil {
		return err
	}
	return nil
}

func (i *Informativo) GetDadosById() error {
	if i.ID_Informativo == "" {
		return errors.New("um id de infomativo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			informativos.ID_Informativo, 
			informativos.ID_Vinculo, 
			informativos.Menssagem, 
			informativos.DataCadastro, 
			informativos.Tempo, 
			informativos.Ativo 
		FROM informativos 
		
		WHERE informativos.ID_Informativo = ?
	`, i.ID_Informativo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(
			&i.ID_Informativo,
			&i.ID_Vinculo,
			&i.Menssagem,
			&i.DataCadastro,
			&i.Tempo,
			&i.Ativo,
		); err != nil {
			return err
		}
		return nil
	}
	return errors.New("informativo não encontrado na base de dados")
}

func (i *Informativo) AlteraById() error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO informativos(			
			Menssagem, 		
			Tempo
		) VALUES ( ?, ?  )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		i.Menssagem,
		i.Tempo,
	); err != nil {
		return err
	}

	return nil
}

func (i *Informativo) DeletaById() error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM informativos 
		WHERE informativos.ID_Informativo = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(i.ID_Informativo); err != nil {
		return err
	}

	return nil
}

func (i *Informativo) ListaByIdVinculo(lista *[]Informativo) error {
	if i.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser infomormado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			informativos.ID_Informativo, 
			informativos.ID_Vinculo, 
			informativos.Menssagem, 
			informativos.DataCadastro, 
			informativos.Tempo, 
			informativos.Ativo 
		FROM informativos 
		
		WHERE informativos.ID_Vinculo = ?
	`, i.ID_Vinculo)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Informativo

		if err := tab.Scan(
			&item.ID_Informativo,
			&item.ID_Vinculo,
			&item.Menssagem,
			&item.DataCadastro,
			&item.Tempo,
			&item.Ativo,
		); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}
