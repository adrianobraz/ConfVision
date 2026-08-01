/*
	MODULO S010

Finaliza automaticamente processos abertos ha mais de 2 horas
sem atendente (ID_Atendente = 0) e nivel > 0
*/
package S010

import (
	"fmt"
	"robot/src/auxiliar"
	"robot/src/config"
	"time"
)

type S010 struct{}

func Start(tempo time.Duration) {

	fmt.Println("Subindo modulo S010 -> Finaliza processo >2h sem atendente")
	time.Sleep(100 * time.Millisecond)

	var erro error

	for { // Loop infinito
		time.Sleep(tempo * time.Second)

		fmt.Println("Modulo S010 -> Processando finalizacao >2h sem atendente")

		var s S010
		if erro = s.buscarProcessos(); erro != nil {
			fmt.Println("Erro modulo S010 ->", erro.Error())
			break
		}

	}

	fmt.Println("Reiniciando modulo S010 -> Finaliza processo >2h sem atendente")

	go Start(tempo)
}

func (s *S010) buscarProcessos() error {

	db, erro := auxiliar.Conectar()
	if erro != nil {
		fmt.Println(erro.Error())
		return nil
	}
	defer db.Close()

	descricao := fmt.Sprintf(
		"FINALIZADO PELO SISTEMA APOS 2HRS ABERTO SEM ATENDENTE (%s)",
		config.Robot.Nome,
	)

	stm, erro := db.Prepare(`
		UPDATE processo
		SET
			processo.DataAtenInicio = ?,
			processo.DataAtenFim = ?,
			processo.ID_Atendente = ?,
			processo.Descricao = ?
		WHERE processo.DataAtenFim IS NULL
		AND processo.Nivel > 0
		AND (processo.ID_Atendente = '0' OR processo.ID_Atendente = '' OR processo.ID_Atendente IS NULL)
		AND processo.DataCriacao < DATE_SUB(NOW(), INTERVAL 2 HOUR)
	`)
	if erro != nil {
		fmt.Println(erro.Error())
		return nil
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		time.Now().Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		config.Robot.Nome,
		descricao,
	); erro != nil {
		fmt.Println(erro.Error())
	}

	return nil
}
