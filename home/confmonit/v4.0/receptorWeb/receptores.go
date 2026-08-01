package main

import (
	"log"
	"receptorWeb/src/benuvem"
	"receptorWeb/src/config"
	evento "receptorWeb/src/eventos"
	"receptorWeb/src/receptor"
	"receptorWeb/src/sendMail"

	"fmt"
	"net/http"
)

func main() {

	// Configura a API
	config.Configurar()

	sendMail.Configurar(
		config.Mail.Host,
		config.Mail.Porta,
		config.Mail.User,
		config.Mail.Senha,
		config.Mail.Remetente,
	)

	evento.Configurar(
		config.Mail.Host,
		config.Mail.Porta,
		config.Mail.User,
		config.Mail.Senha,
		config.Mail.Remetente,
	)

	// areaTeste()

	// inicia o serviço de acesso ao servidor benuvem
	go benuvem.Start(config.Benuven.Email, config.Benuven.Senha, 60, true)

	fmt.Printf("Inicando Receptor Web na porta: %s\n", config.PortaWeb)

	http.HandleFunc("/recebe-evento", receptor.ReceberEvento)
	http.HandleFunc("/recebe-evento-confvision", receptor.ReceberEventoConfVision)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", config.PortaWeb), nil))

}

func areaTeste() {

}

/*
{
    "idFranqueado": "6",
    "evento": "0001112001001",
    "senha": "WHdQkY&RX%W%4RArwm1Q"
}

ConfVision analitico:
{
    "idFranqueado": "6",
    "conta": "0042",
    "particao": "1",
    "zonauser": "5",
    "vis_evento_id": 123,
    "senha": "WHdQkY&RX%W%4RArwm1Q"
}
*/
