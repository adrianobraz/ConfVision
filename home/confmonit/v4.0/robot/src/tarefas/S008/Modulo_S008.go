/*
*******************************************************************************
Modulo S008 -> Efetua o bloqueio financeiro do do sistema
*******************************************************************************
*/
package S008

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	"robot/src/config"
	"time"
)

type fat struct {
	IdFatura string
	Destino  string
}

func Start(horario int, diasBloqueio time.Duration) {
	fmt.Println("Subindo modulo S009 -> Gerenciado bloqueio finaceiro")
	time.Sleep(100 * time.Millisecond)

	for {

		hora := time.Now().Hour()

		if hora == horario {

			db, err := auxiliar.Conectar()
			if err != nil {
				fmt.Println(err)
				break
			}
			defer db.Close()

			var lista []fat
			if err := buscarLista(db, diasBloqueio, &lista); err != nil {
				fmt.Println(err)
				break
			}
			db.Close()
		}

		time.Sleep(time.Hour)
	}

	go Start(horario, diasBloqueio)
}

func buscarLista(db *sql.DB, diasBloqueio time.Duration, lista *[]fat) error {

	data := time.Now().Add(diasBloqueio * (24 * time.Hour)).Format("2006-01-02")

	tab, err := db.Query(`
		SELECT 
			faturas.ID_Fatura,
			faturas.ID_Destino
		FROM faturas
		WHERE faturas.DataVencimento < ?
		AND faturas.Status = "AGUARDANDO PAGAMENTO"
	`, data)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var i fat

		if err := tab.Scan(
			&i.IdFatura,
			i.Destino,
		); err != nil {
			return err
		}

		*lista = append(*lista, i)
	}

	return nil
}

func ProcessarLista(db *sql.DB, lista []fat) error {

	if len(lista) > 0 {
		for _, i := range lista {
			fmt.Printf("Modulo S008 -> Processando fatura %s\n", i.IdFatura)
			stm, err := db.Prepare(`
				INSERT INTO listaBloqueio(
				listaBloqueio.ID_Alvo, 
				listaBloqueio.Descricao
				) VALUES (?,?,?)
			`)
			if err != nil {
				return err
			}
			defer stm.Close()

			descricao := fmt.Sprintf("Bloqueado %s devido atrazo na fatura de numero: %s", config.Robot.Nome, i.IdFatura)
			if _, err := stm.Exec(
				i.Destino,
				descricao,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
