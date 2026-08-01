package main

import (
	"fmt"
	"log"
	"net/http"
	"webCentral/src/config"
	"webCentral/src/router"
	"webCentral/src/seguranca"
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
	fmt.Printf("Web-Central rodando em HTTP na porta %s", config.Porta)
	http.ListenAndServe(fmt.Sprintf(":%s", config.Porta), router.Carregar())
}

func webHttps() {

	fmt.Printf("Web-Central rodando em HTTPS na porta %s", config.Porta)

	log.Fatal(
		http.ListenAndServeTLS(
			fmt.Sprintf(":%s", config.Porta),
			config.Certificado,
			config.Chave,
			router.Carregar(),
		),
	)
}
