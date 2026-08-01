/*
*******************************************************
Modulo S007 -> Efetua a manutenção das tabelas do banco
efetuando backup e removendo dados perdidos
*******************************************************
*/
package S007

import (
	"fmt"
	"robot/src/tarefas/S007/tabProcesso"
	"time"
)

type S007 struct {
	FraqId     string
	FranqNome  string
	Autorizado bool
	// dispId         string
	// dispNome       string
	// dispDataEvt    string
	// dispCodigoFull string
}

func Start(horario int) {

	fmt.Println("Subindo modulo S007 -> Manutenção das tabelas do sistema")
	time.Sleep(100 * time.Millisecond)

	for {
		processar := time.Now().Hour()

		if processar == horario {

			fmt.Println("Modulo S006 -> Processando manutenção tabela")

			if erro := tabProcesso.BackupTabelaProcesso(4); erro != nil {
				fmt.Println(erro)
				break
			}
		}

		time.Sleep(time.Hour)
	}

	fmt.Println("Reiniciando modulo S007 -> Manutenção das tabelas")

	go Start(horario)
}
