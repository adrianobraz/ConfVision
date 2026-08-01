package fabricanteV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Fabricante struct {
	ID_Fabricante string `json:"idFabricante"`
	Nome          string `json:"nome"`
	KeepVisivel   string `json:"keepVisivel"`
	Ativo         string `json:"ativo"`
	DataCadastro  string `json:"dataCriacao"`
}

type SFabricante struct {
	ID_Fabricante sql.NullString
	Nome          sql.NullString
	KeepVisivel   sql.NullString
	DataCadastro  sql.NullTime
}

func (f *Fabricante) Insere() error {

	// if f.ID_Fabricante == "" {
	// 	f.ID_Fabricante = auxiliar.GeradorDeId()
	// }

	if f.Nome == "" {
		return errors.New("um nome de fabricante deve ser informado")
	}

	if f.KeepVisivel == "" {
		return errors.New("um estado de keep alive deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {

		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO fabricantes(
			fabricantes.Nome, 
			fabricantes.KeepVisivel
		) VALUES (?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(f.Nome),
		strings.ToUpper(f.KeepVisivel),
	); err != nil {
		return err
	}

	return nil
}

func (f *Fabricante) GetDadosById() error {
	if f.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE fabricantes.ID_Fabricante = %s", f.ID_Fabricante)
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
	return errors.New("fabricante não encontrado na base de dados")
}

func (f *Fabricante) AlterarById() error {

	if f.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE fabricantes 
		SET 
			fabricantes.Nome = ?, 
			fabricantes.KeepVisivel = ?
		WHERE fabricantes.ID_Fabricante = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(f.Nome),
		strings.ToUpper(f.KeepVisivel),
		f.ID_Fabricante,
	); err != nil {
		return err
	}

	return nil
}

func (f *Fabricante) DeletaById() error {

	// Deleta os modelos do fabricante
	if err := f.deletaAllByIdFabricante(); err != nil {
		return err
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM fabricantes WHERE fabricantes.ID_Fabricante = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Fabricante); err != nil {
		return err
	}

	return nil
}

func (f *Fabricante) Listar(lista *[]Fabricante) error {

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(getSelect(""))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Fabricante
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// Funcoes pra maipular o campo Ativo =========================================
func (f *Fabricante) GetAtivoById() error {
	if f.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = f.ID_Fabricante

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	f.Ativo = lb.Ativo

	return nil
}

func (f *Fabricante) SetAtivoById() error {
	if f.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	if f.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = f.ID_Fabricante
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = f.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}
	return nil
}

func (f *Fabricante) InverteAtivoById() error {
	if f.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = f.ID_Fabricante
	lb.Descricao = "ALTERADO VIA WEB"

	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	f.Ativo = lb.Ativo
	return nil
}

// Funções internas =====================================================================

func sFabricanteToFabricante(sf SFabricante) (f Fabricante) {
	f.ID_Fabricante = sf.ID_Fabricante.String
	f.Nome = sf.Nome.String
	f.KeepVisivel = sf.KeepVisivel.String
	f.DataCadastro = sf.DataCadastro.Time.Format("02/01/2006 15:04:05")
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			fabricantes.ID_Fabricante,
			fabricantes.Nome,
			fabricantes.KeepVisivel,
			fabricantes.DataCadastro,

			listaBloqueio.ID_Alvo

		FROM fabricantes

		LEFT JOIN listaBloqueio 
		ON fabricantes.ID_Fabricante = listaBloqueio.ID_Alvo
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, f *Fabricante) error {
	var (
		tmp            SFabricante
		bloqFabricante sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Fabricante,
		&tmp.Nome,
		&tmp.KeepVisivel,
		&tmp.DataCadastro,

		&bloqFabricante,
	); err != nil {
		return err
	}

	*f = sFabricanteToFabricante(tmp)

	if bloqFabricante.Valid {
		f.Ativo = "N"
	} else {
		f.Ativo = "S"
	}

	return nil
}

func (f *Fabricante) deletaAllByIdFabricante() error {
	if f.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM modeloCentral WHERE modeloCentral.ID_Fabricante = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Fabricante); err != nil {
		return err
	}

	return nil
}
