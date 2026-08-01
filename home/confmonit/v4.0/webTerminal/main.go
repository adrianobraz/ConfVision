package main

import (
	"terminal/src/servidor"
	"terminal/src/setup"
	"terminal/src/templates"
)

func main() {

	// Configuração inicial do app
	setup.ConfigurarApp()

	// Carrega os templates para uso
	templates.CarregarTemplates()
	areaTeste()
	// Carrega o servidor
	servidor.ServidorStartHttp()
}

func areaTeste() {}
