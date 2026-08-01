/*
	MODULO S002

Efetua o atendimento automatico dos processos em
nivel 0
*/
package S002

import (
	"fmt"
	"robot/src/auxiliar"
	"robot/src/config"
	"time"
)

type S002 struct{}

func Start(tempo time.Duration) {

	fmt.Println("Subindo modulo S002 -> Atendimento automatico")
	time.Sleep(100 * time.Millisecond)

	var erro error

	for { // Loop infinito
		time.Sleep(tempo * time.Second)

		fmt.Println("Modulo S002 -> Processando Atendendo")

		var s S002
		if erro = s.buscarProcessos(); erro != nil {
			fmt.Println("Erro modulo S002 ->", erro.Error())
			break
		}

	}

	fmt.Println("Reiniciando modulo S002 -> Atendimento automatico")

	go Start(tempo)
}

func (s *S002) buscarProcessos() error {

	// Abre um canal de conexao
	db, erro := auxiliar.Conectar()
	if erro != nil {
		fmt.Println(erro.Error())
		//goto RESET
	}
	defer db.Close()

	descricao := fmt.Sprintf(
		"'ATENDIDO AUTOMATICAMENTE PELO %s POR SE UM PROCESSO DE NIVEL 0'",
		config.Robot.Nome,
	)

	stm, erro := db.Prepare(`
		UPDATE processo
		
		SET processo.DataAtenInicio = ?, 
			processo.DataAtenFim = ?, 
			processo.ID_Atendente = ?,
			processo.Descricao = ?

		WHERE processo.Nivel = 0
		
		AND processo.DataAtenFim IS NULL
	`)
	if erro != nil {
		fmt.Println(erro.Error())
		//goto RESET
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
