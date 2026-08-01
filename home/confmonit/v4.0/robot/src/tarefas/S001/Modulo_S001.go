/*
	MODULO S001

Este Modulo tem com função retirar os dispositivo de manutenção
após o tempo de manutenção ter espirado
*/
package S001

import (
	"fmt"
	"robot/src/auxiliar"
	"time"
)

type S001 struct {
}

func Start(tempo time.Duration) {
	fmt.Println("Subindo modulo S001 -> Retira da manutenção")
	time.Sleep(100 * time.Millisecond)
	for {
		// Pausa entre operações
		time.Sleep(time.Minute * tempo)
		// time.Sleep(time.Second * tempo)

		fmt.Println("Modulo S001 -> Processando lista manuteção")

		if err := retiraDaLista(); err != nil {
			fmt.Println("Erro modulo S001 ->", err.Error())
			break
		}

	}

	fmt.Println("Reiniciando modulo S001 -> Retira dispositivo de manutenção")

	go Start(tempo)
}

// Funçoes internas ===========================================================

// busca os dispositivos que estao em manutencao
func retiraDaLista() error {

	// Abre um canal de conexao
	db, erro := auxiliar.Conectar()
	if erro != nil {
		fmt.Println(erro)
		return erro
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM listaManutencao WHERE listaManutencao.DataRetirada < ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		time.Now().Format("2006-01-02 15:04:05"),
	); err != nil {
		return err
	}
	return nil
}
