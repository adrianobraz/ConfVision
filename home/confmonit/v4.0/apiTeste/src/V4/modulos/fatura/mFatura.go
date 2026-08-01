package faturaV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Fatura struct {
	ID_Fatura        string       `json:"idFatura"`
	ID_Origem        string       `json:"idOrigem"`
	ID_Destino       string       `json:"idDestino"`
	Status           string       `json:"status"`
	Valor            string       `json:"valor"`
	DataCadastro     string       `json:"dataCadastro"`
	DataUltimoStatus string       `json:"dataUltimoStatus"`
	DataVencimento   string       `json:"dataVencimento"`
	NomeDestino      string       `json:"nomeDestino"`
	Itens            []FaturaItem `json:"itens"`
}

type SFatura struct {
	ID_Fatura        sql.NullString
	ID_Origem        sql.NullString
	ID_Destino       sql.NullString
	Status           sql.NullString
	Valor            sql.NullString
	DataCadastro     sql.NullTime
	DataUltimoStatus sql.NullTime
	DataVencimento   sql.NullTime
	NomeRep          sql.NullString
	NomeFra          sql.NullString
	NomeCli          sql.NullString
	Itens            SFaturaItem
}

func (f *Fatura) Insere() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	if f.ID_Origem == "" {
		return errors.New("um id de origem deve ser informado")
	}

	if f.ID_Destino == "" {
		return errors.New("um id de destino deve ser informado")
	}

	if f.Status == "" {
		f.Status = "ABERTO"
	}

	if f.DataVencimento == "" {
		f.DataVencimento = time.Now().Add(120 * time.Hour).Format("2006-01-02")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO faturas(
			ID_Fatura, 
			ID_Origem, 
			ID_Destino, 
			Status, 
			Valor, 
			DataVencimento
		) VALUES ( ?, ?, ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		f.ID_Fatura,
		f.ID_Origem,
		f.ID_Destino,
		f.Status,
		f.Valor,
		f.DataVencimento,
	); err != nil {
		return err
	}
	return nil
}

func (f *Fatura) GetDadosById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturas.ID_Fatura = '%s'
	`, f.ID_Fatura)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, f); err != nil {
			return err
		}

		return nil
	}
	return errors.New("fatura não encontrada na base de dados")
}

// funcoes para manipular o campo status ======================================
func (f *Fatura) GetStatusById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT faturas.Status
		FROM faturas 
		WHERE faturas.ID_Fatura = ?
	`, f.ID_Fatura)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&f.Status); err != nil {
			return err
		}
		return nil
	}
	return errors.New("fatura não encontrada na base de dados")
}

func (f *Fatura) SetStatusById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	if f.Status == "" {
		return errors.New("um status de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE faturas 
		SET faturas.Status = ? 
		WHERE faturas.ID_Fatura = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.Status, f.ID_Fatura); err != nil {
		return err
	}

	return nil
}

// funcoes para manipular o campo valor =======================================
func (f *Fatura) GetValorById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT faturas.Valor 
		FROM faturas 
		WHERE faturas.ID_Fatura = ? 
	`, f.ID_Fatura)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&f.Valor); err != nil {
			return err
		}
		return nil
	}

	return errors.New("fatura não encontrada na base de dados")
}

func (f *Fatura) SetValorById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	if f.Valor == "" {
		return errors.New("um valor de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE faturas 
		SET faturas.Valor = ? 
		WHERE faturas.ID_Fatura = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.Valor, f.ID_Fatura); err != nil {
		return err
	}

	return nil
}

func (f *Fatura) AddValorById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	if f.Valor == "" {
		return errors.New("um valor de fatura deve ser informado")
	}

	var valorIn float64
	if err := auxiliar.StrigToFloat(f.Valor, &valorIn); err != nil {
		return err
	}

	if err := f.GetValorById(); err != nil {
		return err
	}

	var valorBd float64
	if err := auxiliar.StrigToFloat(f.Valor, &valorBd); err != nil {
		return err
	}

	auxiliar.FloatToString(valorBd+valorIn, &f.Valor)

	if err := f.SetValorById(); err != nil {
		return err
	}

	return nil
}

func (f *Fatura) SubValorById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	if f.Valor == "" {
		return errors.New("um valor de fatura deve ser informado")
	}

	var valorIn float64
	if err := auxiliar.StrigToFloat(f.Valor, &valorIn); err != nil {
		return err
	}

	if err := f.GetValorById(); err != nil {
		return err
	}

	var valorBd float64
	if err := auxiliar.StrigToFloat(f.Valor, &valorBd); err != nil {
		return err
	}

	auxiliar.FloatToString(valorBd-valorIn, &f.Valor)

	if err := f.SetValorById(); err != nil {
		return err
	}

	return nil
}

func (f *Fatura) DeleteById() error {
	if f.ID_Fatura == "" {
		return errors.New("um id de fatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM faturas WHERE faturas.ID_Fatura = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Fatura); err != nil {
		return err
	}

	return nil
}

func (f *Fatura) DeleteAllByIdOrigem() error {
	if f.ID_Origem == "" {
		return errors.New("um id de origem deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM faturas WHERE faturas.ID_Origem = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Origem); err != nil {
		return err
	}
	return nil
}

func (f *Fatura) DeleteAllByIdDestino() error {
	if f.ID_Destino == "" {
		return errors.New("um id de destino deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM faturas WHERE faturas.ID_Destino = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Destino); err != nil {
		return err
	}
	return nil
}

func (f *Fatura) ListaByIdOrigem(lista *[]Fatura) error {
	if f.ID_Origem == "" {
		return errors.New("um id de origem deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturas.Status <> 'PAGO'
		AND	faturas.ID_Origem = '%s'
	`, f.ID_Origem)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Fatura
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

func (f *Fatura) ListaPagoByIdOrigem(lista *[]Fatura) error {
	if f.ID_Origem == "" {
		return errors.New("um id de origem deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturas.Status = 'PAGO'
		AND	faturas.ID_Origem = '%s'
	`, f.ID_Origem)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Fatura
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

func (f *Fatura) ListaByIdDestino(lista *[]Fatura) error {
	if f.ID_Destino == "" {
		return errors.New("um id de destino deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturas.Status <> 'PAGO'
		AND	faturas.ID_Destino = '%s'
	`, f.ID_Fatura)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		var item Fatura

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

func (f *Fatura) ListaPagoByIdDestino(lista *[]Fatura) error {
	if f.ID_Destino == "" {
		return errors.New("um id de destino deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE faturas.Status = 'PAGO'
		AND	faturas.ID_Destino = '%s'
	`, f.ID_Fatura)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		var item Fatura

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

// funcoes internas ===========================================================
func SFaturaToFatura(sf SFatura) (f Fatura) {
	f.ID_Fatura = sf.ID_Fatura.String
	f.ID_Origem = sf.ID_Origem.String
	f.ID_Destino = sf.ID_Destino.String
	f.Status = sf.Status.String
	f.Valor = sf.Valor.String

	if sf.DataCadastro.Valid {
		f.DataCadastro = sf.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		f.DataCadastro = ""
	}

	if sf.DataUltimoStatus.Valid {
		f.DataUltimoStatus = sf.DataUltimoStatus.Time.Format("02/01/2006 15:04:05")
	} else {
		f.DataUltimoStatus = ""
	}

	if sf.DataVencimento.Valid {
		f.DataVencimento = sf.DataVencimento.Time.Format("02/01/2006")
	} else {
		f.DataVencimento = ""
	}

	if sf.NomeRep.Valid {
		f.NomeDestino = sf.NomeRep.String
	} else if sf.NomeFra.Valid {
		f.NomeDestino = sf.NomeFra.String
	} else if sf.NomeCli.Valid {
		f.NomeDestino = sf.NomeCli.String
	} else {
		f.NomeDestino = "NÃO INFORMADO"
	}

	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			faturas.ID_Fatura, 
			faturas.ID_Origem, 
			faturas.ID_Destino, 
			faturas.Status, 
			faturas.Valor, 
			faturas.DataCadastro, 
			faturas.DataUltimoStatus, 
			faturas.DataVencimento, 

			destRep.RazaoSocial,
			destFra.RazaoSocial,
			destCli.Nome
		FROM faturas 

		LEFT JOIN representante AS destRep
		ON faturas.ID_Destino = destRep.ID_Representante

		LEFT JOIN franqueado AS destFra
		ON faturas.ID_Destino = destFra.ID_Franqueado
		
		LEFT JOIN cliente AS destCli
		ON faturas.ID_Destino = destCli.ID_Cliente
		
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, f *Fatura) error {

	var tmp SFatura
	if err := tab.Scan(
		&tmp.ID_Fatura,
		&tmp.ID_Origem,
		&tmp.ID_Destino,
		&tmp.Status,
		&tmp.Valor,
		&tmp.DataCadastro,
		&tmp.DataUltimoStatus,
		&tmp.DataVencimento,
		&tmp.NomeRep,
		&tmp.NomeFra,
		&tmp.NomeCli,
	); err != nil {
		return err
	}

	*f = SFaturaToFatura(tmp)

	// Busca os itens da fatura
	var fi FaturaItem
	fi.ID_Fatura = f.ID_Fatura
	if err := fi.ListaByIdFatura(&f.Itens); err != nil {
		return err
	}
	return nil
}
