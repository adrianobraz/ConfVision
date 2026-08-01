package sendMail

import (
	"errors"
	"fmt"
	"net/smtp"
)

type Dados struct {
	Host      string
	Porta     string
	User      string
	Senha     string
	Remetente string
	Admin     string // INFORMAÇÃO DO EMAIL DO ADMINISTRADOR
}

var SendMail struct {
	Servidor    string
	Credenciais smtp.Auth
	Remetente   string
}

func Configurar(host, porta, usuario, senha, remetente string) {

	SendMail.Servidor = fmt.Sprintf("%s:%s", host, porta)

	SendMail.Credenciais = smtp.PlainAuth("", usuario, senha, host) //autenticação

	SendMail.Remetente = remetente
}

// enviar envia o email para o destinatario localweb
func Enviar(destinatario, messagem string) error {
	if destinatario == "" {
		return errors.New("um destinatario deve ser informado")
	}

	if messagem == "" {
		return errors.New("uma menssagem deve ser informado")
	}

	//Conecta com o servidor SMTP
	erro := smtp.SendMail(
		SendMail.Servidor,
		SendMail.Credenciais,
		SendMail.Remetente,
		[]string{destinatario},
		[]byte(messagem),
	)
	if erro != nil {
		return erro
	}
	return nil
}
