package auxiliar

import (
	"robot/src/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type Email struct {
	Status string `json:"status"`
	Dados  string `json:"dados"`
}

func (e *Email) EnviarEmailFranqueado(idFranqueado, msg string) error {

	// Cria o body para envio do franqueado
	body, erro := json.Marshal(map[string]string{
		"id":       idFranqueado,
		"messagem": msg,
	})
	if erro != nil {
		return erro
	}

	resp, erro := Requisitar(fmt.Sprintf("%s/SendEmailFranqueadoEnviar", config.Url), bytes.NewBuffer(body))
	if erro != nil {
		return erro
	}

	if resp.StatusCode >= 400 {
		return errors.New("erro ao enviar o email")
	}

	return nil
}

func (e *Email) EnviarEmailCliente(idCliente, msg string) error {

	// Cria o body para envio do franqueado
	body, erro := json.Marshal(map[string]string{
		"id":       idCliente,
		"messagem": msg,
	})
	if erro != nil {
		return erro
	}

	resp, erro := Requisitar(fmt.Sprintf("%s/SendEmailClienteEnviar", config.Url), bytes.NewBuffer(body))
	if erro != nil {
		return erro
	}

	if resp.StatusCode >= 400 {
		return errors.New("erro ao enviar o email")
	}

	return nil
}

func (e *Email) EnviaEmailDireto(destinatario, msg string) error {

	// Cria o body para envio do franqueado
	body, erro := json.Marshal(map[string]string{
		"emailAvulso": destinatario,
		"messagem":    msg,
	})
	if erro != nil {
		return erro
	}

	resp, erro := Requisitar("/SendEmailEnviar", bytes.NewBuffer(body))
	if erro != nil {
		return erro
	}

	if resp.StatusCode >= 400 {
		return errors.New("erro ao enviar o email")
	}

	return nil
}
