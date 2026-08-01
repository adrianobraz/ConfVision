package custoAtendimentoV4

import (
	connV4 "api/src/V4/conexao"
	paginacaoV4 "api/src/V4/paginacaoV4"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type CustoAtendimento struct {
	NomeCliente  string `json:"nomeCliente"`
	DataOperacao string `json:"dataOperacao"`
	DadoOperacao string `json:"dadoOperacao"`
	NomeOperador string `json:"nomeOperador"`
	LinkAudio    string `json:"linkAudio"`
}

type FiltroListar struct {
	IDFranqueado string
	IDCliente    string
	DataInicio   string
	DataFim      string
	Limit        int
	Offset       int
}

func ListarLigacoes(f FiltroListar) ([]CustoAtendimento, int, error) {
	idFra := strings.TrimSpace(f.IDFranqueado)
	if idFra == "" {
		return nil, 0, errors.New("idFranqueado obrigatorio")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()

	where := []string{
		"cliente.ID_Franqueado = ?",
		"tarifacao.TipoOperacao = 'LIGAÇÃO'",
	}
	args := []interface{}{idFra}

	idCli := strings.TrimSpace(f.IDCliente)
	if idCli != "" && idCli != "0" {
		where = append(where, "tarifacao.ID_Vinculo = ?")
		args = append(args, idCli)
	}

	di := strings.TrimSpace(f.DataInicio)
	if di != "" && di != "0" {
		where = append(where, "tarifacao.DataOperacao >= ?")
		args = append(args, di)
	}

	df := strings.TrimSpace(f.DataFim)
	if df != "" && df != "0" {
		where = append(where, "tarifacao.DataOperacao <= ?")
		args = append(args, df)
	}

	clause := "WHERE " + strings.Join(where, " AND ")
	baseFrom := `
		FROM tarifacao
		INNER JOIN cliente ON cliente.ID_Cliente = tarifacao.ID_Vinculo
	`

	total := 0
	limit, offset, paginar := paginacaoV4.Normalizar(f.Limit, f.Offset)
	if paginar {
		sqlCount := `SELECT COUNT(*) ` + baseFrom + clause
		var qtd sql.NullInt64
		if err := db.QueryRow(sqlCount, args...).Scan(&qtd); err != nil {
			return nil, 0, err
		}
		total = int(qtd.Int64)
	}

	txtSQL := fmt.Sprintf(`
		SELECT
			cliente.Nome,
			DATE_FORMAT(tarifacao.DataOperacao, '%%d/%%m/%%Y %%H:%%i:%%s'),
			tarifacao.DadoOperacao,
			tarifacao.Aux
		%s
		%s
		ORDER BY tarifacao.DataOperacao DESC
		%s
	`, baseFrom, clause, paginacaoV4.Clausula(limit, offset))

	tab, err := db.Query(txtSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer tab.Close()

	lista := make([]CustoAtendimento, 0)
	for tab.Next() {
		var nome, dataStr, dado, aux sql.NullString
		if err := tab.Scan(&nome, &dataStr, &dado, &aux); err != nil {
			return nil, 0, err
		}

		item := CustoAtendimento{
			NomeCliente:  nome.String,
			DataOperacao: dataStr.String,
			DadoOperacao: dado.String,
			NomeOperador: "",
			LinkAudio:    linkAudioFromAux(aux.String),
		}
		lista = append(lista, item)
	}

	if !paginar {
		total = len(lista)
	}
	return lista, total, nil
}

func linkAudioFromAux(aux string) string {
	a := strings.TrimSpace(aux)
	if a == "" {
		return ""
	}
	lower := strings.ToLower(a)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return a
	}
	return ""
}
