package listaBoqueioV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type ListaBloqueio struct {
	ID_Alvo      string `json:"idAlvoBloqueio"`
	DataBloqueio string `json:"dataBloqueio"`
	DataRetirada string `json:"dataRetirada"`
	Descricao    string `json:"descricao"`
	Ativo        string `json:"ativo"`
}

type SListaBloqueio struct {
	ID_Alvo      sql.NullString
	DataBloqueio sql.NullTime
	DataRetirada sql.NullTime
	Descricao    sql.NullString
}

func (lb *ListaBloqueio) Insere() error {

	if lb.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	if lb.Descricao == "" {
		return errors.New("uma descrição deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	if lb.DataRetirada == "" { // Insere sem data de retirada

		stm, err := db.Prepare(`
			INSERT INTO listaBloqueio (
				listaBloqueio.ID_Alvo,
				listaBloqueio.Descricao
			) VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(
			lb.ID_Alvo,
			lb.Descricao,
		); err != nil {
			return err
		}
	} else { // Insere com data de retirada
		stm, err := db.Prepare(`
			INSERT INTO listaBloqueio (
				listaBloqueio.ID_Alvo,
				listaBloqueio.DataRetirada,
				listaBloqueio.Descricao
			) VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(
			lb.ID_Alvo,
			lb.DataRetirada,
			lb.Descricao,
		); err != nil {
			return err
		}
	}

	return nil
}

func (lb *ListaBloqueio) DeleteByIdAlvo() error {
	if lb.ID_Alvo == "" {
		return errors.New("um id de alvo de bloqueio deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM listaBloqueio 
		WHERE listaBloqueio.ID_Alvo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(lb.ID_Alvo); err != nil {
		return err
	}

	return nil
}

func (lb *ListaBloqueio) GetBloquadoByIdAlvo() error {
	if lb.ID_Alvo == "" {
		return errors.New("um id de alvo de bloqueio deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT listaBloqueio.ID_Alvo
		FROM listaBloqueio 

		WHERE listaBloqueio.ID_Alvo = ?
	`, lb.ID_Alvo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		lb.Ativo = "S"
	} else {
		lb.Ativo = "N"
	}
	return nil
}

func (lb *ListaBloqueio) Lista(lista *[]ListaBloqueio) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
		listaBloqueio.ID_Alvo, 
		listaBloqueio.DataBloqueio,
		listaBloqueio.DataRetirada, 
		listaBloqueio.Descricao
		FROM listaBloqueio
	`)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item ListaBloqueio

		if err := item.processaItem(tab); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

func (lb *ListaBloqueio) SListaBloqueioToListaBloqueio(slb SListaBloqueio) {
	lb.ID_Alvo = slb.ID_Alvo.String

	lb.DataBloqueio = slb.DataBloqueio.Time.Format("02/01/2006 15:04:05")

	lb.DataRetirada = slb.DataRetirada.Time.Format("02/01/2006 15:04:05")

	lb.Descricao = slb.Descricao.String
}

// Funçoes internas sem rota =======================================
func (lb *ListaBloqueio) GetBloqueado() error {
	if lb.ID_Alvo == "" {
		return errors.New("um id de alvo deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			listaBloqueio.DataRetirada
		FROM listaBloqueio
		WHERE listaBloqueio.ID_Alvo = ?
	`, lb.ID_Alvo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var dtReit sql.NullString
		if err := tab.Scan(
			&dtReit,
		); err != nil {
			return err
		}

		lb.DataRetirada = dtReit.String
		lb.Ativo = "N"
	} else {
		lb.DataRetirada = ""
		lb.Ativo = "S"
	}
	return nil
}

func (lb *ListaBloqueio) SetBloqueado() error {
	if lb.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	if lb.Descricao == "" {
		return errors.New("uma descrição deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()
	
	if lb.Ativo == "S" {
		stm, err := db.Prepare(`
			DELETE FROM listaBloqueio WHERE listaBloqueio.ID_Alvo = ?
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(lb.ID_Alvo); err != nil {
			return err
		}
	} else {

		// Verifica se o idalvo ja esta na lista
		tabLista, err := db.Query(`
			SELECT listaBloqueio.ID_Alvo
			FROM listaBloqueio
			WHERE listaBloqueio.ID_Alvo = ?
		`, lb.ID_Alvo)
		if err != nil {
			fmt.Println("erro", err)
			return err
		}
		defer tabLista.Close()

		if tabLista.Next() {
			// Caso ja esteja ele so retorna
			return nil
		} else {

			// Caso não ele insere
			stm, err := db.Prepare(`
						INSERT INTO listaBloqueio(
							listaBloqueio.ID_Alvo, 
							listaBloqueio.DataRetirada,
							listaBloqueio.Descricao
						) VALUES ( ?, ?, ? )
					`)
			if err != nil {

				return err
			}
			defer stm.Close()

			var data sql.NullString
			if lb.DataRetirada != "" {
				data.String = lb.DataRetirada
			}

			if _, err := stm.Exec(
				lb.ID_Alvo,
				data,
				lb.Descricao,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

func (lb *ListaBloqueio) InverteBloqueado() error {
	if lb.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}
	
	if lb.Descricao == "" {
		return errors.New("uma descrição deve ser informada")
	}

	// Busca o estaedo atual
	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	fmt.Println("Antes : ", lb.Ativo)
	if err := auxiliar.InverteEstado(&lb.Ativo); err != nil {
		return err
	}
	fmt.Println("Depois: ", lb.Ativo)

	// Grava o novo estado
	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

// Funções internas ===========================================================

func (lb *ListaBloqueio) processaItem(tab *sql.Rows) error {
	var tmp SListaBloqueio
	if err := tab.Scan(
		&tmp.ID_Alvo,
		&tmp.DataBloqueio,
		&tmp.DataRetirada,
		&tmp.Descricao,
	); err != nil {
		return err
	}

	lb.SListaBloqueioToListaBloqueio(tmp)

	return nil
}
