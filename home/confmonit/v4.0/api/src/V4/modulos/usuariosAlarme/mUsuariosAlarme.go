package usuariosAlarmeV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	listaEnvioV4 "api/src/V4/modulos/listaEnvio"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type UsuarioAlarme struct {
	ID_Usuario     string `json:"idUsuario"`
	ID_Dispositivo string `json:"idDispositivo"`
	Codigo         string `json:"codigo"`
	Nome           string `json:"nome"`
	Email          string `json:"email"`
	Celular        string `json:"celular"`
	Observacao     string `json:"observacao"`
	Ativo          string `json:"ativo"`
	DataCadastro   string `json:"dataCadastro"`

	EmailEnvio string `json:"emailEnvio"`
	SmsEnvio   string `json:"smsEnvio"`
}

type SUsuarioAlarme struct {
	ID_Usuario     sql.NullString
	ID_Dispositivo sql.NullString
	Codigo         sql.NullString
	Nome           sql.NullString
	Email          sql.NullString
	Celular        sql.NullString
	Observacao     sql.NullString
	DataCadastro   sql.NullTime
}

func (ua *UsuarioAlarme) Insere() error {
	if ua.ID_Usuario == "" {
		ua.ID_Usuario = auxiliar.GeradorDeId()
	}

	if ua.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if ua.Nome == "" {
		return errors.New("um nome para o dispositivo deve ser informado")
	}

	if ua.Codigo == "" {
		return errors.New("um codigo para o dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO usuariosAlarme(
			usuariosAlarme.ID_Usuario, 
			usuariosAlarme.ID_Dispositivo, 
			usuariosAlarme.Codigo, 
			usuariosAlarme.Nome, 
			usuariosAlarme.Email, 
			usuariosAlarme.Celular, 
			usuariosAlarme.Observacao 
			) VALUES ( ?, ?, ?, ?, ?, ?, ? ) 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		ua.ID_Usuario,
		ua.ID_Dispositivo,
		ua.Codigo,
		strings.ToUpper(ua.Nome),
		strings.ToLower(ua.Email),
		ua.Celular,
		ua.Observacao,
	); err != nil {
		return err
	}

	// Insere lista envio =====================================================
	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = ua.ID_Usuario
	le.Email = "N"
	le.Sms = "N"

	if err := le.Insere(); err != nil {
		return err
	}

	return nil
}

func (ua *UsuarioAlarme) GetDadosById() error {
	if ua.ID_Usuario == "" {
		return errors.New("um id de usuario de alarme deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE usuariosAlarme.ID_Usuario = '%s'`, ua.ID_Usuario)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, ua); err != nil {
			return err
		}
		return nil
	}

	return errors.New("usuario alarme não encontrado na base de dados")
}

func (ua *UsuarioAlarme) AlteraById() error {
	if ua.ID_Usuario == "" {
		return errors.New("um id de usuario alarme deve ser informado")
	}

	if ua.Nome == "" {
		return errors.New("um nome para o dispositivo deve ser informado")
	}

	if ua.Codigo == "" {
		return errors.New("um codigo para o dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE usuariosAlarme SET 
			usuariosAlarme.Codigo = ?, 
			usuariosAlarme.Nome = ?, 
			usuariosAlarme.Email = ?,
			usuariosAlarme.Celular = ?, 
			usuariosAlarme.Observacao = ?
		WHERE usuariosAlarme.ID_Usuario = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		ua.Codigo,
		strings.ToUpper(ua.Nome),
		strings.ToLower(ua.Email),
		ua.Celular,
		ua.Observacao,
		ua.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

func (ua *UsuarioAlarme) DeletaById() error {
	if ua.ID_Usuario == "" {
		return errors.New("um id de usuario alarme deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM usuariosAlarme WHERE usuariosAlarme.ID_Usuario = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(ua.ID_Usuario); err != nil {
		return err
	}

	return nil
}

func (ua *UsuarioAlarme) DeletaAllByIdDispositivo() error {
	if ua.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM usuariosAlarme WHERE usuariosAlarme.ID_Dispositivo
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(ua.ID_Dispositivo); err != nil {
		return err
	}

	return nil
}

func (ua *UsuarioAlarme) ListaByIdDispositivo(lista *[]UsuarioAlarme) error {
	if ua.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE usuariosAlarme.ID_Dispositivo = '%s'
	`, ua.ID_Dispositivo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item UsuarioAlarme
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

// Manipula o campo ativo =====================================================
func (ua *UsuarioAlarme) GetAtivoByid() error {
	if ua.ID_Usuario == "" {
		return errors.New("um id de usuario alarme deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = ua.ID_Usuario
	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	ua.Ativo = lb.Ativo

	return nil
}

func (ua *UsuarioAlarme) SetAtivoByid() error {
	if ua.ID_Usuario == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	if ua.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = ua.ID_Usuario
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = ua.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (ua *UsuarioAlarme) InverteAtivoByid() error {
	if ua.ID_Usuario == "" {
		return errors.New("um id de procedimento deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = ua.ID_Usuario
	lb.Descricao = "ALTERADO VIA WEB"
	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	ua.Ativo = lb.Ativo
	return nil
}

// funcoes internas ===========================================================

func sUsuarioAlarmeToUsuarioAlarme(sua SUsuarioAlarme) (ua UsuarioAlarme) {
	ua.ID_Usuario = sua.ID_Usuario.String
	ua.ID_Dispositivo = sua.ID_Dispositivo.String
	ua.Codigo = sua.Codigo.String
	ua.Nome = sua.Nome.String
	ua.Email = sua.Email.String
	ua.Celular = sua.Celular.String
	ua.Observacao = sua.Observacao.String
	ua.DataCadastro = sua.DataCadastro.Time.Format("02/01/2006 15:04:05")
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			usuariosAlarme.ID_Usuario,
			usuariosAlarme.ID_Dispositivo,
			usuariosAlarme.Codigo,
			usuariosAlarme.Nome,
			usuariosAlarme.Email,
			usuariosAlarme.Celular,
			usuariosAlarme.Observacao,
			usuariosAlarme.DataCadastro,

			listaBloqueio.ID_Alvo,

			listaEnvio.Email,
			listaEnvio.Sms

		FROM usuariosAlarme

		LEFT JOIN listaEnvio
		ON usuariosAlarme.ID_Usuario = listaEnvio.ID_Alvo

		LEFT JOIN listaBloqueio 
		ON  usuariosAlarme.ID_Usuario = listaBloqueio.ID_Alvo
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, ua *UsuarioAlarme) error {
	var (
		tmp            SUsuarioAlarme
		bloqUserAlarme sql.NullString
		emailEnvio     sql.NullString
		smsEnvio       sql.NullString
	)

	if err := tab.Scan(
		&tmp.ID_Usuario,
		&tmp.ID_Dispositivo,
		&tmp.Codigo,
		&tmp.Nome,
		&tmp.Email,
		&tmp.Celular,
		&tmp.Observacao,
		&tmp.DataCadastro,

		&bloqUserAlarme,

		&emailEnvio,
		&smsEnvio,
	); err != nil {
		return err
	}

	*ua = sUsuarioAlarmeToUsuarioAlarme(tmp)

	if emailEnvio.Valid {
		ua.EmailEnvio = emailEnvio.String
	} else {
		ua.EmailEnvio = "N"
	}

	if smsEnvio.Valid {
		ua.SmsEnvio = smsEnvio.String
	} else {
		ua.SmsEnvio = "N"
	}

	if bloqUserAlarme.Valid {
		ua.Ativo = "N"
	} else {
		ua.Ativo = "S"
	}
	return nil
}
