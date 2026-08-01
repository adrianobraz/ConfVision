package centralV4

import (
	connV4 "api/src/V4/conexao"
	usuariosV4 "api/src/V4/modulos/usuarios"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Central representa uma empresa Central (tenant).
type Central struct {
	ID_Central    string `json:"idCentral"`
	IDCentralUUID string `json:"idCentralUUID"`
	RazaoSocial   string `json:"razaoSocial"`
	NomeFantasia  string `json:"nomeFantasia"`
	Cnpj          string `json:"cnpj"`
	Inscricao     string `json:"inscricaoEstadual"`
	Cep           string `json:"cep"`
	Endereco      string `json:"endereco"`
	Complemento   string `json:"complemento"`
	Bairro        string `json:"bairro"`
	Cidade        string `json:"cidade"`
	Estado        string `json:"estado"`
	Telefone1     string `json:"telefone1"`
	Telefone2     string `json:"telefone2"`
	// Usuário master da nova central
	UserMaster usuariosV4.Usuario `json:"userMaster"`
}

type SCentral struct {
	ID_Central    sql.NullString
	IDCentralUUID sql.NullString
	RazaoSocial   sql.NullString
	NomeFantasia  sql.NullString
	Cnpj          sql.NullString
	Inscricao     sql.NullString
	Cep           sql.NullString
	Endereco      sql.NullString
	Complemento   sql.NullString
	Bairro        sql.NullString
	Cidade        sql.NullString
	Estado        sql.NullString
	Telefone1     sql.NullString
	Telefone2     sql.NullString
}

// Insere cria uma nova Central + usuário master.
// ID_Central recebe o identificador único; IDCentralUUID espelha o mesmo valor.
// Filhas (representante/usuarios/...) vinculam por ID_Central via coluna IDCentralUUID.
func (c *Central) Insere() error {
	c.RazaoSocial = strings.TrimSpace(c.RazaoSocial)
	c.NomeFantasia = strings.TrimSpace(c.NomeFantasia)
	if c.RazaoSocial == "" {
		return errors.New("razao social obrigatoria")
	}
	if c.UserMaster.Email1 == "" {
		return errors.New("email do usuario master obrigatorio")
	}
	if c.UserMaster.Nome == "" {
		c.UserMaster.Nome = c.RazaoSocial
	}
	if c.UserMaster.Nick == "" {
		c.UserMaster.Nick = "MASTER"
	}

	id := auxiliar.GeradorDeId()
	c.ID_Central = id
	c.IDCentralUUID = id

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := c.inserirCentral(db); err != nil {
		return err
	}

	c.UserMaster.ID_Usuario = ""
	c.UserMaster.ID_Vinculo = "CENTRAL"
	c.UserMaster.IDCentralUUID = c.IDCentralUUID
	c.UserMaster.Tipo = "CEN"
	c.UserMaster.UsuarioWeb = "S"
	c.UserMaster.UsuarioTeminal = "S"
	c.UserMaster.Master = "S"
	c.UserMaster.AdmFinanceiro = "S"

	if err := c.UserMaster.Insere(); err != nil {
		_, _ = db.Exec(`DELETE FROM central WHERE ID_Central = ?`, c.ID_Central)
		return err
	}

	c.UserMaster.UsuarioWeb = "S"
	_ = c.UserMaster.SetWebAtivaById()

	c.UserMaster.Senha = ""
	return nil
}

// inserirCentral preenche TODAS as colunas da tabela via SHOW COLUMNS.
// Campos de integração (Benuvem/VoIP/email) recebem placeholder — não serão usados.
func (c *Central) inserirCentral(db *sql.DB) error {
	rows, err := db.Query("SHOW COLUMNS FROM `central`")
	if err != nil {
		return err
	}
	defer rows.Close()

	valores := map[string]interface{}{
		"ID_Central":        c.ID_Central,
		"IDCentralUUID":     c.IDCentralUUID,
		"RazaoSocial":       strings.ToUpper(c.RazaoSocial),
		"NomeFantasia":      strings.ToUpper(c.NomeFantasia),
		"Cnpj":              c.Cnpj,
		"EscricaoEstadual":  c.Inscricao,
		"InscricaoEstadual": c.Inscricao,
		"Cep":               c.Cep,
		"Endereco":          strings.ToUpper(c.Endereco),
		"Complemento":       strings.ToUpper(c.Complemento),
		"Bairro":            strings.ToUpper(c.Bairro),
		"Cidade":            strings.ToUpper(c.Cidade),
		"Estado":            strings.ToUpper(c.Estado),
		"Telefone1":         auxiliar.ClearTel(c.Telefone1),
		"Telefone2":         auxiliar.ClearTel(c.Telefone2),
		"ServidorPorta":     "4000",
		"ServidorVersao":    "4.0",
		"ServidorKey":       "-",
		"EmailEnviarAtivo":  "N",
		"BenuvemAtivo":      "N",
		"BenuvemEmail":      "nao-usar@local",
		"BenuvemSenha":      "nao-usar",
		"VoipAtivo":         "N",
		"VoipExibirLog":     "N",
		"VoipToken":         "-",
		"VoipKey":           "-",
		"voipDeviceId":      "-",
		"VoipDeviceId":      "-",
	}

	var cols []string
	var placeholders []string
	var args []interface{}

	for rows.Next() {
		var field, colType, null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			return err
		}
		extraLower := strings.ToLower(extra)
		if strings.Contains(extraLower, "auto_increment") || strings.Contains(extraLower, "generated") {
			continue
		}

		cols = append(cols, "`"+field+"`")
		placeholders = append(placeholders, "?")

		if v, ok := valores[field]; ok {
			args = append(args, v)
			continue
		}

		// Qualquer outra coluna: valor dummy (não usamos esses dados)
		t := strings.ToLower(colType)
		switch {
		case strings.Contains(t, "int"), strings.Contains(t, "decimal"),
			strings.Contains(t, "float"), strings.Contains(t, "double"):
			args = append(args, 0)
		case strings.HasPrefix(t, "date"), strings.Contains(t, "time"):
			args = append(args, "2000-01-01 00:00:00")
		default:
			args = append(args, "-")
		}
	}

	if len(cols) == 0 {
		return errors.New("tabela central sem colunas")
	}

	q := fmt.Sprintf(
		"INSERT INTO `central` (%s) VALUES (%s)",
		strings.Join(cols, ","),
		strings.Join(placeholders, ","),
	)
	_, err = db.Exec(q, args...)
	return err
}

// AlteraById atualiza dados cadastrais da Central (não altera usuário master).
func (c *Central) AlteraById() error {
	c.ID_Central = strings.TrimSpace(c.ID_Central)
	if c.ID_Central == "" {
		c.ID_Central = strings.TrimSpace(c.IDCentralUUID)
	}
	c.RazaoSocial = strings.TrimSpace(c.RazaoSocial)
	c.NomeFantasia = strings.TrimSpace(c.NomeFantasia)
	if c.ID_Central == "" {
		return errors.New("id da central obrigatorio")
	}
	if c.RazaoSocial == "" {
		return errors.New("razao social obrigatoria")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query("SHOW COLUMNS FROM `central`")
	if err != nil {
		return err
	}
	defer rows.Close()

	valores := map[string]interface{}{
		"RazaoSocial":       strings.ToUpper(c.RazaoSocial),
		"NomeFantasia":      strings.ToUpper(c.NomeFantasia),
		"Cnpj":              c.Cnpj,
		"EscricaoEstadual":  c.Inscricao,
		"InscricaoEstadual": c.Inscricao,
		"Cep":               c.Cep,
		"Endereco":          strings.ToUpper(c.Endereco),
		"Complemento":       strings.ToUpper(c.Complemento),
		"Bairro":            strings.ToUpper(c.Bairro),
		"Cidade":            strings.ToUpper(c.Cidade),
		"Estado":            strings.ToUpper(c.Estado),
		"Telefone1":         auxiliar.ClearTel(c.Telefone1),
		"Telefone2":         auxiliar.ClearTel(c.Telefone2),
	}

	var sets []string
	var args []interface{}
	for rows.Next() {
		var field, colType, null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			return err
		}
		v, ok := valores[field]
		if !ok {
			continue
		}
		sets = append(sets, "`"+field+"` = ?")
		args = append(args, v)
	}
	if len(sets) == 0 {
		return errors.New("nenhum campo cadastral encontrado na tabela central")
	}

	q := fmt.Sprintf(
		"UPDATE `central` SET %s WHERE ID_Central = ? OR IDCentralUUID = ?",
		strings.Join(sets, ", "),
	)
	argsUUID := append(append([]interface{}{}, args...), c.ID_Central, c.ID_Central)

	res, err := db.Exec(q, argsUUID...)
	if err != nil {
		// Schemas sem IDCentralUUID: só ID_Central
		q2 := fmt.Sprintf(
			"UPDATE `central` SET %s WHERE ID_Central = ?",
			strings.Join(sets, ", "),
		)
		argsID := append(append([]interface{}{}, args...), c.ID_Central)
		res, err = db.Exec(q2, argsID...)
		if err != nil {
			return err
		}
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("central nao encontrada")
	}
	return nil
}

func (c *Central) Lista(lista *[]Central) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Tentativas: schemas com InscricaoEstadual / EscricaoEstadual / sem UUID
	queries := []string{
		`SELECT
			ID_Central,
			COALESCE(NULLIF(TRIM(IDCentralUUID), ''), ID_Central),
			RazaoSocial,
			NomeFantasia,
			Cnpj,
			COALESCE(InscricaoEstadual, ''),
			Cep,
			Endereco,
			Complemento,
			Bairro,
			Cidade,
			Estado,
			Telefone1,
			Telefone2
		FROM central
		ORDER BY RazaoSocial`,
		`SELECT
			ID_Central,
			COALESCE(NULLIF(TRIM(IDCentralUUID), ''), ID_Central),
			RazaoSocial,
			NomeFantasia,
			Cnpj,
			COALESCE(EscricaoEstadual, ''),
			Cep,
			Endereco,
			Complemento,
			Bairro,
			Cidade,
			Estado,
			Telefone1,
			Telefone2
		FROM central
		ORDER BY RazaoSocial`,
		`SELECT
			ID_Central,
			ID_Central,
			RazaoSocial,
			NomeFantasia,
			Cnpj,
			'',
			Cep,
			Endereco,
			Complemento,
			Bairro,
			Cidade,
			Estado,
			Telefone1,
			Telefone2
		FROM central
		ORDER BY RazaoSocial`,
		`SELECT
			ID_Central,
			ID_Central,
			RazaoSocial,
			NomeFantasia,
			Cnpj,
			'',
			'',
			'',
			'',
			'',
			'',
			'',
			'',
			''
		FROM central
		ORDER BY RazaoSocial`,
	}

	var tab *sql.Rows
	var lastErr error
	for _, q := range queries {
		tab, lastErr = db.Query(q)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return lastErr
	}
	defer tab.Close()

	for tab.Next() {
		var s SCentral
		if err := tab.Scan(
			&s.ID_Central,
			&s.IDCentralUUID,
			&s.RazaoSocial,
			&s.NomeFantasia,
			&s.Cnpj,
			&s.Inscricao,
			&s.Cep,
			&s.Endereco,
			&s.Complemento,
			&s.Bairro,
			&s.Cidade,
			&s.Estado,
			&s.Telefone1,
			&s.Telefone2,
		); err != nil {
			return err
		}
		*lista = append(*lista, Central{
			ID_Central:    s.ID_Central.String,
			IDCentralUUID: s.IDCentralUUID.String,
			RazaoSocial:   s.RazaoSocial.String,
			NomeFantasia:  s.NomeFantasia.String,
			Cnpj:          s.Cnpj.String,
			Inscricao:     s.Inscricao.String,
			Cep:           s.Cep.String,
			Endereco:      s.Endereco.String,
			Complemento:   s.Complemento.String,
			Bairro:        s.Bairro.String,
			Cidade:        s.Cidade.String,
			Estado:        s.Estado.String,
			Telefone1:     s.Telefone1.String,
			Telefone2:     s.Telefone2.String,
		})
	}
	return nil
}
