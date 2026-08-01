package servidor

import (
	"fmt"
	"log"
	"net/http"
	"terminal/src/roteador"
	"terminal/src/setup"
)

func ServidorStartHttp() {

	// Imprime no console dados do servidor
	fmt.Printf("%s ovindo na porta: %s\n", setup.Servidor, setup.Porta)

	// Inicia o servidor
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", setup.Porta), roteador.ConfigurarRotas()))
}

func ServidorStartHttps() {

	// Imprime no console dados do servidor
	fmt.Printf("%s ovindo na porta: %s\n", setup.Servidor, setup.Porta)

	log.Fatal(
		http.ListenAndServeTLS(
			fmt.Sprintf(":%s", setup.Porta),
			setup.Certificado,
			setup.Chave,
			roteador.ConfigurarRotas(),
		),
	)
}
