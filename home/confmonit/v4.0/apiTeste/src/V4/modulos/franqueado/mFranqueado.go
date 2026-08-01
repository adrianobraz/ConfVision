package franqueadoV4

import (
	connV4 "api/src/V4/conexao"
	contactidV4 "api/src/V4/modulos/contactid"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	listaEnvioV4 "api/src/V4/modulos/listaEnvio"
	pacotev4 "api/src/V4/modulos/pacote"
	procedimentosV4 "api/src/V4/modulos/procedimentos"
	usuariosV4 "api/src/V4/modulos/usuarios"
	"api/src/auxiliar"

	"api/src/V4/seguranca"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const naoEncontrado = "franqueado não encontrado na base de dados"

type Franqueado struct {
	ID_Representante string `json:"repId"`
	RepNome          string `json:"repNome"`

	ID_UsuarioMaster string             `json:"userId"`
	UserDados        usuariosV4.Usuario `json:"userDados,omitempty"`

	ID_Franqueado     string `json:"fraId"`
	ID_Pacote         string `json:"fraIdPacote"`
	RazaoSocial       string `json:"fraRazao"`
	NomeFantasia      string `json:"fraNome"`
	Cnpj              string `json:"fraCnpj"`
	InscricaoEstadual string `json:"fraInscricaoEstadual"`
	Cep               string `json:"fraCep"`
	Endereco          string `json:"fraEndereco"`
	Complemento       string `json:"fraComplemento"`
	Bairro            string `json:"fraBairro"`
	Cidade            string `json:"fraCidade"`
	Uf                string `json:"fraUf"`
	DataCadastro      string `json:"fraDataCadastro"`
	DataCancelamento  string `json:"fraDataCancelamento"`
	CodBenuvem        string `json:"fraCodBenuvem"`
	Observacao        string `json:"fraObservacao"`
	EmailEnvio        string `json:"fraEnviaEmail"`
	SmsEnvio          string `json:"fraEnviaSms"`
	Ativo             string `json:"fraAtivo"`

	//Auxiliares
}

type SFranqueado struct {
	ID_Representante sql.NullString
	RepNome          sql.NullString

	ID_UsuarioMaster sql.NullString
	UserDados        usuariosV4.SUsuario

	ID_Franqueado     sql.NullString
	ID_Pacote         sql.NullString
	RazaoSocial       sql.NullString
	NomeFantasia      sql.NullString
	Cnpj              sql.NullString
	InscricaoEstadual sql.NullString
	Cep               sql.NullString
	Endereco          sql.NullString
	Complemento       sql.NullString
	Bairro            sql.NullString
	Cidade            sql.NullString
	Uf                sql.NullString
	DataCadastro      sql.NullTime
	DataCancelamento  sql.NullTime
	CodBenuvem        sql.NullString
	Observacao        sql.NullString
	EmailEnvio        sql.NullString
	SmsEnvio          sql.NullString
}

type franqueadoLogin struct {
	Token string `json:"token"`
	usuariosV4.Usuario
}

func (fl *franqueadoLogin) logar(email, senha string) error {
	// Valida se um email foi informado
	if email == "" {
		return errors.New("um email deve ser informado")

	}

	// Valida se uma senha foi informada
	if senha == "" {
		return errors.New("uma senha deve ser informada")

	}
	var err error
	fl.Email1 = email
	if err := fl.GetDadosByEmail1(); err != nil {
		fmt.Println(err)
		return err
	}

	// Valida a senha
	if err = seguranca.VerificarSenha(fl.Senha, senha); err != nil {
		return err
	}

	if fl.Ativo == "S" {

		// Caso a senha esteja correta ele cria um token
		fl.Token, err = seguranca.CriarToken(fl.ID_Usuario)
		if err != nil {
			return err
		}
		// Retorna os dados para o requisitante
		return nil
	} else {
		return errors.New("usuario bloqueado")
	}
}

func (f *Franqueado) Insere() error {

	// Caso não receba um id de franqueado ele cria um
	if f.ID_Franqueado == "" {
		f.ID_Franqueado = auxiliar.GeradorDeId()
	}

	// Valida se idRepresentante foi informado
	if f.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	if f.CodBenuvem == "" {
		f.CodBenuvem = f.gerarCodBeNuvem()
	}

	// Gera um id para o usuario master
	idUsuarioMaster := auxiliar.GeradorDeId()

	// Seta o id do usuario master
	f.ID_UsuarioMaster = idUsuarioMaster

	// Abre um conexao com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria o Franqueado ============================================
	stm, err := db.Prepare(`
		INSERT INTO franqueado(
			franqueado.ID_Franqueado, 
			franqueado.ID_Representante, 
			franqueado.ID_UsuarioMaster, 
			franqueado.ID_Pacote, 
			franqueado.RazaoSocial, 
			franqueado.NomeFantasia, 
			franqueado.Cnpj, 
			franqueado.InscricaoEstadual, 
			franqueado.Cep, 
			franqueado.Endereco, 
			franqueado.Complemento, 
			franqueado.Bairro, 
			franqueado.Cidade, 
			franqueado.Uf,  
			franqueado.CodBenuvem,
			franqueado.Observacao
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		f.ID_Franqueado,
		f.ID_Representante,
		f.ID_UsuarioMaster,
		f.ID_Pacote,
		strings.ToUpper(f.RazaoSocial),
		strings.ToUpper(f.NomeFantasia),
		f.Cnpj,
		f.InscricaoEstadual,
		f.Cep,
		strings.ToUpper(f.Endereco),
		strings.ToUpper(f.Complemento),
		strings.ToUpper(f.Bairro),
		strings.ToUpper(f.Cidade),
		strings.ToUpper(f.Uf),
		f.CodBenuvem,
		strings.ToUpper(f.Observacao),
	); err != nil {
		return err
	}

	// Cria o usuario Master ========================================
	f.UserDados.ID_Usuario = idUsuarioMaster
	f.UserDados.ID_Vinculo = f.ID_Franqueado
	f.UserDados.UsuarioTeminal = "S"
	f.UserDados.Master = "S"

	if err := f.UserDados.Insere(); err != nil {
		// Remove o franqueado criado caso não consiga criar o usuario
		f.DeletaById()
		return err
	}

	// Insere um pacote padrao para o franqueado ====================
	var pct pacotev4.Pacote
	pct.ID_Vinculo = f.ID_Franqueado
	pct.Nome = "PADRÃO"
	pct.ContasQtd = "-1"
	pct.AtendimentoQtd = "-1"
	pct.EmailQtd = "-1"
	pct.EmailBloquear = "N"
	pct.SmsQtd = "-1"
	pct.SmsBloquear = "N"
	pct.LigacoesQtd = "-1"
	pct.LigacoesBloquear = "N"
	if err := pct.Insere(); err != nil {
		return err
	}

	// Insere franqueado na lista envio =============================
	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = f.ID_Franqueado
	if err := le.Insere(); err != nil {
		return err
	}

	return nil
}

func (f *Franqueado) GetDadosById() error {
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`AND franqueado.ID_Franqueado = '%s'`, f.ID_Franqueado)

	tab, err := db.Query(f.getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := f.processaItem(tab); err != nil {
			return err
		}
		return nil

	}
	return errors.New(naoEncontrado)
}

func (f *Franqueado) GetNomeById() (string, error) {
	if f.ID_Franqueado == "" {
		return "", errors.New("um id de franqueado deve ser informado")
	}
	db, err := connV4.Conectar()
	if err != nil {
		return "", err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT franqueado.RazaoSocial
		FROM franqueado
		WHERE franqueado.ID_Franqueado = ?
	`, f.ID_Franqueado)
	if err != nil {
		return "", err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&f.RazaoSocial); err != nil {
			return "", err
		}
		return f.RazaoSocial, nil
	}

	return "NE", nil
}

func (f *Franqueado) AlterarById() error {
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE franqueado SET 
			franqueado.ID_Pacote = ?,
			franqueado.RazaoSocial = ?,
			franqueado.NomeFantasia = ?,
			franqueado.Cnpj = ?, 
			franqueado.InscricaoEstadual = ?,
			franqueado.Cep = ?,
			franqueado.Endereco = ?,
			franqueado.Complemento = ?,
			franqueado.Bairro = ?,
			franqueado.Cidade = ?,
			franqueado.Uf = ?,
			franqueado.Observacao = ?
	
		WHERE franqueado.ID_Franqueado = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		f.ID_Pacote,
		strings.ToUpper(f.RazaoSocial),
		strings.ToUpper(f.NomeFantasia),
		f.Cnpj,
		f.InscricaoEstadual,
		f.Cep,
		strings.ToUpper(f.Endereco),
		strings.ToUpper(f.Complemento),
		strings.ToUpper(f.Bairro),
		strings.ToUpper(f.Cidade),
		strings.ToUpper(f.Uf),
		strings.ToUpper(f.Observacao),
		f.ID_Franqueado,
	); err != nil {
		return err
	}

	// Altera os dados do Usuario master
	if err := f.UserDados.AlteraById(); err != nil {
		return err
	}

	return nil
}

func (f *Franqueado) DeletaById() error {
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	// Abre um conexao com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Apaga os usuario usuarios do franqueado ======================
	var use usuariosV4.Usuario
	use.ID_Vinculo = f.ID_Franqueado
	if err := use.DeletaAllByVinculo(); err != nil {
		return err
	}

	// Apaga os pacotes do franqueado ===============================
	var pct pacotev4.Pacote
	pct.ID_Vinculo = f.ID_Franqueado
	if err := pct.DeletaAllByVinculo(); err != nil {
		return err
	}

	// remove o franqueado da lista bloqueio ========================
	var lBloc listaBoqueioV4.ListaBloqueio
	lBloc.ID_Alvo = f.ID_Franqueado
	if err := lBloc.DeleteByIdAlvo(); err != nil {
		return err
	}

	// remove o franqueado da lista envio ===========================
	var lEnvio listaEnvioV4.ListaEnvio
	lEnvio.ID_Alvo = f.ID_Franqueado
	if err := lEnvio.DeleteByIdAlvo(); err != nil {
		return err
	}

	// remove os contactid personalizado ============================
	var cid contactidV4.ContactId
	cid.ID_Viculo = f.ID_Franqueado
	if err := cid.DeletaAllByIdVinculo(); err != nil {
		return err
	}

	// remove os procedimentos ======================================
	var pro procedimentosV4.Procedimentos
	pro.ID_Franqueado = f.ID_Franqueado
	if err := pro.DeletaAllByIdFranqueado(); err != nil {
		return err
	}

	// move os clientes para tabela clientes desativados ============

	// Remove o franqueado
	stm, err := db.Prepare(`
		DELETE FROM franqueado 
		WHERE franqueado.ID_Franqueado = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Franqueado); err != nil {
		return err
	}

	return nil
}

// Corrigir esse para apagar vinculos
func (f *Franqueado) DeletaAllByRepresentante() error {
	if f.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Abre um conexao com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`DELETE FROM franqueado WHERE franqueado.ID_Representante = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(f.ID_Representante); err != nil {
		return err
	}

	return nil
}

func (f *Franqueado) ListarByIdRepresentante(lista *[]Franqueado) error {
	if f.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		AND franqueado.ID_Representante = '%s'
		ORDER BY franqueado.RazaoSocial
	`, f.ID_Representante)
	tab, err := db.Query(f.getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Franqueado
		if err := item.processaItem(tab); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

func (f *Franqueado) CancelaById() error {
	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria uma consulta de exclusao
	stm, err := db.Prepare(`
		UPDATE franqueado 
		SET franqueado.DataCancelamento = ?
		WHERE franqueado.ID_Franqueado = ?
	`)
	if err != nil {
		return nil
	}
	defer stm.Close()

	data := time.Now().Format("2006-01-02 15:04:05")
	// Executa a exclusao
	if _, err := stm.Exec(data, f.ID_Franqueado); err != nil {
		return err
	}
	//=========================================================================

	return nil
}

func (f *Franqueado) ReverteCancelaById() error {
	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria uma consulta de exclusao
	stm, err := db.Prepare(`
		UPDATE franqueado
		SET franqueado.DataCancelamento = NULL
		WHERE franqueado.ID_Franqueado = ?
	`)
	if err != nil {
		return nil
	}
	defer stm.Close()

	// Executa a exclusao
	if _, err := stm.Exec(f.ID_Franqueado); err != nil {
		return err
	}
	//=========================================================================

	return nil
}

//===================================================================
// funcoes de manipulacao do campo ativo do franqueado ==============

// GetAtivoById retorna o estado de ativo do franqueado
func (f *Franqueado) GetAtivoById() error {
	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = f.ID_Franqueado
	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	f.Ativo = lb.Ativo

	return nil
}

// SetAtivoById seta o estado de ativo do franqueado
func (f *Franqueado) SetAtivoById() error {

	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	if f.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = f.ID_Franqueado
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = f.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return nil
	}

	return nil
}

// InverterAtivoById inverte o estado de ativo do franqueado
func (f *Franqueado) InverterAtivoById() error {

	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = f.ID_Franqueado

	if lb.Descricao == "" {
		lb.Descricao = "ALTERADO VIA WEB"
	}

	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	f.Ativo = lb.Ativo

	return nil
}

//===================================================================
// funcoes de manipulação do campo de envio de email ================

// GetEmailAtivoById retorna o estado de envio de email
func (f *Franqueado) GetEmailAtivoById() error {
	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de representante deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = f.ID_Franqueado

	if err := le.GetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	f.EmailEnvio = le.Email
	return nil
}

// SetEmailAtivoById seta o estado de envio de email
func (f *Franqueado) SetEmailAtivoById() error {

	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	if f.EmailEnvio == "" {
		return errors.New("um estado de bloqueado deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = f.ID_Franqueado
	le.Email = f.EmailEnvio

	if err := le.SetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	return nil
}

// InverterEmailAtivoById inverte o estado de envio do email
func (f *Franqueado) InverterEmailAtivoById() error {

	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de Franqueado deve ser informado")
	}

	//Consulta o estado atual de ativo
	if err := f.GetEmailAtivoById(); err != nil {
		return err
	}

	// Inverte o estado de ativo ==============================================
	if err := auxiliar.InverteEstado(&f.EmailEnvio); err != nil {
		return err
	}

	if err := f.SetEmailAtivoById(); err != nil {
		return err
	}

	return nil
}

//===================================================================
// funcoes de manipulação do campo de envio de sms ==================

// GetSmsAtivoById retorna o estado de envio do sms
func (f *Franqueado) GetSmsAtivoById() error {
	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio

	le.ID_Alvo = f.ID_Franqueado

	if err := le.GetSmsAtivoByIdAlvo(); err != nil {
		return err
	}

	f.SmsEnvio = le.Sms

	return nil
}

// SetSmsAtivoById seta o estado de envio do sms
func (f *Franqueado) SetSmsAtivoById() error {

	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	if f.SmsEnvio == "" {
		return errors.New("um estado de bloqueado deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = f.ID_Franqueado
	le.Sms = f.SmsEnvio
	if err := le.SetSmsAtivoByIdAlvo(); err != nil {
		return err
	}

	return nil
}

// InverterSmsAtivoById inverte o estado de envio de sms
func (f *Franqueado) InverterSmsAtivoById() error {

	// Valida se um id de representante foi informado
	if f.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	//Consulta o estado atual de ativo
	if err := f.GetSmsAtivoById(); err != nil {
		return err
	}

	// Inverte o estado de ativo ==============================================
	if err := auxiliar.InverteEstado(&f.SmsEnvio); err != nil {
		return err
	}

	if err := f.SetSmsAtivoById(); err != nil {
		return err
	}

	return nil
}

// Funcoes internas =====================================================================

// ok
func (f *Franqueado) sFranqueadoToFranqueado(fra SFranqueado) {
	f.ID_Representante = fra.ID_Representante.String
	f.RepNome = fra.RepNome.String

	f.ID_UsuarioMaster = fra.ID_UsuarioMaster.String
	f.UserDados.ID_Usuario = fra.UserDados.IdUsuario.String
	f.UserDados.ID_Vinculo = fra.UserDados.IdVinculo.String
	f.UserDados.Nome = fra.UserDados.Nome.String
	f.UserDados.Nick = fra.UserDados.Nick.String
	f.UserDados.Email1 = fra.UserDados.Email1.String
	f.UserDados.Email2 = fra.UserDados.Email2.String
	f.UserDados.Senha = "" // Limpa a senha para não retornar -> fra.UserDados.Senha.String
	f.UserDados.Telefone1 = fra.UserDados.Telefone1.String
	f.UserDados.Telefone2 = fra.UserDados.Telefone2.String
	f.UserDados.UsuarioTeminal = fra.UserDados.UsuarioTeminal.String
	f.UserDados.Master = fra.UserDados.Master.String

	if fra.UserDados.DataCadastro.Valid {
		f.UserDados.DataCadastro = fra.UserDados.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		f.UserDados.DataCadastro = ""
	}

	f.ID_Franqueado = fra.ID_Franqueado.String
	f.ID_Pacote = fra.ID_Pacote.String
	f.RazaoSocial = fra.RazaoSocial.String
	f.NomeFantasia = fra.NomeFantasia.String
	f.Cnpj = fra.Cnpj.String
	f.InscricaoEstadual = fra.InscricaoEstadual.String
	f.Cep = fra.Cep.String
	f.Endereco = fra.Endereco.String
	f.Complemento = fra.Complemento.String
	f.Bairro = fra.Bairro.String
	f.Cidade = fra.Cidade.String
	f.Uf = fra.Uf.String

	if fra.DataCadastro.Valid {
		f.DataCadastro = fra.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		f.DataCadastro = ""
	}

	if fra.DataCancelamento.Valid {
		f.DataCancelamento = fra.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		f.DataCancelamento = ""
	}

	f.CodBenuvem = fra.CodBenuvem.String
	f.Observacao = fra.Observacao.String
}

func (f *Franqueado) gerarCodBeNuvem() string {

	agora := time.Now()
	parte1 := agora.Format("05")
	parte2 := rand.Intn(999) + 1
	return fmt.Sprintf("%05s", parte1+strconv.Itoa(parte2))
}

func (f *Franqueado) getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			franqueado.ID_Representante,
			representante.RazaoSocial, 
			
			franqueado.ID_Franqueado, 
			franqueado.ID_UsuarioMaster, 
			franqueado.ID_Pacote, 
			franqueado.RazaoSocial, 
			franqueado.NomeFantasia, 
			franqueado.Cnpj, 
			franqueado.InscricaoEstadual, 
			franqueado.Cep, 
			franqueado.Endereco, 
			franqueado.Complemento, 
			franqueado.Bairro, 
			franqueado.Cidade, 
			franqueado.Uf,  
			franqueado.DataCadastro,
			franqueado.DataCancelamento,
			franqueado.CodBenuvem,
			franqueado.Observacao,
			
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
			usuarios.Master,
			usuarios.DataCadastro,
			
			envio.Email,
			envio.Sms,

			bloqUsu.ID_Alvo AS blUse,
			
			bloqFra.ID_Alvo AS blFra,

			bloqRep.ID_Alvo AS blRep
			
		FROM franqueado

		LEFT JOIN representante
		ON franqueado.ID_Representante = representante.ID_Representante

		LEFT JOIN usuarios
		ON franqueado.ID_UsuarioMaster = usuarios.ID_Usuario
		

		LEFT JOIN listaBloqueio AS bloqUsu
		ON usuarios.ID_Usuario = bloqUsu.ID_Alvo   

		LEFT JOIN listaBloqueio AS bloqFra
		ON franqueado.ID_Franqueado = bloqFra.ID_Alvo   

		LEFT JOIN listaBloqueio AS bloqRep
		ON franqueado.ID_Representante = bloqRep.ID_Alvo  
		
		LEFT JOIN listaEnvio AS envio 
		ON franqueado.ID_Franqueado = envio.ID_Alvo

		WHERE franqueado.visivel = 'S'
		
		%s
	`, filtro)
}

func (f *Franqueado) processaItem(tab *sql.Rows) error {
	var (
		tmp        SFranqueado
		emailEnvio sql.NullString
		smsEnvio   sql.NullString
		bloqFra    sql.NullString
		bloqRep    sql.NullString
		bloqUse    sql.NullString
	)

	if err := tab.Scan(
		&tmp.ID_Representante,
		&tmp.RepNome,

		&tmp.ID_Franqueado,
		&tmp.ID_UsuarioMaster,
		&tmp.ID_Pacote,
		&tmp.RazaoSocial,
		&tmp.NomeFantasia,
		&tmp.Cnpj,
		&tmp.InscricaoEstadual,
		&tmp.Cep,
		&tmp.Endereco,
		&tmp.Complemento,
		&tmp.Bairro,
		&tmp.Cidade,
		&tmp.Uf,
		&tmp.DataCadastro,
		&tmp.DataCancelamento,
		&tmp.CodBenuvem,
		&tmp.Observacao,

		&tmp.UserDados.IdUsuario,
		&tmp.UserDados.IdVinculo,
		&tmp.UserDados.Nome,
		&tmp.UserDados.Nick,
		&tmp.UserDados.Email1,
		&tmp.UserDados.Email2,
		&tmp.UserDados.Senha,
		&tmp.UserDados.Telefone1,
		&tmp.UserDados.Telefone2,
		&tmp.UserDados.UsuarioTeminal,
		&tmp.UserDados.Master,
		&tmp.UserDados.DataCadastro,

		&emailEnvio,
		&smsEnvio,

		&bloqUse,
		&bloqFra,
		&bloqRep,
	); err != nil {
		return err
	}

	f.sFranqueadoToFranqueado(tmp)

	// Verifica se o envio de email esta liberado
	if emailEnvio.String == "" {
		f.EmailEnvio = "N"
	} else {
		f.EmailEnvio = emailEnvio.String
	}

	// Verifica se o envio de sms esta liberado
	if smsEnvio.String == "" {
		f.SmsEnvio = "N"
	} else {
		f.SmsEnvio = smsEnvio.String
	}

	// Pega informação se o usuario master esta bloqueado
	if bloqUse.Valid {
		f.UserDados.Ativo = "N"
	} else {
		f.UserDados.Ativo = "S"
	}

	// Pega informação se o rep ou fra esta bloqueado
	if bloqFra.Valid {
		f.Ativo = "N"
	} else if bloqRep.Valid {
		f.Ativo = "N"
	} else {
		f.Ativo = "S"
	}

	return nil
}
