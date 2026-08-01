package viaturaV4

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

type Viatura struct {
	ID_Viatura    string `json:"idViatura"`
	ID_Franqueado string `json:"idFranqueado"`
	Nome          string `json:"nome"`
	Nick          string `json:"nick"`
	Cpf           string `json:"cpf"`
	Rg            string `json:"rg"`
	Endereco      string `json:"endereco"`
	Bairro        string `json:"bairro"`
	Complemento   string `json:"complemento"`
	Cidade        string `json:"cidade"`
	Uf            string `json:"uf"`
	Cep           string `json:"cep"`
	Modelo        string `json:"modelo"`
	Placa         string `json:"placa"`
	Cor           string `json:"cor"`
	Telefone1     string `json:"telefone1"`
	Telefone2     string `json:"telefone2"`
	Telefone3     string `json:"telefone3"`
	Email         string `json:"email"`
	Senha         string `json:"senha,omitempty"`
	Observacao    string `json:"observacao"`
	DataCadastro  string `json:"dataCadastro"`
	Ativo         string `json:"ativo"`

	// Dados envio de email e sms
	EnvioEmail string `json:"envioEmail"`
	EnvioSms   string `json:"envioSms"`

	// Dados do franqueado
	FraId    string `json:"fraId"`
	FraRazao string `json:"fraRazao"`
	FraAtivo string `json:"fraAtivo"`

	// Dados do Representante
	RepId    string `json:"repId"`
	RepRazao string `json:"repRazao"`
	RepAtivo string `json:"repAtivo"`
}

type SViatura struct {
	ID_Viatura    sql.NullString
	ID_Franqueado sql.NullString
	Nome          sql.NullString
	Nick          sql.NullString
	Cpf           sql.NullString
	Rg            sql.NullString
	Endereco      sql.NullString
	Bairro        sql.NullString
	Complemento   sql.NullString
	Cidade        sql.NullString
	Uf            sql.NullString
	Cep           sql.NullString
	Modelo        sql.NullString
	Placa         sql.NullString
	Cor           sql.NullString
	Telefone1     sql.NullString
	Telefone2     sql.NullString
	Telefone3     sql.NullString
	Email         sql.NullString
	Senha         sql.NullString
	Observacao    sql.NullString
	DataCadastro  sql.NullTime

	// Dados do franqueado
	FraId    sql.NullString
	FraRazao sql.NullString

	// Dados do representante
	RepId    sql.NullString
	RepRazao sql.NullString
}

func (v *Viatura) Insere() error {
	if v.ID_Viatura == "" {
		v.ID_Viatura = auxiliar.GeradorDeId()
	}

	if v.ID_Franqueado == "" {
		return errors.New("um de franqueado deve ser informado")
	}

	if v.Nome == "" {
		return errors.New("um de viatura deve ser informado")
	}

	if v.Email == "" {
		return errors.New("um email viatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO viaturas(
			viaturas.ID_Viatura, 
			viaturas.ID_Franqueado, 
			viaturas.Nome, 
			viaturas.Nick, 
			viaturas.Cpf, 
			viaturas.Rg, 
			viaturas.Endereco, 
			viaturas.Complemento, 
			viaturas.Bairro, 
			viaturas.Cidade, 
			viaturas.Uf, 
			viaturas.Cep, 
			viaturas.Modelo, 
			viaturas.Placa, 
			viaturas.Cor, 
			viaturas.Telefone1, 
			viaturas.Telefone2, 
			viaturas.Telefone3, 
			viaturas.Email, 
			viaturas.Senha, 
			viaturas.Observacao
		) 
		VALUES ( 
			?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,? 
		)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	senha, err := seguranca.HashString("usuario123")
	if err != nil {
		return err
	}

	if _, err := stm.Exec(
		v.ID_Viatura,
		v.ID_Franqueado,
		strings.ToUpper(v.Nome),
		strings.ToUpper(v.Nick),
		v.Cpf,
		v.Rg,
		strings.ToUpper(v.Endereco),
		strings.ToUpper(v.Complemento),
		strings.ToUpper(v.Bairro),
		strings.ToUpper(v.Cidade),
		strings.ToUpper(v.Uf),
		strings.ToUpper(v.Cep),
		strings.ToUpper(v.Modelo),
		strings.ToUpper(v.Placa),
		strings.ToUpper(v.Cor),
		v.Telefone1,
		v.Telefone2,
		v.Telefone3,
		strings.ToLower(v.Email),
		senha,
		strings.ToUpper(v.Observacao),
	); err != nil {
		return err
	}

	//=========================================================================
	// Cria o setup de envio ==================================================
	var lista listaEnvioV4.ListaEnvio

	lista.ID_Alvo = v.ID_Viatura
	lista.Email = "S"
	lista.Sms = "N"

	if err := lista.Insere(); err != nil {
		return err
	}

	return nil
}

func (v *Viatura) GetDadosById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE viaturas.ID_Viatura = '%s'`, v.ID_Viatura)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, v); err != nil {
			return err
		}

		return nil
	}

	return errors.New("viatura não encontrada na base de dados")
}

func (v *Viatura) AlteraById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	if v.Nome == "" {
		return errors.New("um nome de viatura deve ser informado")
	}

	if v.Placa == "" {
		return errors.New("um placa de viatura deve ser informado")
	}

	if v.Telefone1 == "" {
		return errors.New("um telefone1 de viatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE viaturas SET 
			viaturas.Nome = ?,
			viaturas.Nick = ?,
			viaturas.Cpf = ?,
			viaturas.Rg = ?,
			viaturas.Endereco = ?,
			viaturas.Complemento = ?, 
			viaturas.Bairro = ?,
			viaturas.Cidade = ?,
			viaturas.Uf = ?,
			viaturas.Cep = ?,
			viaturas.Modelo = ?,
			viaturas.Placa = ?,
			viaturas.Cor = ?,
			viaturas.Telefone1 = ?,
			viaturas.Telefone2 = ?,
			viaturas.Telefone3 = ?,			 
			viaturas.Observacao = ?
		WHERE viaturas.ID_Viatura = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		strings.ToUpper(v.Nome),
		strings.ToUpper(v.Nick),
		v.Cpf,
		v.Rg,
		strings.ToUpper(v.Endereco),
		strings.ToUpper(v.Complemento),
		strings.ToUpper(v.Bairro),
		strings.ToUpper(v.Cidade),
		strings.ToUpper(v.Uf),
		v.Cep,
		strings.ToUpper(v.Modelo),
		strings.ToUpper(v.Placa),
		strings.ToUpper(v.Cor),
		v.Telefone1,
		v.Telefone2,
		v.Telefone3,
		strings.ToUpper(v.Observacao),
		v.ID_Viatura,
	); err != nil {
		return err
	}

	return nil
}

func (v *Viatura) DeleteById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM viaturas WHERE viaturas.ID_Viatura = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(v.ID_Viatura); err != nil {
		return err
	}
	return nil
}

func (v *Viatura) DeleteAllByIdFranqueado() error {
	if v.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM viaturas WHERE viaturas.ID_Franqueado = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(v.ID_Franqueado); err != nil {
		return err
	}
	return nil
}

func (v *Viatura) ListaByIdFranqueado(lista *[]Viatura) error {
	if v.ID_Viatura == "" {
		v.ID_Viatura = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE viaturas.ID_Franqueado = '%s'`, v.ID_Franqueado)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {

		var item Viatura

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

func (v *Viatura) SetEmailById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	if v.Email == "" {
		return errors.New("um email deve ser informado")
	}

	if err := v.VerificaEmailLivreByEmail(); err != nil {
		return err
	}

	if v.Nome == "LIVRE" {

		db, err := connV4.Conectar()
		if err != nil {
			return err
		}
		defer db.Close()

		stm, err := db.Prepare(`
			UPDATE viaturas 
			SET viaturas.Email = ?
			WHERE viaturas.ID_Viatura = ?
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(v.Email, v.ID_Viatura); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("email ja esta sendo usado por: %s", v.Nome)
	}

	return nil
}

func (v *Viatura) VerificaEmailLivreByEmail() error {
	if v.ID_Viatura == "" {
		return errors.New("um email de viatura deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT viaturas.Nome
		FROM viaturas 
		WHERE viaturas.Email = ?
	`, v.ID_Viatura)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&v.Nome); err != nil {
			return err
		}

		return nil
	}

	v.Nome = "LIVRE"
	return nil
}

// Manipula campo Ativo =======================================================

func (v *Viatura) GetAtivoById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = v.ID_Viatura

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	v.Ativo = lb.Ativo
	return nil
}

func (v *Viatura) SetAtivoById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	if v.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = v.ID_Viatura
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = v.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (v *Viatura) InverteAtivoById() error {
	if v.ID_Viatura == "" {
		return errors.New("um id de viatura deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = v.ID_Viatura

	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	v.Ativo = lb.Ativo
	return nil
}

// funcoes internas ===========================================================

func sViaturaToViatura(sv SViatura) (v Viatura) {
	v.ID_Viatura = sv.ID_Viatura.String
	v.ID_Franqueado = sv.ID_Franqueado.String
	v.Nome = sv.Nome.String
	v.Nick = sv.Nick.String
	v.Cpf = sv.Cpf.String
	v.Rg = sv.Rg.String
	v.Endereco = sv.Endereco.String
	v.Complemento = sv.Complemento.String
	v.Bairro = sv.Bairro.String
	v.Cidade = sv.Cidade.String
	v.Uf = sv.Uf.String
	v.Cep = sv.Cep.String
	v.Modelo = sv.Modelo.String
	v.Placa = sv.Placa.String
	v.Cor = sv.Cor.String
	v.Telefone1 = sv.Telefone1.String
	v.Telefone2 = sv.Telefone2.String
	v.Telefone3 = sv.Telefone3.String
	v.Email = sv.Email.String
	v.Senha = ""
	v.Observacao = sv.Observacao.String

	v.FraId = sv.FraId.String
	v.FraRazao = sv.FraRazao.String

	v.RepId = sv.RepId.String
	v.RepRazao = sv.RepRazao.String

	if sv.DataCadastro.Valid {
		v.DataCadastro = sv.DataCadastro.Time.Format("02/01/2006 15:04:05")
	} else {
		v.DataCadastro = ""
	}

	return
}

func getSelect(filtro string) string {

	return fmt.Sprintf(`
		SELECT 
			viaturas.ID_Viatura, 
			viaturas.ID_Franqueado, 
			viaturas.Nome, 
			viaturas.Nick, 
			viaturas.Cpf, 
			viaturas.Rg, 
			viaturas.Endereco,
			viaturas.Complemento, 
			viaturas.Bairro, 
			viaturas.Cidade, 
			viaturas.Uf, 
			viaturas.Cep, 
			viaturas.Modelo, 
			viaturas.Placa, 
			viaturas.Cor, 
			viaturas.Telefone1, 
			viaturas.Telefone2, 
			viaturas.Telefone3, 
			viaturas.Email, 
			viaturas.Senha, 
			viaturas.Observacao, 
			viaturas.DataCadastro, 

			listaEnvio.Email,
			listaEnvio.Sms,

			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,

			representante.ID_Representante,
			representante.RazaoSocial,

			bloqViatura.ID_Alvo,

			bloqFranqueado.ID_Alvo,

			bloqRepresentante.ID_Alvo

		FROM viaturas 

		LEFT JOIN franqueado
		ON viaturas.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN representante
		ON franqueado.ID_Representante = representante.ID_Representante

		LEFT JOIN listaEnvio
		ON viaturas.ID_Viatura = listaEnvio.ID_Alvo

		LEFT JOIN listaBloqueio AS bloqViatura
		ON bloqViatura.ID_Alvo = viaturas.ID_Viatura
		
		LEFT JOIN listaBloqueio AS bloqFranqueado
		ON bloqFranqueado.ID_Alvo =  franqueado.ID_Franqueado

		LEFT JOIN listaBloqueio AS bloqRepresentante
		ON bloqRepresentante.ID_Alvo =  franqueado.ID_Representante
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, v *Viatura) error {

	var (
		tmp SViatura

		envioEmail sql.NullString
		envioSms   sql.NullString

		bloqVia sql.NullString
		bloqFra sql.NullString
		bloqRep sql.NullString
	)

	if err := tab.Scan(
		&tmp.ID_Viatura,
		&tmp.ID_Franqueado,
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
		&tmp.Modelo,
		&tmp.Placa,
		&tmp.Cor,
		&tmp.Telefone1,
		&tmp.Telefone2,
		&tmp.Telefone3,
		&tmp.Email,
		&tmp.Senha,
		&tmp.Observacao,
		&tmp.DataCadastro,

		&envioEmail,
		&envioSms,

		&tmp.FraId,
		&tmp.FraRazao,

		&tmp.RepId,
		&tmp.RepRazao,

		&bloqVia,
		&bloqFra,
		&bloqRep,
	); err != nil {
		return err
	}

	*v = sViaturaToViatura(tmp)

	// Retorna o status de envio de email da viatura
	if envioEmail.Valid {
		v.EnvioEmail = envioEmail.String
	} else {
		v.EnvioEmail = "N"
	}

	// Retorna o status de envio de sms da viatura
	if envioSms.Valid {
		v.EnvioSms = envioEmail.String
	} else {
		v.EnvioSms = "N"
	}

	if bloqVia.Valid {
		v.Ativo = "N"
	} else if bloqFra.Valid {
		v.Ativo = "N"
	} else if bloqRep.Valid {
		v.Ativo = "N"
	} else {
		v.Ativo = "S"
	}

	// Retorna o status de ativo do franqueado
	if bloqFra.Valid {
		v.FraAtivo = "N"
	} else {
		v.FraAtivo = "S"
	}

	// Retorna o status de ativo do franqueado
	if bloqRep.Valid {
		v.RepAtivo = "N"
	} else {
		v.RepAtivo = "S"
	}

	return nil
}
