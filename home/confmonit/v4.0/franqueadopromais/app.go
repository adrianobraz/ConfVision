package main

import (
	"fmt"
	"log"
	"net/http"
	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"
	"franqueadopro/src/roteador"
	"franqueadopro/src/seguranca"
)

func webHttps() {
	fmt.Printf("Web-Franqueado rodando em HTTPS na porta %s", config.WebPorta)

	log.Fatal(
		http.ListenAndServeTLS(
			fmt.Sprintf(":%s", config.WebPorta),
			config.Certificado,
			config.Chave,
			roteador.ConfigurarRotas(),
		),
	)

}

func webHttp() {
	fmt.Printf("Web-Franqueado rodando em HTTP na porta %s", config.WebPorta)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(
		":%s", config.WebPorta),
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
