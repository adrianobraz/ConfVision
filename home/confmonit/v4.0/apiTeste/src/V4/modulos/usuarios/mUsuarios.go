package usuariosV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	listaEnvioV4 "api/src/V4/modulos/listaEnvio"
	"api/src/V4/seguranca"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Usuario struct {
	ID_Usuario     string `json:"idUsuario"`
	ID_Vinculo     string `json:"idVinculo"`
	Tipo           string `json:"tipo"`
	Nome           string `json:"nome"`
	Nick           string `json:"nick"`
	Email1         string `json:"email1"`
	Email2         string `json:"email2"`
	Senha          string `json:"senha"`
	Telefone1      string `json:"telefone1"`
	Telefone2      string `json:"telefone2"`
	UsuarioTeminal string `json:"usuarioTerminal"`
	UsuarioWeb     string `json:"usuarioWeb"`
	Master         string `json:"master"`
	Ativo          string `json:"ativo"`
	DataCadastro   string `json:"dataCriacao"`
	EnviarEmail    string `json:"enviarEmail"`
	EnviarSms      string `json:"enviarSms"`
}

type SUsuario struct {
	IdUsuario      sql.NullString
	IdVinculo      sql.NullString
	Nome           sql.NullString
	Nick           sql.NullString
	Email1         sql.NullString
	Email2         sql.NullString
	Senha          sql.NullString
	Telefone1      sql.NullString
	Telefone2      sql.NullString
	UsuarioTeminal sql.NullString
	UsuarioWeb     sql.NullString
	Master         sql.NullString
	DataCadastro   sql.NullTime
}

// Insere insere um novo usuario no sistema
func (u *Usuario) Insere() error {
	// Gera um id para o usuario caso venha vazio
	if u.ID_Usuario == "" {
		u.ID_Usuario = auxiliar.GeradorDeId()
	}

	// Valida se um id de vinculo foi informado\
	if u.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	// Valida Email 1
	if u.Email1 == "" {
		return errors.New("um email 1 deve ser informado")
	}

	if u.Master == "" {
		u.Master = "N"
	}

	var livre string
	if err := u.GetEmail1Livre(&livre); err != nil {
		return err
	}

	if livre == "S" {

		// Abre uma conexão
		db, err := connV4.Conectar()
		if err != nil {
			return err
		}
		defer db.Close()

		stm, err := db.Prepare(`
			 INSERT INTO usuarios(
				usuarios.ID_Usuario, 
				usuarios.ID_Vinculo, 
				usuarios.Nome, 
				usuarios.Nick, 
				usuarios.Email1, 
				usuarios.Email2, 
				usuarios.Senha,
				usuarios.Telefone1,
				usuarios.Telefone2,
				usuarios.Master
			) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		// Gera uma senha padrao com hash
		u.Senha, err = seguranca.HashString("usuario123")
		if err != nil {
			return err
		}

		if _, err := stm.Exec(
			u.ID_Usuario,
			u.ID_Vinculo,
			strings.ToUpper(u.Nome),
			strings.ToUpper(u.Nick),
			strings.ToLower(u.Email1),
			strings.ToLower(u.Email2),
			u.Senha,
			auxiliar.ClearTel(u.Telefone1),
			auxiliar.ClearTel(u.Telefone2),
			strings.ToUpper(u.Master),
		); err != nil {
			return err
		}

		// Insere usuario na lista envio ==========================================
		var le listaEnvioV4.ListaEnvio
		le.ID_Alvo = u.ID_Usuario
		le.Email = "S"
		le.Sms = "N"

		if err := le.Insere(); err != nil {
			return err
		}
	} else {
		return errors.New("email já em uso por outro usuário")
	}

	return nil
}

// GetDadosById recupera os dados de um usuario pelo seu id
func (u *Usuario) GetDadosById() error {
	// Valida se um id de usuario foi informado
	if u.ID_Usuario == "" {
		return errors.New("um id de usuário deve ser informado")
	}

	// Abre uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE usuarios.ID_Usuario = '%s'", u.ID_Usuario)

	// Cria a consulta
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processarItem(tab, u); err != nil {
			return err
		}

		return nil
	}

	return errors.New("Usuario não encontrado na base de dados")
}

// GetDadosById recupera os dados de um usuario pelo seu Email1
func (u *Usuario) GetDadosByEmail1() error {

	// Valida se um email foi informado
	if u.Email1 == "" {
		return errors.New("um email de usuário deve ser informado")
	}

	// Abre uma conexao
	db, err := connV4.Conectar()
	if err != nil {

		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE usuarios.Email1 = '%s'", u.Email1)

	// Cria a consulta
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	// Recupera os dado retornado da consulta
	if tab.Next() {
		if err := processarItem(tab, u); err != nil {
			return err
		}

		return nil
	}

	return errors.New("Usuario não encontrado na base de dados")
}

func (u *Usuario) GetIdVinculoByEmail1() error {

	// Valida se um email foi informado
	if u.Email1 == "" {
		return errors.New("um email de usuário deve ser informado")
	}

	// Abre uma conexao
	db, err := connV4.Conectar()
	if err != nil {

		return err
	}
	defer db.Close()

	// Cria a consulta
	tab, err := db.Query(`
		SELECT usuarios.ID_Vinculo
		FROM usuarios
		WHERE usuarios.Email1 = ?
	`, u.Email1)
	if err != nil {
		return err
	}
	defer tab.Close()

	// Recupera os dado retornado da consulta
	if tab.Next() {
		if err := tab.Scan(&u.ID_Vinculo); err != nil {
			return err
		}

		return nil
	}

	return errors.New("Usuario não encontrado na base de dados")
}

func (u *Usuario) GetEmail1Livre(livre *string) error {

	// Valida se um email foi informado
	if u.Email1 == "" {
		return errors.New("um email deve ser informado")
	}

	// Abre uma conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria a consulta
	tab, err := db.Query(`
		SELECT usuarios.ID_Usuario  
		FROM usuarios 
		WHERE usuarios.Email1 = ?
	`, u.Email1)
	if err != nil {
		return err
	}
	defer tab.Close()

	// Recupera os dado retornado da consulta
	if tab.Next() {
		*livre = "N"
	} else {
		*livre = "S"
	}

	return nil
}

func (u *Usuario) GetEmail2Livre(livre *string) error {

	// Valida se um email foi informado
	if u.Email2 == "" {
		return errors.New("um email deve ser informado")
	}

	// Abre uma conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria a consulta
	tab, err := db.Query(`
		SELECT usuarios.ID_Usuario  
		FROM usuarios 
		WHERE usuarios.Email2 = ?
	`, u.Email2)
	if err != nil {
		return err
	}
	defer tab.Close()

	// Recupera os dado retornado da consulta
	if tab.Next() {
		*livre = "N"
	} else {
		*livre = "S"
	}

	return nil
}

// Altera altera os dados do usuario
func (u *Usuario) AlteraById() error {
	// Valida se um id de usuario foi informado
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	// Valida se um id de usuario foi informado
	if u.Email1 == "" {
		return errors.New("um email 1 deve informado")
	}

	// Abre uma conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria a consulta
	stm, err := db.Prepare(`
		UPDATE usuarios SET 
			usuarios.Nome = ?,
			usuarios.Nick = ?,
			usuarios.Email1 = ?,
			usuarios.Email2 = ?,
			usuarios.Telefone1 = ?,
			usuarios.Telefone2 = ?
		WHERE usuarios.ID_Usuario = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	// Executa a consulta
	if _, err := stm.Exec(
		strings.ToUpper(u.Nome),
		strings.ToUpper(u.Nick),
		strings.ToLower(u.Email1),
		strings.ToLower(u.Email2),
		auxiliar.ClearTel(u.Telefone1),
		auxiliar.ClearTel(u.Telefone2),
		u.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) DeletaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`DELETE FROM usuarios WHERE usuarios.ID_Usuario = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(u.ID_Usuario); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) DeletaAllByVinculo() error {
	if u.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`DELETE FROM usuarios WHERE usuarios.ID_Vinculo = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(u.ID_Vinculo); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) Listar(lista *[]Usuario) error {
	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(getSelect(""))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Usuario
		if err := processarItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}

	return nil
}

func (u *Usuario) ListarByVinculo(lista *[]Usuario) error {

	if u.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE usuarios.ID_Vinculo = '%s'", u.ID_Vinculo)
	//fmt.Println(u.getSelect(filtro))
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Usuario
		if err := processarItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

func (u *Usuario) ListarByVinculoToMaster(lista *[]Usuario) error {
	// Valida se um id de vinculo foi informado
	if u.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}
	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		"WHERE usuarios.Master = 'S' AND usuarios.ID_Vinculo = %s", u.ID_Vinculo,
	)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Usuario
		if err := processarItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}
	return nil
}

func (u *Usuario) ListarByVinculoToNotMaster(lista *[]Usuario) error {
	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		"WHERE usuarios.Master = 'N' AND usuarios.ID_Vinculo = %s", u.ID_Vinculo,
	)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Usuario
		if err := processarItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}

	return nil
}

func (u *Usuario) ListarByVinculoToAtivo(lista *[]Usuario) error {
	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		"WHERE usuarios.Ativo = 'S' AND usuarios.ID_Vinculo = %s", u.ID_Vinculo,
	)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Usuario
		if err := processarItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}

	return nil
}

func (u *Usuario) ListarByVinculoToNotAtivo(lista *[]Usuario) error {
	// Abre um canal de conexao
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		"WHERE usuarios.Ativo = 'N' AND usuarios.ID_Vinculo = %s", u.ID_Vinculo,
	)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Usuario
		if err := processarItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}

	return nil
}

// Funcoes para manipular a ativação do usuario ===============================
func (u *Usuario) GetUsuarioAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = u.ID_Usuario

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	u.Ativo = lb.Ativo

	return nil
}

func (u *Usuario) SetUsuarioAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	if u.Ativo == "" {
		return errors.New("um status de ativo deve informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = u.ID_Usuario
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = u.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) InverteUsuarioAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = u.ID_Usuario
	lb.Descricao = "ALTERADO VIA WEB"
	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	u.Ativo = lb.Ativo
	return nil
}

// Funcoes para manipular a ativação do uso do terminal =======================
func (u *Usuario) GetTerminalAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT usuarios.UsuarioTeminal 
		FROM usuarios 
		WHERE usuarios.ID_Usuario = ?
	`, u.ID_Usuario)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&u.UsuarioTeminal); err != nil {
			return err
		}
		return nil
	}
	return errors.New("usuario não encontrado na base de dados")
}

func (u *Usuario) SetTerminalAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
			UPDATE usuarios 
			SET usuarios.UsuarioTeminal = ? 			
			WHERE usuarios.ID_Usuario = ?
		`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		u.UsuarioTeminal,
		u.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) InverteTerminalAtivaById() error {

	if err := u.GetTerminalAtivaById(); err != nil {
		return err
	}
	fmt.Println(u.Ativo)
	// Inverte o estado de u.Ativo
	if err := auxiliar.InverteEstado(&u.UsuarioTeminal); err != nil {
		return err
	}

	if err := u.SetTerminalAtivaById(); err != nil {
		return err
	}

	return nil
}

// Funcoes para manipular a ativação do uso da web ============================
func (u *Usuario) GetWebAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT usuarios.UsuarioWeb 
		FROM usuarios 
		WHERE usuarios.ID_Usuario = ?
	`, u.ID_Usuario)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&u.UsuarioWeb); err != nil {
			return err
		}
		return nil
	}
	return errors.New("usuario não encontrado na base de dados")
}

func (u *Usuario) SetWebAtivaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
			UPDATE usuarios 
			SET usuarios.UsuarioWeb = ? 			
			WHERE usuarios.ID_Usuario = ?
		`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		u.UsuarioWeb,
		u.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) InverteWebAtivaById() error {

	if err := u.GetWebAtivaById(); err != nil {
		return err
	}

	// Inverte o estado de u.Ativo
	if err := auxiliar.InverteEstado(&u.UsuarioWeb); err != nil {
		return err
	}

	if err := u.SetWebAtivaById(); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) ResetarSenhaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	// Gera uma senha padrao com hash
	senha, err := seguranca.Hash("usuario123")
	if err != nil { //usuario123
		return err
	}
	u.Senha = string(senha)
	fmt.Println(u.ID_Usuario, u.Senha)

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
			UPDATE usuarios 
			SET usuarios.Senha = ? 			
			WHERE usuarios.ID_Usuario = ?
		`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		u.Senha,
		u.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) AlterarSenhaById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	if u.Senha == "" {
		return errors.New("uma senha de usuario deve ser informada")
	}

	// Gera uma senha padrao com hash
	senha, err := seguranca.HashString(u.Senha)
	if err != nil {
		return err
	}

	u.Senha = senha

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
			UPDATE usuarios 
			SET usuarios.Senha = ? 			
			WHERE usuarios.ID_Usuario = ?
		`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		u.Senha,
		u.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

// Funcoes para manipular o envio de emails ===================================
func (u *Usuario) GetAtivarEnviarEmailById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
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
	`, u.ID_Usuario)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var tmp sql.NullString
		if err := tab.Scan(&tmp); err != nil {
			return err
		}
		if tmp.Valid {
			u.EnviarEmail = tmp.String
		}
	} else {
		u.EnviarEmail = "N"

		var le listaEnvioV4.ListaEnvio
		le.ID_Alvo = u.ID_Usuario
		if err := le.Insere(); err != nil {
			return err
		}

	}
	return nil
}

func (u *Usuario) SetAtivarEnviarEmailById() error {
	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
			UPDATE listaEnvio 
			SET listaEnvio.Email = ? 			
			WHERE listaEnvio.ID_Alvo = ?
		`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		u.EnviarEmail,
		u.ID_Usuario,
	); err != nil {
		return err
	}

	return nil
}

func (u *Usuario) InverteAtivarEnviarEmailById() error {

	if u.ID_Usuario == "" {
		return errors.New("um id de usuario deve ser informado")
	}

	if err := u.GetAtivarEnviarEmailById(); err != nil {
		return err
	}
	fmt.Println(u.EnviarEmail)
	if err := auxiliar.InverteEstado(&u.EnviarEmail); err != nil {
		return err
	}
	fmt.Println(u.EnviarEmail)

	if err := u.SetAtivarEnviarEmailById(); err != nil {
		return err
	}
	return nil
}

// Funções internas ===========================================================
func sUsuarioToUsuario(s SUsuario) (u Usuario) {
	u.ID_Usuario = s.IdUsuario.String
	u.ID_Vinculo = s.IdVinculo.String
	u.Nome = s.Nome.String
	u.Nick = s.Nick.String
	u.Email1 = s.Email1.String
	u.Email2 = s.Email2.String
	u.Senha = s.Senha.String
	u.Telefone1 = s.Telefone1.String
	u.Telefone2 = s.Telefone2.String
	u.UsuarioTeminal = s.UsuarioTeminal.String
	u.UsuarioWeb = s.UsuarioWeb.String
	u.Master = s.Master.String
	u.DataCadastro = s.DataCadastro.Time.Format("02/01/2006 15:04:05")
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			usuarios.ID_Usuario,
			usuarios.ID_Vinculo,
			usuarios.Nome,
			usuarios.Nick,
			usuarios.Email1,
			usuarios.Email2,		
			usuarios.Senha,		
			usuarios.Telefone1,
			usuarios.Telefone2,
			usuarios.UsuarioTeminal,
			usuarios.UsuarioWeb,
			usuarios.Master,
			usuarios.DataCadastro,

			rep.ID_Representante AS tipoRep,
			fra.ID_Franqueado AS tipoFra,
			cli.ID_Cliente AS tipoCli,
			
			blUser.ID_Alvo AS blocUser,
			blRepId.ID_Alvo AS blocRepId, 	
			blFraId.ID_Alvo AS blocFraId,
			blFraRep.ID_Alvo AS blocFraRep,			
			blCliId.ID_Alvo AS blocCliId,
			blCliFra.ID_Alvo AS blocCliFra,
			blCliRep.ID_Alvo AS blocCliRep,

			listaEnvio.Email,
			listaEnvio.Sms

		FROM usuarios

		LEFT JOIN listaBloqueio AS blUser
		ON usuarios.ID_Usuario = blUser.ID_Alvo
		

		LEFT JOIN representante AS rep
		ON usuarios.ID_Vinculo = rep.ID_Representante

		LEFT JOIN listaBloqueio AS blRepId
		ON rep.ID_Representante = blRepId.ID_Alvo



		LEFT JOIN franqueado AS fra
		ON usuarios.ID_Vinculo = fra.ID_Franqueado

		LEFT JOIN listaBloqueio AS blFraId
		ON fra.ID_Franqueado = blFraId.ID_Alvo
		
		LEFT JOIN listaBloqueio AS blFraRep
		ON fra.ID_Representante = blFraRep.ID_Alvo

		
		
		LEFT JOIN cliente AS cli
		ON usuarios.ID_Vinculo = cli.ID_Cliente
		
		LEFT JOIN franqueado AS cliFra
		ON cli.ID_Franqueado = cliFra.ID_Franqueado
		
		LEFT JOIN listaBloqueio AS blCliId
		ON cli.ID_Cliente = blCliId.ID_Alvo

		LEFT JOIN listaBloqueio AS blCliFra
		ON cli.ID_Franqueado = blCliFra.ID_Alvo

		LEFT JOIN listaBloqueio AS blCliRep
		ON cliFra.ID_Representante = blCliRep.ID_Alvo

		LEFT JOIN listaEnvio 
		ON usuarios.ID_Usuario = listaEnvio.ID_Alvo

		%s
	`, filtro)
}

func processarItem(tab *sql.Rows, u *Usuario) error {
	var (
		temp SUsuario

		tipoRep sql.NullString
		tipoFra sql.NullString
		tipoCli sql.NullString

		blocUser sql.NullString

		blocRepId sql.NullString

		blocFraId  sql.NullString
		blocFraRep sql.NullString

		blocCliId  sql.NullString
		blocCliFra sql.NullString
		blocCliRep sql.NullString

		enviaEmail sql.NullString
		enviaSms   sql.NullString
	)

	if err := tab.Scan(
		&temp.IdUsuario,
		&temp.IdVinculo,
		&temp.Nome,
		&temp.Nick,
		&temp.Email1,
		&temp.Email2,
		&temp.Senha,
		&temp.Telefone1,
		&temp.Telefone2,
		&temp.UsuarioTeminal,
		&temp.UsuarioWeb,
		&temp.Master,
		&temp.DataCadastro,

		&tipoRep,
		&tipoFra,
		&tipoCli,

		&blocUser,

		&blocRepId,

		&blocFraId,
		&blocFraRep,

		&blocCliId,
		&blocCliFra,
		&blocCliRep,

		&enviaEmail,
		&enviaSms,
	); err != nil {
		return err
	}

	*u = sUsuarioToUsuario(temp)

	if enviaEmail.Valid {
		u.EnviarEmail = enviaEmail.String
	} else {
		u.EnviarEmail = "N"
	}

	if enviaSms.Valid {
		u.EnviarSms = enviaSms.String
	} else {
		u.EnviarSms = "N"
	}

	if temp.IdVinculo.String == "CENTRAL" {
		u.Tipo = "CEN"
	} else if tipoRep.Valid {
		u.Tipo = "REP"
	} else if tipoFra.Valid {
		u.Tipo = "FRA"
	} else if tipoCli.Valid {
		u.Tipo = "CLI"
	}

	if blocUser.Valid {
		u.Ativo = "N"
	} else if blocRepId.Valid {
		u.Ativo = "N"
	} else if blocFraId.Valid {
		u.Ativo = "N"
	} else if blocFraRep.Valid {
		u.Ativo = "N"
	} else if blocCliId.Valid {
		u.Ativo = "N"
	} else if blocCliFra.Valid {
		u.Ativo = "N"
	} else if blocCliRep.Valid {
		u.Ativo = "N"
	} else {
		u.Ativo = "S"
	}

	return nil
}
