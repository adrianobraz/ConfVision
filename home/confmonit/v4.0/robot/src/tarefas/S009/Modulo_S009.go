/*
* Modulo S009 - Gerencia o envio de email apartir da tabela envio
 */
package S009

import (
	"database/sql"
	"errors"
	"fmt"
	"net/smtp"
	"regexp"
	"robot/src/auxiliar"
	"robot/src/config"
	"time"
)

type tEmail struct {
	idEmail      string
	idVinculo    string
	tipoVinculo  string
	nomeVinculo  string
	destinatario string
	menssagem    string
	// auxiliares
	chave string
	valor string
}

func Start(tempo time.Duration) {
	fmt.Println("Subindo modulo S009 -> Envio de email")
	time.Sleep(100 * time.Millisecond)

	bd, err := auxiliar.Conectar()
	if err != nil {
		fmt.Println("modulo S009 -> ", err)
		fmt.Println("modulo S009 -> Reiniciando modulo")
		go Start(tempo)
		return
	}
	defer bd.Close()

	for {
		time.Sleep(tempo * time.Second)
		fmt.Println("Modulo S009 -> Processando Emails")
		if err := EnviarLista(bd); err != nil {
			fmt.Println("modulo S009 -> ", err)
			break
		}
	}

	fmt.Println("modulo S009 -> Reiniciando modulo")
	go Start(tempo)
}

func EnviarLista(db *sql.DB) error {

	// Busco a lista de envio
	tab, err := db.Query(`
		SELECT 
			email.ID_Email, 
			email.ID_Vinculo, 
			email.TipoVinculo, 
			email.Destinatario, 
			email.Menssagem 
		FROM email
	`)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var email tEmail

		// Inicia a chave em N
		email.chave = "N"

		// inica o valor em 0,00
		email.valor = "0,00"

		// Carrega os dados do email
		if err := tab.Scan(
			&email.idEmail,
			&email.idVinculo,
			&email.tipoVinculo,
			&email.destinatario,
			&email.menssagem,
		); err != nil {
			return err
		}

		switch email.tipoVinculo {
		case "CEN":
			email.chave = "L"
			email.nomeVinculo = "CENTRAL"
		case "REP":
			if err := representante(db, &email); err != nil {
				return err
			}

		case "FRA":
			if err := franqueado(db, &email); err != nil {
				return err
			}

		case "CLI":
			if err := cliente(db, &email); err != nil {
				return err
			}

		}

		// Se chave = N ele não envia o email
		// Se chave = S ele envia e grava custo
		// se chave = L ele envia e não grava custo

		if email.chave != "N" {
			// Envia o email
			if erro := send(email.destinatario, email.menssagem); erro != nil {
				fmt.Printf("%s (%s) -> %s \n", email.nomeVinculo, email.destinatario, erro)
			} else {
				fmt.Printf("%s (%s) -> enviado com sucesso \n", email.nomeVinculo, email.destinatario)
			}

			time.Sleep(3 * time.Second)
		}

		// Se chave for s ele grava o custo 
		if email.chave == "S" {
			gravaCusto(
				db,
				email.idVinculo,
				email.destinatario,
				email.tipoVinculo,
				email.valor,
			)
		}

		// Deleta o email da lista
		if err := deletaEmail(db, email.idEmail); err != nil {
			return err
		}
	}

	return nil
}

// Funcoes internas =================================================

func send(destinatario, messagem string) error {

	// Valida email
	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if emailRegex.MatchString(destinatario) {
		credenciais := smtp.PlainAuth(
			"",
			config.Mail.User,
			config.Mail.Senha,
			config.Mail.Host,
		) //autenticação

		msg := fmt.Sprintf(
			"From: %s\nTo: %s\nSubject: %s\nContent-Type: text/html; charset=UTF-8\n\n%s",
			config.Mail.Remetente,
			destinatario,
			"Informativo Monitoramento",
			messagem,
		)

		servidor := fmt.Sprintf("%s:%s", config.Mail.Host, config.Mail.Porta)

		//conecta com o servidor SMTP
		erro := smtp.SendMail(
			servidor,
			credenciais,
			config.Mail.Remetente,
			[]string{destinatario},
			[]byte(msg),
		)
		if erro != nil {
			return erro
		}
	} else {
		return errors.New("email Invalido")
	}

	return nil
}

func gravaCusto(db *sql.DB, idVinculo, destinatario, tipo, valor string) error {
	if tipo != "CEN" {
		if idVinculo == "" {
			return errors.New("[gravaCusto]: um id de vinculo deve ser informado")
		}

		if destinatario == "" {
			return errors.New("[gravaCusto]: um destinatario deve ser informado")
		}

		if valor == "" {
			return errors.New("[gravaCusto]: um valor de debito deve ser informado")
		}

		stm, err := db.Prepare(`
		INSERT INTO tarifacao(
			ID_Tarifacao, 
			ID_Vinculo, 
			TipoOperacao, 
			DadoOperacao, 
			Credito, 
			Debito
			) VALUES (?, ?, ?, ?, ?, ?)
	`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(
			auxiliar.GeradorDeId(),
			idVinculo,
			"EMAIL",
			destinatario,
			"0,00",
			valor,
		); err != nil {
			return err
		}
	}
	return nil
}

func deletaEmail(db *sql.DB, id string) error {

	stm, err := db.Prepare(`
		DELETE FROM email WHERE email.ID_Email = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(id); err != nil {
		return err
	}

	return nil
}
