package listaInformativo

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type ListaInformativo struct {
	ID_Lista         string `json:"idLista"`
	ID_Alvo          string `json:"idAlvo"`
	ID_Informativo   string `json:"idInformativo"`
	DataCadastro     string `json:"dataCadastro"`
	DataEncerramento string `json:"dataEncerramento"`
	Msg              string `json:"msg"`
	Tempo            int    `json:"tempo"`
}

type SListaInformativo struct {
	ID_Lista         sql.NullString
	ID_Alvo          sql.NullString
	ID_Informativo   sql.NullString
	DataCadastro     sql.NullTime
	DataEncerramento sql.NullTime
	Msg              sql.NullString
	Tempo            sql.NullString
}

func (li *ListaInformativo) Insere() error {

	if li.ID_Lista == "" {
		li.ID_Lista = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO listaInformativo(
			listaInformativo.ID_Lista, 
			listaInformativo.ID_Alvo, 
			listaInformativo.ID_Informativo,
			listaInformativo.DataEncerramento
		) VALUES ( ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if li.DataEncerramento == "" {
		// 120 =  5 dias
		li.DataEncerramento = time.Now().Add(120 * time.Hour).Format("2006-01-02 15:04:05")
	}

	if _, err := stm.Exec(
		li.ID_Lista,
		li.ID_Alvo,
		li.ID_Informativo,
		li.DataEncerramento,
	); err != nil {
		return err
	}

	return nil
}

func (li *ListaInformativo) GetDadosById() error {
	if li.ID_Lista == "" {
		return errors.New("um id de lista deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			listaInformativo.ID_Lista, 
			listaInformativo.ID_Alvo, 
			listaInformativo.ID_Informativo, 
			listaInformativo.DataCadastro, 
			listaInformativo.DataEncerramento 
		FROM listaInformativo 
		WHERE listaInformativo.ID_Lista = ?
	`, li.ID_Lista)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(
			&li.ID_Lista,
			&li.ID_Alvo,
			&li.ID_Informativo,
			&li.DataCadastro,
			&li.DataEncerramento,
		); err != nil {
			return nil
		}
		return nil

	}
	return errors.New("lista não encontrado na base de dados")
}

func (li *ListaInformativo) DeletaById() error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM listaInformativo WHERE listaInformativo.ID_Lista = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(li.ID_Lista); err != nil {
		return err
	}

	return nil
}

func (li *ListaInformativo) ListaByIdAlvo(lista *[]ListaInformativo) error {
	if li.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser infomormado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			listaInformativo.ID_Lista, 
			listaInformativo.ID_Alvo, 
			listaInformativo.ID_Informativo, 
			listaInformativo.DataCadastro, 
			listaInformativo.DataEncerramento,

			informativos.Menssagem,			
			informativos.Tempo			

		FROM listaInformativo 

		LEFT JOIN informativos
		ON listaInformativo.ID_Informativo =  informativos.ID_Informativo
		
		WHERE listaInformativo.ID_Alvo = ?
		AND listaInformativo.DataEncerramento > ?
		AND informativos.Ativo = 'S'
	`, li.ID_Alvo, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item SListaInformativo
		if err := tab.Scan(
			&item.ID_Lista,
			&item.ID_Alvo,
			&item.ID_Informativo,
			&item.DataCadastro,
			&item.DataEncerramento,
			&item.Msg,
			&item.Tempo,
		); err != nil {
			return err
		}

		*lista = append(*lista, sListaInformativoToListaInformativo(item))
	}

	return nil
}

func sListaInformativoToListaInformativo(sli SListaInformativo) (li ListaInformativo) {
	li.ID_Lista = sli.ID_Lista.String
	li.ID_Alvo = sli.ID_Alvo.String
	li.ID_Informativo = sli.ID_Informativo.String
	li.DataCadastro = sli.DataCadastro.Time.Format("02/01/2006 15:04:05")
	li.DataEncerramento = sli.DataEncerramento.Time.Format("02/01/2006 15:04:05")
	li.Msg = sli.Msg.String
	li.Tempo, _ = strconv.Atoi(sli.Tempo.String)
	return
}
