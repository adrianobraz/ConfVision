package faturaV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type FaturaItem struct {
	ID_FaturaItem string `json:"idFaturaItem"`
	ID_Fatura     string `json:"idFatura"`
	Quantidade    string `json:"quantidade"`
	Descricao     string `json:"descricao"`
	Credito       string `json:"credito"`
	Debito        string `json:"debito"`
	DataCadastro  string `json:"dataCadastro"`
}

type SFaturaItem struct {
	ID_FaturaItem sql.NullString
	ID_Fatura     sql.NullString
	Quantidade    sql.NullString
	Descricao     sql.NullString
	Credito       sql.NullString
	Debito        sql.NullString
	DataCadastro  sql.NullTime
}

func (fi *FaturaItem) Insere() error {
	if fi.ID_FaturaItem == "" {
		fi.ID_FaturaItem = auxiliar.GeradorDeId()
	}

	if fi.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO faturasItem(
			ID_FaturaItem, 
			ID_Fatura, 
			Quantidade, 
			Descricao, 
			Credito, 
			Debito
		) VALUES ( ?, ?, ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		fi.ID_FaturaItem,
		fi.ID_Fatura,
		fi.Quantidade,
		fi.Descricao,
		fi.Credito,
		fi.Debito,
	); err != nil {
		return err
	}

	return nil
}

func (fi *FaturaItem) GetDadosById() error {
	if fi.ID_FaturaItem == "" {
		fi.ID_FaturaItem = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturasItem.ID_FaturaItem = '%s'
	`, fi.ID_FaturaItem)

	tab, err := db.Query(getSelectItens(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItemItens(tab, fi); err != nil {
			return err
		}

		return nil
	}

	return errors.New("item de fatura não encontrado na base de dados")
}

func (fi *FaturaItem) AlteraById() error {
	if fi.ID_FaturaItem == "" {
		return errors.New("um id de item fatura deve ser informado")
	}

	if fi.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE faturasItem SET 
			faturasItem.Quantidade = ?,
			faturasItem.Descricao = ?,
			faturasItem.Credito = ?,
			faturasItem.Debito = ?,
		WHERE faturasItem.ID_FaturaItem = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		fi.Quantidade,
		fi.Descricao,
		fi.Credito,
		fi.Debito,
		fi.ID_FaturaItem,
	); err != nil {
		return err
	}

	return nil
}

func (fi *FaturaItem) DeleteById() error {
	if fi.ID_FaturaItem == "" {
		return errors.New("um id de item fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM faturasItem 
		WHERE faturasItem.ID_FaturaItem = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(fi.ID_FaturaItem); err != nil {
		return err
	}

	return nil
}

func (fi *FaturaItem) DeleteAllByIdFatura() error {
	if fi.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM faturasItem 
		WHERE faturasItem.ID_Fatura = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(fi.ID_Fatura); err != nil {
		return err
	}

	return nil
}

func (fi *FaturaItem) ListaByIdFatura(lista *[]FaturaItem) error {
	if fi.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturasItem.ID_Fatura = '%s'
	`, fi.ID_Fatura)

	tab, err := db.Query(getSelectItens(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item FaturaItem

		if err := processaItemItens(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// funcoes internas ===========================================================
func SFaturaItemToFaturaItem(sfi SFaturaItem) (fi FaturaItem) {

	fi.ID_FaturaItem = sfi.ID_FaturaItem.String
	fi.ID_Fatura = sfi.ID_Fatura.String
	fi.Quantidade = sfi.Quantidade.String
	fi.Descricao = sfi.Descricao.String
	fi.Credito = sfi.Credito.String
	fi.Debito = sfi.Debito.String

	if sfi.DataCadastro.Valid {
		fi.DataCadastro = sfi.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		fi.DataCadastro = ""
	}

	return
}

func getSelectItens(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			faturasItem.ID_FaturaItem, 
			faturasItem.ID_Fatura, 
			faturasItem.Quantidade, 
			faturasItem.Descricao, 
			faturasItem.Credito, 
			faturasItem.Debito,
			faturasItem.DataCadastro 
		FROM faturasItem 
		%s
	
	`, filtro)
}

func processaItemItens(tab *sql.Rows, fi *FaturaItem) error {

	var tmp SFaturaItem

	if err := tab.Scan(
		&tmp.ID_FaturaItem,
		&tmp.ID_Fatura,
		&tmp.Quantidade,
		&tmp.Descricao,
		&tmp.Credito,
		&tmp.Debito,
		&tmp.DataCadastro,
	); err != nil {
		return err
	}

	*fi = SFaturaItemToFaturaItem(tmp)

	return nil
}
