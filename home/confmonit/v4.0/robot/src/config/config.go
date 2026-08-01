package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	Robot struct {
		Nome string
	}
	// String para uso da conexao
	StringConexao   string
	Url             string
	SecretKey       []byte
	EmailAdmin      string
	SomaGradeCriada bool

	// INFORMAÇÕES PARA ENVIO DE EMAIL
	Mail struct {
		Host      string
		Porta     string
		User      string
		Remetente string
		Senha     string
	}

	// Extrutura para receber valores de ativação
	Liga struct {
		S001 bool
		S002 bool
		S003 bool
		S004 bool
		S005 bool
		S006 bool
		S007 bool
		S008 bool
		S009 bool
		S010 bool
	}
)

func Configurar() {

	if erro := godotenv.Load(); erro != nil {
		log.Fatal(erro.Error())
	}

	Robot.Nome = os.Getenv("NOME_ROBOT")

	Url = os.Getenv("API_URL")
	SecretKey = []byte(os.Getenv("API_KEY"))

	// Conexao Master
	StringConexao = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("BD_USER"),
		os.Getenv("BD_PASS"),
		os.Getenv("BD_HOST"),
		os.Getenv("BD_PORT"),
		os.Getenv("BD_BASE"),
	)
	EmailAdmin = os.Getenv("EMAIL_ADMIN")

	// Caso true ele soma cada grade cadastrada no valor do pacote e caso false
	// ele acrecenta somente uma grade para cada cliente que tenha pelomenos 1
	// grade cadastrada
	SomaGradeCriada = false

	//***** Configura o envio de Email do sistema *****//
	Mail.Porta = os.Getenv("MAIL_PORTA")
	Mail.Host = os.Getenv("MAIL_HOST")
	Mail.User = os.Getenv("MAIL_USER")
	Mail.Remetente = os.Getenv("MAIL_REMETENTE")
	Mail.Senha = os.Getenv("MAIL_SENHA")
	//*************************************************//

	//*************** Ativa os modulos ****************//
	Liga.S001 = stringToBool(os.Getenv("LIGA_S001"))
	Liga.S002 = stringToBool(os.Getenv("LIGA_S002"))
	Liga.S003 = stringToBool(os.Getenv("LIGA_S003"))
	Liga.S004 = stringToBool(os.Getenv("LIGA_S004"))
	Liga.S005 = stringToBool(os.Getenv("LIGA_S005"))
	Liga.S006 = stringToBool(os.Getenv("LIGA_S006"))
	Liga.S007 = stringToBool(os.Getenv("LIGA_S007"))
	Liga.S008 = stringToBool(os.Getenv("LIGA_S008"))
	Liga.S009 = stringToBool(os.Getenv("LIGA_S009"))
	Liga.S010 = stringToBool(os.Getenv("LIGA_S010"))
	//*************************************************//
}

func stringToBool(valor string) bool {

	if valor == "S" {
		return true
	} else {
		return false
	}
}
