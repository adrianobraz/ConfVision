package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	IdReceptor         string
	QtdConexoes        int
	
	StringConexaoBanco string
	MonitorarEnviar    int

	SenhaWeb string
	PortaWeb string

	ExibirLog          bool
	LigarMonitoramento bool

	API struct {
		Url            string
		SecretKey      string
		AtualizarToken int
		Logado         bool
		Token          string
	}

	// INFORMAÇÕES PARA ENVIO DE EMAIL
	Mail struct {
		Host      string
		Porta     string
		User      string
		Senha     string
		Remetente string
		Admin     string // INFORMAÇÃO DO EMAIL DO ADMINISTRADOR
	}

	Benuven struct {
		Email string
		Senha string

		ExibirLog bool
	}
)

func Configurar() {

	QtdConexoes = 0

	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro.Error())
	}
	IdReceptor = os.Getenv("ID_RECEPTOR")

	// Configura Conexao com a api
	API.Url = os.Getenv("API_URL")
	API.SecretKey = os.Getenv("API_SECRET_KEY")
	API.AtualizarToken = 0
	API.Logado = false
	API.Token = ""

	// Configura o montitoramento do receptor
	MonitorarEnviar, _ = strconv.Atoi(os.Getenv("TEMPO_INTERVALO_MONITORAMENTO"))

	StringConexaoBanco = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("BD_USER"),
		os.Getenv("BD_PASS"),
		os.Getenv("BD_HOST"),
		os.Getenv("BD_PORT"),
		os.Getenv("BD_BASE"),
	)

	Mail.Porta = os.Getenv("MAIL_PORTA")
	Mail.Host = os.Getenv("MAIL_HOST")
	Mail.User = os.Getenv("MAIL_USER")
	Mail.Senha = os.Getenv("MAIL_SENHA")
	Mail.Remetente = os.Getenv("MAIL_REMETENTE")
	Mail.Admin = os.Getenv("MAIL_ADMIN")

	//*************************************************//

	//************* Configura o Receptor **************//

	SenhaWeb = os.Getenv("SENHA_WEB")
	PortaWeb = os.Getenv("PORTA_MASTER")

	if os.Getenv("EXIBIR_LOG") == "ON" {
		ExibirLog = true
	} else {
		ExibirLog = false
	}

	if os.Getenv("LIGAR_MONITORAMENTO") == "ON" {
		LigarMonitoramento = true
	} else {
		LigarMonitoramento = false
	}

	Benuven.Email = os.Getenv("BENUVEM_EMAIL")
	Benuven.Senha = os.Getenv("BENUVEM_SENHA")

	//*************************************************//

}
