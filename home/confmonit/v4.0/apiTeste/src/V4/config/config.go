package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	// ApiPorta possui a porta usada para conexão
	ApiPorta        string
	SecretKey       []byte
	ConexaoV4Master string
	ConexaoV4Slave  string

	Benuven struct {
		ExibirLog   bool
		Email       string
		Senha       string
		Token       string
		TokenTipo   string
		TokenExpira int
	}

	RepInformativo string
)

func Configurar() {

	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro.Error())
	}

	//** Configura a porta em que o servidor escuta **//
	ApiPorta = os.Getenv("API_PORTA")
	//*************************************************//

	SecretKey = []byte(os.Getenv("API_KEY"))
	//*************************************************//

	//**** Configura a conexao com o banco de dados master V4 ***//
	ConexaoV4Master = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("BD_USER_MV4"),
		os.Getenv("BD_PASS_MV4"),
		os.Getenv("BD_HOST_MV4"),
		os.Getenv("BD_PORT_MV4"),
		os.Getenv("BD_BASE_MV4"),
	)
	//*************************************************//

	//**** Configura a conexao com o banco de dados slave V4 ***//
	ConexaoV4Slave = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("BD_USER_SV4"),
		os.Getenv("BD_PASS_SV4"),
		os.Getenv("BD_HOST_SV4"),
		os.Getenv("BD_PORT_SV4"),
		os.Getenv("BD_BASE_SV4"),
	)
	//*************************************************//

	//**** Configura o acesso ao sistema da Benuvem ***//
	Benuven.Email = os.Getenv("BENUVEM_EMAIL")
	Benuven.Senha = os.Getenv("BENUVEM_SENHA")
	if os.Getenv("BENUVEM_EXIBIR_LOG") == "ON" {
		Benuven.ExibirLog = true
	} else {
		Benuven.ExibirLog = false
	}
	//*************************************************//

	RepInformativo = os.Getenv("REP_INFORMATIVO")
}
