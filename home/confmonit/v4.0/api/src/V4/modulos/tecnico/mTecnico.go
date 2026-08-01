package tecnicoV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"api/src/V4/seguranca"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type Tecnico struct {
	ID_Tecnico  string `json:"idTecnico"`
	ID_Vinculo  string `json:"idVinculo"`
	Nome        string `json:"nome"`
	Nick        string `json:"nick"`
	Cpf         string `json:"cpf"`
	Rg          string `json:"rg"`
	Endereco    string `json:"endereco"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Cidade      string `json:"cidade"`
	Uf          string `json:"uf"`
	Cep         string `json:"cep"`
	Email       string `json:"email"`
	Senha       string `json:"senha,omitempty"`
	Telefone1   string `json:"telefone1"`
	Telefone2   string `json:"telefone2"`
	Ativo       string `json:"ativo"`
	RepId       string `json:"repId"`
	RepRazao    string `json:"repRazao"`
	RepAtivo    string `json:"repAtivo"`
	FraId       string `json:"fraId"`
	FraRazao    string `json:"fraRazao"`
	FraAtivo    string `json:"fraAtivo"`
	EnviaEmail  string `json:"enviaEmail"`
	EnviaSms    string `json:"enviaSMS"`

	// auxiliar
	Token string `json:"token,omitempty"`
}

type STecnico struct {
	ID_Tecnico,
	ID_Vinculo sql.NullString
	Nome        sql.NullString
	Nick        sql.NullString
	Cpf         sql.NullString
	Rg          sql.NullString
	Endereco    sql.NullString
	Complemento sql.NullString
	Bairro      sql.NullString
	Cidade      sql.NullString
	Uf          sql.NullString
	Cep         sql.NullString
	Email       sql.NullString
	Senha       sql.NullString
	Telefone1   sql.NullString
	Telefone2   sql.NullString

	RepId    sql.NullString
	RepRazao sql.NullString

	FraId    sql.NullString
	FraRazao sql.NullString
}

func (t *Tecnico) logar(email, senha string) error {

	// Valida se um email foi informado
	if email == "" {
		return errors.New("um email deve ser informado")

	}

	// Valida se uma senha foi informada
	if senha == "" {
		return errors.New("uma senha deve ser informada")

	}

	var err error
	t.Email = email
	if err := t.GetDadosByEmail(); err != nil {
		return err
	}
	
	// Valida a senha
	if err = seguranca.VerificarSenha(t.Senha, senha); err != nil {
		return err
	}

	// Valida se usuario nao esta bloqueado
	if t.Ativo == "S" {
		// Caso a senha esteja correta ele cria um token
		t.Token, err = seguranca.CriarToken(t.ID_Tecnico)
		if err != nil {
			return err
		}
	} else {
		return errors.New("usuario bloqueado")
	}

	// Retorna os dados para o requisitante
	return nil
}

func (t *Tecnico) Insere() error {
	if t.ID_Tecnico == "" {
		t.ID_Tecnico = auxiliar.GeradorDeId()
	}

	if t.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	senha, err := seguranca.HashString("usuario123")
	if err != nil {
		return err
	}

	t.Senha = senha

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO tecnico(
			tecnico.ID_Tecnico, 
			tecnico.ID_Vinculo, 
			tecnico.Nome, 
			tecnico.Nick, 
			tecnico.Cpf, 
			tecnico.Rg, 
			tecnico.Endereco, 
			tecnico.Complemento, 
			tecnico.Bairro, 
			tecnico.Cidade, 
			tecnico.Uf, 
			tecnico.Cep, 
			tecnico.Email, 
			tecnico.Senha, 
			tecnico.Telefone1, 
			tecnico.Telefone2
		) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		t.ID_Tecnico,
		t.ID_Vinculo,
		t.Nome,
		t.Nick,
		t.Cpf,
		t.Rg,
		t.Endereco,
		t.Complemento,
		t.Bairro,
		t.Cidade,
		t.Uf,
		t.Cep,
		t.Email,
		t.Senha,
		t.Telefone1,
		t.Telefone2,
	); err != nil {
		return err
	}

	return nil
}

func (t *Tecnico) GetDadosById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE tecnico.ID_Tecnico = '%s'`, t.ID_Tecnico)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, t); err != nil {
			return err
		}
		return nil
	}
	return errors.New("tecnico não encontrado na base de dados")
}

func (t *Tecnico) GetDadosByEmail() error {
	if t.Email == "" {
		return errors.New("um email de tecnico deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE tecnico.Email = '%s'`, t.Email)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, t); err != nil {
			return err
		}
		return nil
	}
	return errors.New("tecnico não encontrado na base de dados")
}

func (t *Tecnico) AlteraById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE tecnico SET 
			tecnico.Nome = ?,
			tecnico.Nick = ?,
			tecnico.Cpf = ?,
			tecnico.Rg = ?,

			tecnico.Endereco = ?, 
			tecnico.Complemento = ?,
			tecnico.Bairro = ?,
			tecnico.Cidade = ?,
			tecnico.Uf = ?,
			tecnico.Cep = ?,
			tecnico.Email = ?,
			
			tecnico.Telefone1 = ?,
			tecnico.Telefone2 = ?
		WHERE tecnico.ID_Tecnico = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		t.Nome,
		t.Nick,
		t.Cpf,
		t.Rg,

		t.Endereco,
		t.Complemento,
		t.Bairro,
		t.Cidade,
		t.Uf,
		t.Cep,
		t.Email,

		t.Telefone1,
		t.Telefone2,
		t.ID_Tecnico,
	); err != nil {
		return err
	}

	return nil
}

func (t *Tecnico) DeletaById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM tecnico WHERE tecnico.ID_Tecnico = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(t.ID_Tecnico); err != nil {
		return err
	}

	return nil
}

func (t *Tecnico) DeletaAllByIdVinculo() error {
	if t.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM tecnico WHERE.ID_Vinculo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(t.ID_Vinculo); err != nil {
		return err
	}

	return nil
}

func (t *Tecnico) ListaByIdVinculo(lista *[]Tecnico) error {

	if t.ID_Vinculo == "" {
		return errors.New("um id de vinculo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE tecnico.ID_Vinculo = '%s'
	`, t.ID_Vinculo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Tecnico

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// Funcoes para manipular a senha =============================================
func (t *Tecnico) ResetSenhaById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	senha, err := seguranca.HashString("usuario123")
	if err != nil {
		return err
	}

	t.Senha = senha
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE tecnico 
		SET tecnico.Senha = ?
		WHERE tecnico.ID_Tecnico = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()
	if _, err := stm.Exec(t.Senha, t.ID_Tecnico); err != nil {
		return err
	}

	return nil
}

func (t *Tecnico) AlteraSenhaById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	if t.Senha == "" {
		return errors.New("uma senha para tecnico deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE tecnico 
		SET tecnico.Senha = ?
		WHERE tecnico.ID_Tecnico = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(t.Senha, t.ID_Tecnico); err != nil {
		return err
	}

	return nil
}

// Funcoes pra maipular o campo Ativo =========================================
func (t *Tecnico) GetAtivoById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = t.ID_Tecnico

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	t.Ativo = lb.Ativo

	return nil
}

func (t *Tecnico) SetAtivoById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	if t.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = t.ID_Tecnico
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = t.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}
	return nil
}

func (t *Tecnico) InverteAtivoById() error {
	if t.ID_Tecnico == "" {
		return errors.New("um id de tecnico deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = t.ID_Tecnico
	lb.Descricao = "ALTERADO VIA WEB"
	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	t.Ativo = lb.Ativo
	return nil
}

// funcoes internas ===========================================================

func sTecnicoToTecnico(st STecnico) (t Tecnico) {
	t.ID_Tecnico = st.ID_Tecnico.String
	t.ID_Vinculo = st.ID_Vinculo.String
	t.Nome = st.Nome.String
	t.Nick = st.Nick.String
	t.Cpf = st.Cpf.String
	t.Rg = st.Rg.String
	t.Endereco = st.Endereco.String
	t.Complemento = st.Complemento.String
	t.Bairro = st.Bairro.String
	t.Cidade = st.Cidade.String
	t.Uf = st.Uf.String
	t.Cep = st.Cep.String
	t.Email = st.Email.String
	t.Senha = st.Senha.String
	t.Telefone1 = st.Telefone1.String
	t.Telefone2 = st.Telefone2.String

	// Dados do franqueado
	t.RepId = st.RepId.String
	t.RepRazao = st.RepRazao.String

	// Dados do Representante
	t.FraId = st.FraId.String
	t.FraRazao = st.FraRazao.String

	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			tecnico.ID_Tecnico, 
			tecnico.ID_Vinculo, 
			tecnico.Nome, 
			tecnico.Nick, 
			tecnico.Cpf, 
			tecnico.Rg, 
			tecnico.Endereco, 
			tecnico.Complemento, 
			tecnico.Bairro, 
			tecnico.Cidade, 
			tecnico.Uf, 
			tecnico.Cep, 
			tecnico.Email, 
			tecnico.Senha, 
			tecnico.Telefone1, 
			tecnico.Telefone2, 
			
			representante.ID_Representante,
			representante.RazaoSocial,
			

			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,

			blTec.ID_Alvo,
			blFra.ID_Alvo,
			blRep.ID_Alvo,

			listaEnvio.Email,
			listaEnvio.Sms

		FROM tecnico 
		
		LEFT JOIN franqueado
		ON  tecnico.ID_Vinculo = franqueado.ID_Franqueado

		LEFT JOIN representante
		ON  franqueado.ID_Representante = representante.ID_Representante

		LEFT JOIN listaBloqueio AS blTec
		ON  tecnico.ID_Tecnico = blTec.ID_Alvo

		LEFT JOIN listaBloqueio AS blFra
		ON  franqueado.ID_Franqueado = blFra.ID_Alvo

		LEFT JOIN listaBloqueio AS blRep
		ON  representante.ID_Representante = blRep.ID_Alvo

		LEFT JOIN  listaEnvio
		ON tecnico.ID_Tecnico = listaEnvio.ID_Alvo

		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, t *Tecnico) error {
	var tmp STecnico
	var (
		bloqTec    sql.NullString
		bloqFra    sql.NullString
		bloqRep    sql.NullString
		enviaEmail sql.NullString
		enviaSms   sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Tecnico,
		&tmp.ID_Vinculo,
		&tmp.Nome,
		&tmp.Nick,
		&tmp.Cpf,
		&tmp.Rg,
		&tmp.Endereco,
		&tmp.Complemento,
		&tmp.Bairro,
		&tmp.Cidade,
		&tmp.Uf,
		&tmp.Cep,
		&tmp.Email,
		&tmp.Senha,
		&tmp.Telefone1,
		&tmp.Telefone2,

		// Dados do representante
		&tmp.RepId,
		&tmp.RepRazao,

		// Dados do Franqueado
		&tmp.FraId,
		&tmp.FraRazao,

		// Dados do Bloqueio
		&bloqTec,
		&bloqFra,
		&bloqRep,

		&enviaEmail,
		&enviaSms,
	); err != nil {
		return err
	}
	*t = sTecnicoToTecnico(tmp)

	// Carrega t.FraAtivo
	if bloqFra.Valid {
		t.FraAtivo = "N"
	} else {
		t.FraAtivo = "S"
	}

	// Carrega t.RepAtivo
	if bloqRep.Valid {
		t.RepAtivo = "N"
	} else {
		t.RepAtivo = "S"
	}

	// Carrega t.Ativo
	if bloqTec.Valid {
		t.Ativo = "N"
	} else if bloqFra.Valid {
		t.Ativo = "N"
	} else if bloqRep.Valid {
		t.Ativo = "N"
	} else {
		t.Ativo = "S"
	}

	if enviaEmail.String == "" {
		t.EnviaEmail = "N"
	} else {
		t.EnviaEmail = enviaEmail.String
	}

	if enviaSms.String == "" {
		t.EnviaSms = "N"
	} else {
		t.EnviaSms = enviaSms.String
	}

	return nil
}
