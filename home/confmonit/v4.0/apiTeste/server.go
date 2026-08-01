package main

import (
	"api/src/V4/config"
	benuvemV4 "api/src/V4/modulos/benuvem"
	roteadorV4 "api/src/V4/roteador"

	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

func main() {
	// Configura a API
	config.Configurar()

	AreaTeste()

	// inicia o serviço de acesso ao servidor benuvem
	go benuvemV4.Start()
	//go nvoip.Start()

	// Sobe o servidor
	fmt.Printf("API rodando na porta: %s\n\n", config.ApiPorta)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.ApiPorta), handlers.CORS(
		handlers.AllowedOrigins([]string{
			//			"*",
		}),
		handlers.AllowedMethods([]string{
			"GET",
			"POST",
			"DELETE",
			"OPTIONS",
		}),
		handlers.AllowedHeaders([]string{
			"XMLHttpRequest",
			"X-Requested-With",
			"Authorization",
			"Content-Type",
		}),
	)(roteadorV4.Configurar())))

}

//###############################################################################

// funcao usada para testar codigos
func AreaTeste() {

}
