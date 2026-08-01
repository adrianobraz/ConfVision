package main

import (
	"log"
	"terminal/assets/modulos/device"
	"terminal/src/servidor"
	"terminal/src/setup"
	"terminal/src/templates"
)

func main() {

	// Configuração inicial do app
	setup.ConfigurarApp()

	// Carrega os templates para uso
	templates.CarregarTemplates()

	if err := device.EnsureTable(); err != nil {
		log.Println("aviso: tabela operador_fcm_token:", err)
	}

	areaTeste()
	// Carrega o servidor
	servidor.ServidorStartHttp()
}

func areaTeste() {}
