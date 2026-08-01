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

	go servidor("2010")
	servidor(config.ApiPorta)

}

func corsExtra(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func servidor(porta string) {
	fmt.Printf("\nAPI rodando na porta: %s\n\n", porta)

	handler := corsExtra(roteadorV4.Configurar())

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", porta), handlers.CORS(
		handlers.AllowedOriginValidator(func(origin string) bool {
			return true
		}),
		handlers.AllowedMethods([]string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		}),
		handlers.AllowedHeaders([]string{
			"XMLHttpRequest",
			"X-Requested-With",
			"Authorization",
			"Content-Type",
			"Accept",
			"Origin",
		}),
		handlers.OptionStatusCode(http.StatusNoContent),
	)(handler)))
}

//###############################################################################

// funcao usada para testar codigos
func AreaTeste() {

}
