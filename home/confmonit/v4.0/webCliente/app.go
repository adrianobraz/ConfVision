package main

import (
	"fmt"
	"log"
	"net/http"
	"webCliente/src/config"
	"webCliente/src/router"
	"webCliente/src/seguranca"
)

func main() {
	config.Configurar()
	seguranca.ConfigurarCookies()

	if config.SiteHttps {
		webHttps()
	} else {
		webHttp()
	}

}

func webHttp() {
	fmt.Printf("Web-Central rodando em HTTP na porta %s\n", config.Porta)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.Porta), router.Carregar()))
}

func webHttps() {

	fmt.Printf("Web-Central rodando em HTTPS na porta %s\n", config.Porta)

	log.Fatal(
		http.ListenAndServeTLS(
			fmt.Sprintf(":%s", config.Porta),
			config.Certificado,
			config.Chave,
			router.Carregar(),
		),
	)
}
