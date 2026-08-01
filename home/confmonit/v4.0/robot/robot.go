package main

import (
	"errors"
	"fmt"
	"robot/src/auxiliar"
	"robot/src/config"
	"robot/src/tarefas/S001"
	"robot/src/tarefas/S002"
	"robot/src/tarefas/S003"
	"robot/src/tarefas/S004"
	"robot/src/tarefas/S005"
	"robot/src/tarefas/S006"
	"robot/src/tarefas/S007"
	"robot/src/tarefas/S008"
	"robot/src/tarefas/S009"
	"robot/src/tarefas/S010"
	"time"
)

//###############################################################################

func AreaTeste() {
	erro := errors.New("teste")
	fmt.Println("Reiniciando modulo S001 -> Retira dispositivo de manutenção")

	messagem := fmt.Sprintf(`
	<h3>
		Modulo S002 -> Retira dispositivo de manutenção foi reinicado devido ao erro %s em %s
	</h3>`, erro.Error(), time.Now().Format("02/01/2006 15:04:05"))

	var e auxiliar.Email
	if erro := e.EnviaEmailDireto(config.EmailAdmin, messagem); erro != nil {
		fmt.Println(erro)
	}
	//var fat S004.ObjAutoGerarFatura
	//fat.Executar("")
	//S004.EXECUTAR()

} 

// MODULO S001 - Retirar da manutencao ok
// MODULO S002 - Atendimento Automatico ok
// MODULO S003 - Sistema Nao Armado ok
// MODULO S004 - Auto Gerar Faturas ok
// MODULO S005 -
// MODULO S006 - Controle dipositivos sem comunicacao + 24H
// MODULO S007 - Manutenção das tabelas
// MODULO S008 - Bloqueio Finaceiro
// MODULO S009 - Envio de email
// MODULO S010 - Finaliza processo >2h sem atendente

func servicos() {
	if config.Liga.S001 {
		go S001.Start(2) // em Minutos
	}

	if config.Liga.S002 {
		go S002.Start(120) // em segundos
	}

	if config.Liga.S003 {
		go S003.Start(180) // em segundos
	}

	if config.Liga.S004 {
		go S004.Start(1) // em que hora do dia
	}

	if config.Liga.S005 {
		go S005.Start(0)
	}

	if config.Liga.S006{
		go S006.Start(15) // em que hora do dia ex: 15 processa as 3 horas da tarde
	}

	if config.Liga.S007{
		go S007.Start(11) // em que hora do dia
	}

	if config.Liga.S008{
		go S008.Start(24, 10) // horas
	}

	if config.Liga.S009{
		go S009.Start(30) // Segundos
	}

	if config.Liga.S010 {
		go S010.Start(120) // em segundos
	}

}

func main() {
	// Configura a API
	config.Configurar()
	//AreaTeste()

	servicos()

	for {
		time.Sleep(time.Minute * 60)
	}
}
