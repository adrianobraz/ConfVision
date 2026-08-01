package main

import (
	"fmt"
	"log"
	"net/http"

	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/roteador"
	"confvision/src/seguranca"
)

func endereco() string {
	if config.BindHost != "" {
		return fmt.Sprintf("%s:%s", config.BindHost, config.WebPorta)
	}
	return fmt.Sprintf(":%s", config.WebPorta)
}

func webHttps() {
	fmt.Printf("ConfVision rodando em HTTPS em %s", endereco())
	log.Fatal(http.ListenAndServeTLS(
		endereco(),
		config.Certificado,
		config.Chave,
		roteador.ConfigurarRotas(),
	))
}

func webHttp() {
	fmt.Printf("ConfVision rodando em HTTP em %s", endereco())
	log.Fatal(http.ListenAndServe(
		endereco(),
		roteador.ConfigurarRotas(),
	))
}

func main() {
	config.ConfigurarApp()
	auxiliar.CarregarTemplates()
	seguranca.ConfigurarCookies()

	if config.SiteHttps {
		webHttps()
	} else {
		webHttp()
	}
}
