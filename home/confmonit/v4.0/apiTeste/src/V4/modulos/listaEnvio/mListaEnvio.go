package listaEnvioV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
)

type ListaEnvio struct {
	ID_Alvo string `json:"idListaAlvo"`
	Email   string `json:"email"`
	Sms     string `json:"sms"`
}

type SListaEnvio struct {
	ID_Alvo sql.NullString
	Email   sql.NullString
	Sms     sql.NullString
}

func (le *ListaEnvio) Insere() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	if le.Email == "" {
		le.Email = "S"
	}

	if le.Sms == "" {
		le.Sms = "N"
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO listaEnvio(
			listaEnvio.ID_Alvo, 
			listaEnvio.Email, 
			listaEnvio.Sms
		) VALUES (?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(le.ID_Alvo, le.Email, le.Sms); err != nil {
		return err
	}
	return nil
}

func (le *ListaEnvio) DeleteByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM listaEnvio 
		WHERE listaEnvio.ID_Alvo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(le.ID_Alvo); err != nil {
		return err
	}
	return nil
}

// Funções para manipular envio de email ======================================
func (le *ListaEnvio) GetEmailAtivoByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT listaEnvio.Email 
		FROM listaEnvio 
		WHERE listaEnvio.ID_Alvo = ?
	`, le.ID_Alvo)
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&le.Email); err != nil {
			return err
		}
		return nil
	}

	// Caso não encontre o item ele gera um item
	le.Email = "N"
	le.Sms = "N"

	if err := le.Insere(); err != nil {
		return err
	}
	return nil
}

func (le *ListaEnvio) SetEmailAtivoByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT listaEnvio.Email 
		FROM listaEnvio 
		WHERE listaEnvio.ID_Alvo = ?
	`, le.ID_Alvo)
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		stm, err := db.Prepare(`
			UPDATE listaEnvio 
			SET listaEnvio.Email = ? 
			WHERE listaEnvio.ID_Alvo = ? 
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(le.Email, le.ID_Alvo); err != nil {
			return err
		}
	} else { // Caso não encontre o item ele gera um item
		le.Sms = "N"
		if err := le.Insere(); err != nil {
			return err
		}
	}

	return nil
}

func (le *ListaEnvio) InverteEmailAtivoByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	if err := le.GetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	if err := auxiliar.InverteEstado(&le.Email); err != nil {
		return err
	}

	if err := le.SetEmailAtivoByIdAlvo(); err != nil {
		return err
	}
	return nil
}

// Funções para manipular envio de sms ========================================
func (le *ListaEnvio) GetSmsAtivoByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT listaEnvio.Sms 
		FROM listaEnvio 
		WHERE listaEnvio.ID_Alvo = ?
	`, le.ID_Alvo)
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&le.Sms); err != nil {
			return err
		}
		return nil
	}

	// Caso não encontre o item ele gera um item
	le.Email = "N"
	le.Sms = "N"

	if err := le.Insere(); err != nil {
		return err
	}

	return nil
}

func (le *ListaEnvio) SetSmsAtivoByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	if le.Sms == "" {
		return errors.New("um status de sms deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT listaEnvio.Sms 
		FROM listaEnvio 
		WHERE listaEnvio.ID_Alvo = ?
	`, le.ID_Alvo)
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {
		stm, err := db.Prepare(`
			UPDATE listaEnvio 
			SET listaEnvio.Sms = ? 
			WHERE listaEnvio.ID_Alvo = ? 
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(le.Sms, le.ID_Alvo); err != nil {
			return err
		}

	} else { // Caso não encontre o item ele gera um item
		le.Email = "N"
		if err := le.Insere(); err != nil {
			return err
		}
	}

	return nil
}

func (le *ListaEnvio) InverteSmsAtivoByIdAlvo() error {
	if le.ID_Alvo == "" {
		return errors.New("um id de alvo deve ser informado")
	}

	if err := le.GetSmsAtivoByIdAlvo(); err != nil {
		return err
	}

	if err := auxiliar.InverteEstado(&le.Sms); err != nil {
		return err
	}

	if err := le.SetSmsAtivoByIdAlvo(); err != nil {
		return err
	}
	return nil
}
