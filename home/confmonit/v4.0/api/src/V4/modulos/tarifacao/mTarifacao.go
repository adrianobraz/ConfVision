package tarifacaoV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Tarifacao struct {
	ID_Tarifacao   string `json:"idTarifacao"`
	ID_Vinculo     string `json:"idVinculo"`
	IDCentralUUID  string `json:"idCentralUUID"`
	TipoOperacao   string `json:"tipoOperacao"`
	DadoOperacao   string `json:"dadoOperacao"`
	Credito        string `json:"credito"`
	Debito         string `json:"debito"`
	DataOperacao   string `json:"dataOperacao"`
	NomeVinculo    string `json:"nomeVinculo"`
	// Tipo: "0"=Debito, "1"=Credito (atalho do formulário)
	Tipo string `json:"tipo"`
	// Valor mascarado do formulário
	Valor               string `json:"valor"`
	ListarTodasCentrais bool   `json:"listarTodasCentrais"`
}

func (t *Tarifacao) InsereLancamento() error {
	if t.ID_Vinculo == "" {
		return errors.New("um representante deve ser informado")
	}
	desc := strings.TrimSpace(t.DadoOperacao)
	if desc == "" {
		return errors.New("uma descricao deve ser informada")
	}

	valorStr := t.Valor
	if valorStr == "" {
		if t.Debito != "" {
			valorStr = t.Debito
		} else if t.Credito != "" {
			valorStr = t.Credito
		}
	}
	var valor float64
	if err := auxiliar.MoneyBrToFloat(valorStr, &valor); err != nil || valor <= 0 {
		return errors.New("valor invalido")
	}
	auxiliar.FloatToString(valor, &valorStr)

	credito := "0,00"
	debito := "0,00"
	tipo := strings.TrimSpace(t.Tipo)
	if tipo == "" {
		tipo = "0"
	}
	switch tipo {
	case "1", "CREDITO", "C":
		credito = valorStr
	default:
		debito = valorStr
	}

	if t.TipoOperacao == "" {
		t.TipoOperacao = "LANCAMENTO"
	}
	if t.ID_Tarifacao == "" {
		t.ID_Tarifacao = auxiliar.GeradorDeId()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO tarifacao(
			ID_Tarifacao,
			ID_Vinculo,
			TipoOperacao,
			DadoOperacao,
			Credito,
			Debito
		) VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		t.ID_Tarifacao,
		t.ID_Vinculo,
		t.TipoOperacao,
		strings.ToUpper(desc),
		credito,
		debito,
	); err != nil {
		return err
	}

	t.Credito = credito
	t.Debito = debito
	t.DadoOperacao = strings.ToUpper(desc)
	return nil
}

func (t *Tarifacao) ListaLancamentosPendentes(lista *[]Tarifacao) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlTxt := `
		SELECT
			tarifacao.ID_Tarifacao,
			tarifacao.ID_Vinculo,
			tarifacao.TipoOperacao,
			tarifacao.DadoOperacao,
			tarifacao.Credito,
			tarifacao.Debito,
			DATE_FORMAT(tarifacao.DataOperacao, '%d/%m/%Y %H:%i:%s'),
			COALESCE(destRep.RazaoSocial, destFra.RazaoSocial, destCli.Nome, tarifacao.ID_Vinculo)
		FROM tarifacao
		LEFT JOIN representante AS destRep ON tarifacao.ID_Vinculo = destRep.ID_Representante
		LEFT JOIN franqueado AS destFra ON tarifacao.ID_Vinculo = destFra.ID_Franqueado
		LEFT JOIN cliente AS destCli ON tarifacao.ID_Vinculo = destCli.ID_Cliente
		WHERE tarifacao.TipoOperacao = 'LANCAMENTO'
		AND tarifacao.ID_FaturaNuc = '0'
	`
	args := []interface{}{}
	if err := auxiliar.ExigirFiltroCentralUUID(t.IDCentralUUID, t.ListarTodasCentrais); err != nil {
		return err
	}
	if !t.ListarTodasCentrais {
		sqlTxt += ` AND destRep.IDCentralUUID = ?`
		args = append(args, t.IDCentralUUID)
	}
	if t.ID_Vinculo != "" {
		sqlTxt += ` AND tarifacao.ID_Vinculo = ?`
		args = append(args, t.ID_Vinculo)
	}
	sqlTxt += ` ORDER BY tarifacao.DataOperacao DESC LIMIT 200`

	tab, err := db.Query(sqlTxt, args...)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Tarifacao
		var data sql.NullString
		var nome sql.NullString
		if err := tab.Scan(
			&item.ID_Tarifacao,
			&item.ID_Vinculo,
			&item.TipoOperacao,
			&item.DadoOperacao,
			&item.Credito,
			&item.Debito,
			&data,
			&nome,
		); err != nil {
			return err
		}
		item.DataOperacao = data.String
		item.NomeVinculo = nome.String
		*lista = append(*lista, item)
	}
	return nil
}

func (t *Tarifacao) DeleteById() error {
	if t.ID_Tarifacao == "" {
		return errors.New("um id de tarifacao deve ser informado")
	}
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM tarifacao
		WHERE ID_Tarifacao = ?
		AND TipoOperacao = 'LANCAMENTO'
		AND ID_FaturaNuc = '0'
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	res, err := stm.Exec(t.ID_Tarifacao)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return fmt.Errorf("lancamento nao encontrado ou ja faturado")
	}
	return nil
}
