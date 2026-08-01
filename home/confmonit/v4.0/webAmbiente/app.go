package main

import (
	"fmt"
	"log"
	"net/http"
	"webAmbiente/src/config"
	"webAmbiente/src/router"
	"webAmbiente/src/seguranca"
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

func endereco() string {
	if config.BindHost != "" {
		return fmt.Sprintf("%s:%s", config.BindHost, config.Porta)
	}
	return fmt.Sprintf(":%s", config.Porta)
}

func webHttp() {
	fmt.Printf("WebAmbiente rodando em HTTP em %s\n", endereco())
	log.Fatal(http.ListenAndServe(endereco(), router.Carregar()))
}

func webHttps() {

	fmt.Printf("WebAmbiente rodando em HTTPS em %s\n", endereco())

	log.Fatal(
		http.ListenAndServeTLS(
			endereco(),
			config.Certificado,
			config.Chave,
			router.Carregar(),
		),
	)
}
