package representanteV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/V4/config"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	listaEnvioV4 "api/src/V4/modulos/listaEnvio"
	pacotev4 "api/src/V4/modulos/pacote"
	ticketV4 "api/src/V4/modulos/ticket"
	usuariosV4 "api/src/V4/modulos/usuarios"
	"api/src/V4/seguranca"
	tiposV4 "api/src/V4/tipos"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type representanteLogin struct {
	Token string `json:"token"`
	usuariosV4.Usuario
}

type Representante struct {
	tiposV4.TRepresentante

	EmailEnvio string             `json:"emailEnvio"`
	SmsEnvio   string             `json:"smsEnvio"`
	User       usuariosV4.Usuario `json:"userMaster,omitempty"`
	// ListarTodasCentrais: break-glass lista todas as centrais (nao persistido)
	ListarTodasCentrais bool `json:"listarTodasCentrais"`
}

// logar loga o representante
func (rl *representanteLogin) logar(email, senha string) error {

	// Valida se um email foi informado
	if email == "" {
		return errors.New("um email deve ser informado")

	}

	// Valida se uma senha foi informada
	if senha == "" {
		return errors.New("uma senha deve ser informada")

	}

	// Busca o dados do usuario ===============================================
	var err error
	rl.Email1 = email
	if err := rl.GetDadosByEmail1(); err != nil {
		return err
	}

	// Verifica se a esta correta senha =======================================
	if err = seguranca.VerificarSenha(rl.Senha, senha); err != nil {
		return err
	}

	// Verifica se usuario esta bloqueado =====================================
	if rl.Ativo == "N" {
		return errors.New("usuario bloqueado")
	}

	// Verifica se o representante esta na lista de bloqueio ==================
	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = rl.ID_Vinculo
	if err := lb.GetBloquadoByIdAlvo(); err != nil {
		return err
	}
	if lb.Ativo == "S" {
		return errors.New("representante na lista de bloqueio bloqueado")
	}

	// Caso esteja tudo ok até aqui ele cria um token =========================
	rl.Token, err = seguranca.CriarToken(rl.ID_Usuario)
	if err != nil {
		return err
	}

	// Retorna os dados para o requisitante
	return nil
}

// Insere insere um novo representante
func (r *Representante) Insere() error {

	// Gerea um id caso não venha um na requisição
	if r.ID_Representante == "" {
		r.ID_Representante = auxiliar.GeradorDeId()
	}

	r.ID_UsuarioMaster = auxiliar.GeradorDeId()

	r.User.ID_Usuario = r.ID_UsuarioMaster
	r.User.ID_Vinculo = r.ID_Representante
	r.User.Tipo = "REP"
	r.User.UsuarioWeb = "S"
	r.User.UsuarioTeminal = "S"
	r.User.Master = "S"
	r.User.AdmFinanceiro = "S"

	r.IDCentralUUID = auxiliar.GarantirIDCentralUUID(r.IDCentralUUID)
	r.User.IDCentralUUID = r.IDCentralUUID

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// cria a consulta
	stm, err := db.Prepare(`
		INSERT INTO representante(
			representante.ID_Representante,
			representante.IDCentralUUID,
			representante.ID_UsuarioMaster, 
			representante.ID_Pacote, 
			representante.RazaoSocial, 
			representante.NomeFantasia, 
			representante.Cnpj, 
			representante.InscricaoEstadual, 
			representante.Cep, 
			representante.Endereco, 
			representante.Complemento, 
			representante.Bairro, 
			representante.Cidade, 
			representante.Uf,
			representante.Informativo
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if r.ID_Pacote == "" {
		r.Informativo = config.RepInformativo
	}

	// Executa a consulta
	if _, err := stm.Exec(
		strings.ToUpper(r.ID_Representante),
		r.IDCentralUUID,
		strings.ToUpper(r.ID_UsuarioMaster),
		strings.ToUpper(r.ID_Pacote),
		strings.ToUpper(r.RazaoSocial),
		strings.ToUpper(r.NomeFantasia),
		r.Cnpj,
		r.InscricaoEstadual,
		r.Cep,
		strings.ToUpper(r.Endereco),
		strings.ToUpper(r.Complemento),
		strings.ToUpper(r.Bairro),
		strings.ToUpper(r.Cidade),
		strings.ToUpper(r.Uf),
		strings.ToUpper(r.Informativo),
	); err != nil {
		return err
	}

	// Cria o usuario master
	if err := r.User.Insere(); err != nil {
		// Caso não consiga criar o usuario master ele deleta o representante
		if err := r.DeletaById(); err != nil {
			return err
		}
		return err
	}

	// Insere o representante na lista de envio
	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = r.ID_Representante

	if err := le.Insere(); err != nil {
		return err
	}
	return nil
}

// GetDadosById busca os dados do representante
func (r *Representante) GetDadosById() error {
	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE representante.ID_Representante = '%s'
		AND representante.visivel = 'S'
	`,
		r.ID_Representante,
	)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, r); err != nil {
			return err
		}
		return nil
	}

	return errors.New("representante não encontrado na base de dados")
}

func (r *Representante) GetDadosFullById() error {
	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE representante.ID_Representante = '%s'
		AND representante.visivel = 'S'
		`, r.ID_Representante)

	tab, err := db.Query(getSelect(filtro))

	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, r); err != nil {
			return err
		}

		var use usuariosV4.Usuario
		use.ID_Usuario = r.ID_UsuarioMaster
		if err := use.GetDadosById(); err != nil {
			return err
		}

		r.User = use
		r.User.Senha = ""
		return nil
	}

	return errors.New("representante não encontrado na base de dados")
}

// GetIdPacoteById busca o id do pacote ao qual o representate faz parte pelo id
// do representante
func (r *Representante) GetIdPacoteById() error {
	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria a consulta
	tab, err := db.Query(`
		SELECT representante.ID_Pacote
		FROM representante
		WHERE representante.ID_Representante
	`, r.ID_Representante)
	if err != nil {
		return err
	}
	defer tab.Close()

	// Verifica se o representante foi encontrado
	if tab.Next() {
		if err = tab.Scan(&r.ID_Pacote); err != nil {
			return err
		}
		return nil
	}
	return errors.New("representante não encontrado na base de dados")
}

// AlteraById altera os dados representante pelo seu id juntamente com dados
// do usuario master
func (r *Representante) AlteraById() error {

	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE representante 
		SET 
			representante.ID_Pacote = ?,
			representante.RazaoSocial = ?,
			representante.NomeFantasia = ?,
			representante.Cnpj = ?,
			representante.InscricaoEstadual = ?,
			representante.Cep = ?,
			representante.Endereco = ?,
			representante.Complemento = ?,
			representante.Bairro = ?,
			representante.Cidade = ?,
			representante.Uf = ?,
			representante.Informativo = ?,
			representante.IDCentralUUID = COALESCE(NULLIF(?, ''), representante.IDCentralUUID)
		WHERE representante.ID_Representante = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	// Executa a alteração do representante
	if _, err := stm.Exec(
		strings.ToUpper(r.ID_Pacote),
		strings.ToUpper(r.RazaoSocial),
		strings.ToUpper(r.NomeFantasia),
		r.Cnpj,
		r.InscricaoEstadual,
		r.Cep,
		strings.ToUpper(r.Endereco),
		strings.ToUpper(r.Complemento),
		strings.ToUpper(r.Bairro),
		strings.ToUpper(r.Cidade),
		strings.ToUpper(r.Uf),
		strings.ToUpper(r.Informativo),
		r.IDCentralUUID,
		r.ID_Representante,
	); err != nil {
		return err
	}

	// Altera os dados do Usuario master
	if err := r.User.AlteraById(); err != nil {
		return err
	}

	return nil
}

func (r *Representante) CancelaById() error {
	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria uma consulta de exclusao
	stm, err := db.Prepare(`
		UPDATE representante 
		SET representante.DataCancelamento = ?
		WHERE representante.ID_Representante = ?
	`)
	if err != nil {
		return nil
	}
	defer stm.Close()

	data := time.Now().Format("2006-01-02 15:04:05")
	// Executa a exclusao
	if _, err := stm.Exec(data, r.ID_Representante); err != nil {
		return err
	}
	//=========================================================================

	return nil
}

func (r *Representante) ReverteCancelaById() error {
	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria uma consulta de exclusao
	stm, err := db.Prepare(`
		UPDATE representante 
		SET representante.DataCancelamento = ?
		WHERE representante.ID_Representante = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	var nulo sql.NullTime
	// Executa a exclusao
	if _, err := stm.Exec(nulo, r.ID_Representante); err != nil {
		return err
	}
	//=========================================================================

	return nil
}

// DeletaById apaga um representante pelo seu id
func (r *Representante) DeletaById() error {
	// Valida se um id de representante foi informado
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	// Deleta os usuarios do representante ====================================
	//
	// Carrega o id no vinculo
	r.User.ID_Vinculo = r.ID_Representante

	// Deleta os Franqueados do representante ==================================
	//

	// Deleta os usuarios do representante
	if err := r.User.DeletaAllByVinculo(); err != nil {
		return err
	}
	//=========================================================================

	// Deleta os pacotes do representante =====================================
	//
	// Cria objeto pacote
	var p pacotev4.Pacote

	//Carrega o parametro idvinculo do pacote com id do representante
	p.ID_Vinculo = r.ID_Representante

	// Apaga todos os usuarios vinculado ao representante
	if err := p.DeletaAllByVinculo(); err != nil {
		return err
	}
	//=========================================================================

	// Deleta os tickets do representante =====================================
	//
	// Cria objeto ticket
	var t ticketV4.Ticket

	// Carrega o id master do objeto com o id do representante
	t.ID_Master = r.ID_Representante
	if err := t.DeletaAllByMster(); err != nil {
		return err
	}
	//=========================================================================

	// Deleta lista envio =====================================================
	//
	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = r.ID_Representante
	if err := le.DeleteByIdAlvo(); err != nil {
		return err
	}

	// Deleta lista bloqueio ==================================================
	//
	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = r.ID_Representante
	if err := lb.DeleteByIdAlvo(); err != nil {
		return err
	}

	// Deleta o representante =================================================
	//
	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Cria uma consulta de exclusao
	stm, err := db.Prepare(`
		DELETE FROM representante WHERE representante.ID_Representante = ?
	`)
	if err != nil {
		return nil
	}
	defer stm.Close()

	// Executa a exclusao
	if _, err := stm.Exec(r.ID_Representante); err != nil {
		return err
	}
	//=========================================================================

	return nil
}

// Funcoes para manipular se o representante esta ativo ou nao ================
func (r *Representante) GetAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = r.ID_Representante

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	r.Ativo = lb.Ativo
	return nil
}

func (r *Representante) SetAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	if r.Ativo == "" {
		return errors.New("um estado de bloqueado deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = r.ID_Representante
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = r.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (r *Representante) InverteAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = r.ID_Representante
	lb.Descricao = "ALTERADO VIA WEB"

	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	r.Ativo = lb.Ativo
	return nil
}

// Funcoes para manipular o envio de email ====================================
func (r *Representante) GetEmailAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = r.ID_Representante

	if err := le.GetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	r.EmailEnvio = le.Email
	return nil
}

func (r *Representante) SetEmailAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	if r.EmailEnvio == "" {
		return errors.New("um estado de bloqueado deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = r.ID_Representante
	le.Email = r.EmailEnvio

	if err := le.SetEmailAtivoByIdAlvo(); err != nil {
		return err
	}

	return nil
}

func (r *Representante) InverteEmailAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	if err := r.GetEmailAtivoById(); err != nil {
		return err
	}

	if err := auxiliar.InverteEstado(&r.EmailEnvio); err != nil {
		return err
	}

	if err := r.SetEmailAtivoById(); err != nil {
		return err
	}

	return nil
}

// Funcoes para manipular se o representante esta ativo ou nao ================
func (r *Representante) GetSmsAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = r.ID_Representante
	if err := le.GetSmsAtivoByIdAlvo(); err != nil {
		return err
	}
	r.SmsEnvio = le.Sms
	return nil
}

func (r *Representante) SetSmsAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	if r.SmsEnvio == "" {
		return errors.New("um estado de bloqueado deve ser informado")
	}

	var le listaEnvioV4.ListaEnvio
	le.ID_Alvo = r.ID_Representante
	le.Sms = r.SmsEnvio
	if err := le.SetSmsAtivoByIdAlvo(); err != nil {
		return err
	}

	return nil
}

func (r *Representante) InverteSmsAtivoById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}

	if err := r.GetSmsAtivoById(); err != nil {
		return err
	}

	if err := auxiliar.InverteEstado(&r.SmsEnvio); err != nil {
		return err
	}

	if err := r.SetSmsAtivoById(); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// funcoes para listar os representantes ======================================

// Listar lista todos os franqueados sem os usuarios master
func (r *Representante) Listar(lista *[]Representante) error {
	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := `
		WHERE representante.visivel = 'S'
	`
	args := []interface{}{}
	if err := auxiliar.ExigirFiltroCentralUUID(r.IDCentralUUID, r.ListarTodasCentrais); err != nil {
		return err
	}
	if !r.ListarTodasCentrais {
		filtro += ` AND representante.IDCentralUUID = ?`
		args = append(args, r.IDCentralUUID)
	}
	filtro += ` ORDER BY representante.RazaoSocial`

	tab, err := db.Query(getSelect(filtro), args...)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Representante

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)

	}
	return nil
}

// ListarCentral lista representantes do portal Central com regra hibrida de tenant.
// Central legada (ID_Central = CENTRAL): representantes da central legada sem exigir match exato em IDCentralUUID.
// Central nova: representante.IDCentralUUID = ID_Central da central selecionada.
func (r *Representante) ListarCentral(lista *[]Representante) error {
	r.IDCentralUUID = strings.TrimSpace(r.IDCentralUUID)

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := `
		WHERE representante.visivel = 'S'
	`
	args := []interface{}{}

	if r.ListarTodasCentrais {
		// break-glass: todas as centrais
	} else if err := auxiliar.ExigirFiltroCentralUUID(r.IDCentralUUID, false); err != nil {
		return err
	} else if r.IDCentralUUID == auxiliar.IDCentralPadrao {
		filtro += ` AND (
			representante.IDCentralUUID IS NULL
			OR TRIM(representante.IDCentralUUID) = ''
			OR TRIM(representante.IDCentralUUID) = ?
			OR NOT EXISTS (
				SELECT 1 FROM central c
				WHERE c.ID_Central <> ?
				AND c.ID_Central = TRIM(representante.IDCentralUUID)
			)
		)`
		args = append(args, auxiliar.IDCentralPadrao, auxiliar.IDCentralPadrao)
	} else {
		filtro += ` AND TRIM(representante.IDCentralUUID) = ?`
		args = append(args, r.IDCentralUUID)
	}
	filtro += ` ORDER BY representante.RazaoSocial`

	tab, err := db.Query(getSelect(filtro), args...)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Representante
		if err := processaItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}
	return nil
}

// ListarCentralPorReferencia resolve idCentralUUID (hex ou ID_Central) via tabela central
// e lista representantes com a mesma regra hibrida de ListarCentral.
func (r *Representante) ListarCentralPorReferencia(lista *[]Representante) error {
	r.IDCentralUUID = strings.TrimSpace(r.IDCentralUUID)

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	if !r.ListarTodasCentrais {
		if err := auxiliar.ExigirFiltroCentralUUID(r.IDCentralUUID, false); err != nil {
			return err
		}
		r.IDCentralUUID = auxiliar.ResolverIDCentralPorReferencia(db, r.IDCentralUUID)
	}

	filtro := `
		WHERE representante.visivel = 'S'
	`
	args := []interface{}{}

	if r.ListarTodasCentrais {
		// break-glass: todas as centrais
	} else if r.IDCentralUUID == auxiliar.IDCentralPadrao {
		filtro += ` AND (
			representante.IDCentralUUID IS NULL
			OR TRIM(representante.IDCentralUUID) = ''
			OR TRIM(representante.IDCentralUUID) = ?
			OR NOT EXISTS (
				SELECT 1 FROM central c
				WHERE c.ID_Central <> ?
				AND c.ID_Central = TRIM(representante.IDCentralUUID)
			)
		)`
		args = append(args, auxiliar.IDCentralPadrao, auxiliar.IDCentralPadrao)
	} else {
		filtro += ` AND TRIM(representante.IDCentralUUID) = ?`
		args = append(args, r.IDCentralUUID)
	}
	filtro += ` ORDER BY representante.RazaoSocial`

	tab, err := db.Query(getSelect(filtro), args...)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Representante
		if err := processaItem(tab, &item); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}
	return nil
}

// ListarFull lista todos os franqueados com seu respectivos usuarios master
func (r *Representante) ListarFull(lista *[]Representante) error {
	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(getSelect(`
		WHERE representante.visivel = 'S'
		ORDER BY representante.RazaoSocial
	`))
	if err != nil {
		return err
	}
	defer tab.Close()
	
	for tab.Next() {

		var item Representante
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		item.User.ID_Usuario = item.ID_UsuarioMaster
		if err := item.User.GetDadosById(); err != nil {
			return err
		}
		*lista = append(*lista, item)

	}
	return nil
}

// ListarToAtivos lista todos os representantes ativos
func (r *Representante) ListarToAtivos(lista *[]Representante) error {
	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := `
		WHERE representante.Ativo = 'S'
		AND representante.visivel = 'S'
	`

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Representante
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		// So adiciona na lista caso esteja ativo
		if item.Ativo == "S" {
			*lista = append(*lista, item)
		}

	}
	return nil
}

// ListarToBloqueado lista todos o representantes bloqueado
func (r *Representante) ListarToDesativado(lista *[]Representante) error {
	// Cria uma conexão com o banco
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := `
		WHERE representante.Ativo = 'N'
		AND representante.visivel = 'S'
	`
	args := []interface{}{}
	if err := auxiliar.ExigirFiltroCentralUUID(r.IDCentralUUID, r.ListarTodasCentrais); err != nil {
		return err
	}
	if !r.ListarTodasCentrais {
		filtro += ` AND representante.IDCentralUUID = ?`
		args = append(args, r.IDCentralUUID)
	}

	tab, err := db.Query(getSelect(filtro), args...)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Representante
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		if item.Ativo == "N" {
			*lista = append(*lista, item)
		}

	}
	return nil
}

//=============================================================================
// Funcoes internas ===========================================================

// SRepresentanteToRepresentante converte tipo SRepresentante para Representante
func (r *Representante) sRepresentanteToRepresentante(sr tiposV4.SRepresentante) {
	r.ID_Representante = sr.ID_Representante.String
	r.IDCentralUUID = sr.IDCentralUUID.String
	r.ID_UsuarioMaster = sr.ID_UsuarioMaster.String
	r.ID_Pacote = sr.ID_Pacote.String
	r.RazaoSocial = sr.RazaoSocial.String
	r.NomeFantasia = sr.NomeFantasia.String
	r.Cnpj = sr.Cnpj.String
	r.InscricaoEstadual = sr.InscricaoEstadual.String
	r.Cep = sr.Cep.String
	r.Endereco = sr.Endereco.String
	r.Complemento = sr.Complemento.String
	r.Bairro = sr.Bairro.String
	r.Cidade = sr.Cidade.String
	r.Uf = sr.Uf.String

	if sr.DataCadastro.Valid {
		r.DataCadastro = sr.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		r.DataCadastro = ""
	}

	if sr.DataCancelamento.Valid {
		r.DataCancelamento = sr.DataCancelamento.Time.Format("02/01/2006 15:04:05")
	} else {
		r.DataCancelamento = ""
	}

	r.Informativo = sr.Informativo.String
	r.UsaAdmConfmonit = strings.ToUpper(strings.TrimSpace(sr.UsaAdmConfmonit.String))
	if r.UsaAdmConfmonit != "S" {
		r.UsaAdmConfmonit = "N"
	}
}

// SetUsaAdmConfmonitById grava se o REP opera o admConfmonit (S/N).
func (r *Representante) SetUsaAdmConfmonitById() error {
	if r.ID_Representante == "" {
		return errors.New("um id de representante deve ser informado")
	}
	flag := strings.ToUpper(strings.TrimSpace(r.UsaAdmConfmonit))
	if flag != "S" && flag != "N" {
		return errors.New("usaAdmConfmonit deve ser S ou N")
	}
	r.UsaAdmConfmonit = flag

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE representante
		SET representante.UsaAdmConfmonit = ?
		WHERE representante.ID_Representante = ?
	`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "usaadmconfmonit") {
			return errors.New("coluna UsaAdmConfmonit ausente — rode o ALTER TABLE no MySQL")
		}
		return err
	}
	defer stm.Close()

	res, err := stm.Exec(flag, r.ID_Representante)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "usaadmconfmonit") {
			return errors.New("coluna UsaAdmConfmonit ausente — rode o ALTER TABLE no MySQL")
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("representante nao encontrado")
	}
	return nil
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT
			representante.ID_Representante,
			representante.IDCentralUUID,
			representante.ID_UsuarioMaster,
			representante.ID_Pacote,
			representante.RazaoSocial,
			representante.NomeFantasia,
			representante.Cnpj,
			representante.InscricaoEstadual,
			representante.Cep,
			representante.Endereco,
			representante.Complemento,
			representante.Bairro,
			representante.Cidade,
			representante.Uf,
			representante.DataCadastro,
			representante.DataCancelamento,
			representante.Informativo,
			COALESCE(representante.UsaAdmConfmonit, 'N') AS UsaAdmConfmonit,

			listaBloqueio.ID_Alvo,
			
			listaEnvio.Email,
			listaEnvio.Sms

		FROM representante
		
		LEFT JOIN listaBloqueio
		ON representante.ID_Representante = listaBloqueio.ID_Alvo
		AND (
			listaBloqueio.DataRetirada IS NULL
			OR listaBloqueio.DataRetirada <= NOW()
		)
		
		LEFT JOIN listaEnvio
		ON representante.ID_Representante = listaEnvio.ID_Alvo
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, item *Representante) error {
	var (
		tmp     tiposV4.SRepresentante
		bloqRep sql.NullString
		email   sql.NullString
		sms     sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Representante,
		&tmp.IDCentralUUID,
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
		&tmp.Informativo,
		&tmp.UsaAdmConfmonit,

		&bloqRep,
		&email,
		&sms,
	); err != nil {
		return err
	}

	item.sRepresentanteToRepresentante(tmp)

	if bloqRep.Valid {
		item.Ativo = "N"
	} else {
		item.Ativo = "S"
	}

	if email.String == "" {
		item.EmailEnvio = "N"
	} else {
		item.EmailEnvio = email.String
	}

	if sms.String == "" {
		item.SmsEnvio = "N"
	} else {
		item.SmsEnvio = sms.String
	}

	return nil
}
