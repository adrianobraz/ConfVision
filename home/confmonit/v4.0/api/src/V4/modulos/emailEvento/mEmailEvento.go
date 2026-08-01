package emailEventoV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type EmailEvento struct {
	ID_SetupEnvioEvento string `json:"idSetupEnvioEvento"`
	ID_Dispositivo      string `json:"idDispositivo"`
	EmailAlarme         string `json:"emailAlarme"`
	EmailArme           string `json:"emailArme"`
	EmailDesarme        string `json:"emailDesarme"`
	EmailEmergencia     string `json:"emailEmergencia"`
	EmailFalhas         string `json:"emailFalhas"`
	EmailGeral          string `json:"emailGeral"`
	EmailMedico         string `json:"emailMedico"`
	EmailPanico         string `json:"emailPanico"`
	EmailRestaure       string `json:"emailRestaure"`
	EmailSetup          string `json:"emailSetup"`
	EmailTeste          string `json:"emailTeste"`
	EmailSemComunicar   string `json:"emailSemComunicar"`
}

type SEmailEvento struct {
	ID_SetupEnvioEvento sql.NullString
	ID_Dispositivo      sql.NullString
	EmailAlarme         sql.NullString
	EmailArme           sql.NullString
	EmailDesarme        sql.NullString
	EmailEmergencia     sql.NullString
	EmailFalhas         sql.NullString
	EmailGeral          sql.NullString
	EmailMedico         sql.NullString
	EmailPanico         sql.NullString
	EmailRestaure       sql.NullString
	EmailSetup          sql.NullString
	EmailTeste          sql.NullString
	EmailSemComunicar   sql.NullString
}

func (ee *EmailEvento) Insere() error {
	if ee.ID_SetupEnvioEvento == "" {
		ee.ID_SetupEnvioEvento = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO setupEnvioEvento(
			ID_SetupEnvioEvento, 
			ID_Dispositivo, 
			EmailAlarme, 
			EmailArme, 
			EmailDesarme, 
			EmailEmergencia, 
			EmailFalhas, 
			EmailGeral, 
			EmailMedico, 
			EmailPanico, 
			EmailRestaure, 
			EmailSetup, 
			EmailTeste, 
			EmailSemComunicar
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()
	
	if _, err := stm.Exec(
		ee.ID_SetupEnvioEvento,
		ee.ID_Dispositivo,
		ee.EmailAlarme,
		ee.EmailArme,
		ee.EmailDesarme,
		ee.EmailEmergencia,
		ee.EmailFalhas,
		ee.EmailGeral,
		ee.EmailMedico,
		ee.EmailPanico,
		ee.EmailRestaure,
		ee.EmailSetup,
		ee.EmailTeste,
		ee.EmailSemComunicar,
	); err != nil {
		return err
	}

	return nil
}

func (ee *EmailEvento) GetDadosById() error {
	if ee.ID_SetupEnvioEvento == "" {
		return errors.New("um id de setup envia evento deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE setupEnvioEvento.ID_SetupEnvioEvento = '%s'
	`, ee.ID_SetupEnvioEvento)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, ee); err != nil {
			return err
		}

		return nil
	}

	return errors.New("setup não encontrado na base de dados")
}

func (ee *EmailEvento) GetDadosByIdDispositivo() error {
	if ee.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE setupEnvioEvento.ID_Dispositivo = '%s'
	`, ee.ID_Dispositivo)
	
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, ee); err != nil {
			return err
		}

		return nil
	}

	return errors.New("setup não encontrado na base de dados")
}

func (ee *EmailEvento) AlteraById() error {
	if ee.ID_SetupEnvioEvento == "" {
		return errors.New("um id de setup envio evento deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE setupEnvioEvento SET 
			setupEnvioEvento.EmailAlarme = ?,
			setupEnvioEvento.EmailArme = ?,
			setupEnvioEvento.EmailDesarme = ?,
			setupEnvioEvento.EmailEmergencia = ?, 
			setupEnvioEvento.EmailFalhas = ?,
			setupEnvioEvento.EmailGeral = ?,
			setupEnvioEvento.EmailMedico = ?, 
			setupEnvioEvento.EmailPanico = ?, 
			setupEnvioEvento.EmailRestaure = ?,
			setupEnvioEvento.EmailSetup = ?, 
			setupEnvioEvento.EmailTeste = ?,
			setupEnvioEvento.EmailSemComunicar = ?
		WHERE setupEnvioEvento.ID_SetupEnvioEvento = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		ee.EmailAlarme,
		ee.EmailArme,
		ee.EmailDesarme,
		ee.EmailEmergencia,
		ee.EmailFalhas,
		ee.EmailGeral,
		ee.EmailMedico,
		ee.EmailPanico,
		ee.EmailRestaure,
		ee.EmailSetup,
		ee.EmailTeste,
		ee.EmailSemComunicar,
		ee.ID_SetupEnvioEvento,
	); err != nil {
		return err
	}
	return nil
}

func (ee *EmailEvento) DeletaById() error {
	return nil
}

func (ee *EmailEvento) DeletaByIdDispositivo() error {
	return nil
}

//

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			setupEnvioEvento.ID_SetupEnvioEvento, 
			setupEnvioEvento.ID_Dispositivo, 
			setupEnvioEvento.EmailAlarme, 
			setupEnvioEvento.EmailArme, 
			setupEnvioEvento.EmailDesarme, 
			setupEnvioEvento.EmailEmergencia, 
			setupEnvioEvento.EmailFalhas, 
			setupEnvioEvento.EmailGeral, 
			setupEnvioEvento.EmailMedico, 
			setupEnvioEvento.EmailPanico, 
			setupEnvioEvento.EmailRestaure, 
			setupEnvioEvento.EmailSetup, 
			setupEnvioEvento.EmailTeste, 
			setupEnvioEvento.EmailSemComunicar 
		FROM setupEnvioEvento 
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, ee *EmailEvento) error {

	var (
		tmp SEmailEvento
	)

	if err := tab.Scan(
		&tmp.ID_SetupEnvioEvento,
		&tmp.ID_Dispositivo,
		&tmp.EmailAlarme,
		&tmp.EmailArme,
		&tmp.EmailDesarme,
		&tmp.EmailEmergencia,
		&tmp.EmailFalhas,
		&tmp.EmailGeral,
		&tmp.EmailMedico,
		&tmp.EmailPanico,
		&tmp.EmailRestaure,
		&tmp.EmailSetup,
		&tmp.EmailTeste,
		&tmp.EmailSemComunicar,
	); err != nil {
		return nil
	}

	*ee = SClienteToCliente(tmp)

	return nil
}

func SClienteToCliente(see SEmailEvento) (ee EmailEvento) {
	ee.ID_SetupEnvioEvento = see.ID_SetupEnvioEvento.String
	ee.ID_Dispositivo = see.ID_Dispositivo.String
	ee.EmailAlarme = see.EmailAlarme.String
	ee.EmailArme = see.EmailArme.String
	ee.EmailDesarme = see.EmailDesarme.String
	ee.EmailEmergencia = see.EmailEmergencia.String
	ee.EmailFalhas = see.EmailFalhas.String
	ee.EmailGeral = see.EmailGeral.String
	ee.EmailMedico = see.EmailMedico.String
	ee.EmailPanico = see.EmailPanico.String
	ee.EmailRestaure = see.EmailRestaure.String
	ee.EmailSetup = see.EmailSetup.String
	ee.EmailTeste = see.EmailTeste.String
	ee.EmailSemComunicar = see.EmailSemComunicar.String

	return
}
