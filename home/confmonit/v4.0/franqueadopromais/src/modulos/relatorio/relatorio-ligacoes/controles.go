package relatorioLigacoes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"franqueadopro/src/auxiliar"
	"franqueadopro/src/config"

	_ "github.com/go-sql-driver/mysql"
)

type ligacaoItem struct {
	NomeCliente  string `json:"nomeCliente"`
	DataOperacao string `json:"dataOperacao"`
	DadoOperacao string `json:"dadoOperacao"`
	NomeOperador string `json:"nomeOperador"`
	LinkAudio    string `json:"linkAudio"`
}

type ligacaoReq struct {
	IDFranqueado string `json:"idFranqueado"`
	IDCliente    string `json:"idCliente"`
	DataInicio   string `json:"dataInicio"`
	DataFim      string `json:"dataFim"`
	Filtro       string `json:"filtro"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
}

func CarregarRelatorioLigacoes(w http.ResponseWriter, r *http.Request) {
	var d auxiliar.Pagina
	d.TituloSite = config.TituloSite
	d.NomeTela = "Relatório de Ligações"
	d.LogoMarca = "logo2Id6.png"
	d.LinkRetorno = "/carregar-menu-relatorio"
	auxiliar.ExecutarTemplate(w, "relatorio-ligacoes.html", d)
}

func CustoAtendimentoListar(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var req ligacaoReq
	if erro := json.Unmarshal(corpo, &req); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if strings.TrimSpace(req.IDFranqueado) == "" && strings.TrimSpace(req.Filtro) != "" {
		parseFiltroLegado(req.Filtro, &req)
	}
	if strings.TrimSpace(req.IDFranqueado) == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("idFranqueado obrigatorio"))
		return
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	lista, total, fonte, erro := listarLigacoes(req)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	if lista == nil {
		lista = []ligacaoItem{}
	}

	hasMore := req.Offset+len(lista) < total
	auxiliar.RespostaJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "OK",
		"dados":   lista,
		"total":   total,
		"limit":   req.Limit,
		"offset":  req.Offset,
		"hasMore": hasMore,
		"fonte":   fonte,
	})
}

func listarLigacoes(req ligacaoReq) ([]ligacaoItem, int, string, error) {
	dsn := config.DBDSN()
	if dsn == "" {
		return nil, 0, "", fmt.Errorf("banco nao configurado no FranqueadoPro (.env: BD_HOST_MV4, BD_USER_MV4, BD_PASS_MV4, BD_BASE_MV4)")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, 0, "", err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return nil, 0, "", fmt.Errorf("falha ao conectar no MySQL: %v", err)
	}

	// 1) tarifacao = onde o terminal grava as ligacoes hoje
	lista, total, errTar := queryTarifacao(db, req)
	if errTar == nil {
		return lista, total, "tarifacao", nil
	}

	// 2) custoAtendimento = tabela/view legada (so se tarifacao falhou)
	lista2, total2, errCusto := queryCustoAtendimento(db, req)
	if errCusto == nil {
		return lista2, total2, "custoAtendimento", nil
	}

	return nil, 0, "", errTar
}

func queryTarifacao(db *sql.DB, req ligacaoReq) ([]ligacaoItem, int, error) {
	where := []string{
		"cliente.ID_Franqueado = ?",
		"UPPER(TRIM(tarifacao.TipoOperacao)) LIKE 'LIGA%'",
	}
	args := []interface{}{strings.TrimSpace(req.IDFranqueado)}

	if id := strings.TrimSpace(req.IDCliente); id != "" && id != "0" {
		where = append(where, "tarifacao.ID_Vinculo = ?")
		args = append(args, id)
	}
	if di := strings.TrimSpace(req.DataInicio); di != "" && di != "0" {
		where = append(where, "tarifacao.DataOperacao >= ?")
		args = append(args, normalizarData(di))
	}
	if df := strings.TrimSpace(req.DataFim); df != "" && df != "0" {
		where = append(where, "tarifacao.DataOperacao <= ?")
		args = append(args, normalizarData(df))
	}
	clause := "WHERE " + strings.Join(where, " AND ")
	from := `
		FROM tarifacao
		INNER JOIN cliente ON cliente.ID_Cliente = tarifacao.ID_Vinculo
	`

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) `+from+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sqlTxt := `
		SELECT
			COALESCE(cliente.Nome, ''),
			DATE_FORMAT(tarifacao.DataOperacao, '%d/%m/%Y %H:%i:%s'),
			COALESCE(tarifacao.DadoOperacao, ''),
			'',
			COALESCE(tarifacao.Aux, '')
		` + from + clause + `
		ORDER BY tarifacao.DataOperacao DESC
		LIMIT ? OFFSET ?`
	argsQ := append(append([]interface{}{}, args...), req.Limit, req.Offset)

	rows, err := db.Query(sqlTxt, argsQ...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	lista := make([]ligacaoItem, 0)
	for rows.Next() {
		var item ligacaoItem
		if err := rows.Scan(&item.NomeCliente, &item.DataOperacao, &item.DadoOperacao, &item.NomeOperador, &item.LinkAudio); err != nil {
			return nil, 0, err
		}
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(item.LinkAudio)), "http") {
			item.LinkAudio = ""
		}
		lista = append(lista, item)
	}
	return lista, total, nil
}

func queryCustoAtendimento(db *sql.DB, req ligacaoReq) ([]ligacaoItem, int, error) {
	where := []string{
		"c.ID_Franqueado = ?",
		"UPPER(c.TipoOperacao) LIKE 'LIGA%'",
	}
	args := []interface{}{strings.TrimSpace(req.IDFranqueado)}

	if id := strings.TrimSpace(req.IDCliente); id != "" && id != "0" {
		where = append(where, "c.ID_Cliente = ?")
		args = append(args, id)
	}
	if di := strings.TrimSpace(req.DataInicio); di != "" && di != "0" {
		where = append(where, "c.DataOperacao >= ?")
		args = append(args, normalizarData(di))
	}
	if df := strings.TrimSpace(req.DataFim); df != "" && df != "0" {
		where = append(where, "c.DataOperacao <= ?")
		args = append(args, normalizarData(df))
	}
	clause := "WHERE " + strings.Join(where, " AND ")

	// JOIN cliente para nome — evita depender da coluna NomeCliente
	from := `
		FROM custoAtendimento c
		LEFT JOIN cliente cli ON cli.ID_Cliente = c.ID_Cliente
	`

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) `+from+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sqlTxt := `
		SELECT
			COALESCE(cli.Nome, ''),
			DATE_FORMAT(c.DataOperacao, '%d/%m/%Y %H:%i:%s'),
			COALESCE(c.DadoOperacao, ''),
			COALESCE(c.NomeOperador, ''),
			COALESCE(c.LinkAudio, '')
		` + from + clause + `
		ORDER BY c.DataOperacao DESC
		LIMIT ? OFFSET ?`
	argsQ := append(append([]interface{}{}, args...), req.Limit, req.Offset)

	rows, err := db.Query(sqlTxt, argsQ...)
	if err != nil {
		// Schema legado pode nao ter NomeOperador/LinkAudio — tenta SELECT minimo
		return queryCustoAtendimentoMinimo(db, req)
	}
	defer rows.Close()

	lista := make([]ligacaoItem, 0)
	for rows.Next() {
		var item ligacaoItem
		if err := rows.Scan(&item.NomeCliente, &item.DataOperacao, &item.DadoOperacao, &item.NomeOperador, &item.LinkAudio); err != nil {
			return nil, 0, err
		}
		lista = append(lista, item)
	}
	return lista, total, nil
}

func queryCustoAtendimentoMinimo(db *sql.DB, req ligacaoReq) ([]ligacaoItem, int, error) {
	where := []string{
		"c.ID_Franqueado = ?",
		"UPPER(c.TipoOperacao) LIKE 'LIGA%'",
	}
	args := []interface{}{strings.TrimSpace(req.IDFranqueado)}
	if id := strings.TrimSpace(req.IDCliente); id != "" && id != "0" {
		where = append(where, "c.ID_Cliente = ?")
		args = append(args, id)
	}
	if di := strings.TrimSpace(req.DataInicio); di != "" && di != "0" {
		where = append(where, "c.DataOperacao >= ?")
		args = append(args, normalizarData(di))
	}
	if df := strings.TrimSpace(req.DataFim); df != "" && df != "0" {
		where = append(where, "c.DataOperacao <= ?")
		args = append(args, normalizarData(df))
	}
	clause := "WHERE " + strings.Join(where, " AND ")
	from := `
		FROM custoAtendimento c
		LEFT JOIN cliente cli ON cli.ID_Cliente = c.ID_Cliente
	`

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) `+from+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sqlTxt := `
		SELECT
			COALESCE(cli.Nome, ''),
			DATE_FORMAT(c.DataOperacao, '%d/%m/%Y %H:%i:%s'),
			COALESCE(c.DadoOperacao, '')
		` + from + clause + `
		ORDER BY c.DataOperacao DESC
		LIMIT ? OFFSET ?`
	argsQ := append(append([]interface{}{}, args...), req.Limit, req.Offset)
	rows, err := db.Query(sqlTxt, argsQ...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	lista := make([]ligacaoItem, 0)
	for rows.Next() {
		var item ligacaoItem
		if err := rows.Scan(&item.NomeCliente, &item.DataOperacao, &item.DadoOperacao); err != nil {
			return nil, 0, err
		}
		lista = append(lista, item)
	}
	return lista, total, nil
}

func normalizarData(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "T", " ")
	for strings.Count(s, ":") > 2 {
		s = s[:strings.LastIndex(s, ":")]
	}
	if strings.Count(s, ":") == 1 {
		s += ":00"
	}
	return s
}

func parseFiltroLegado(filtro string, req *ligacaoReq) {
	pick := func(key string) string {
		i := strings.Index(strings.ToUpper(filtro), strings.ToUpper(key))
		if i < 0 {
			return ""
		}
		rest := filtro[i+len(key):]
		rest = strings.TrimLeft(rest, " ='\"")
		end := strings.IndexAny(rest, "'\" \n\r\t")
		if end < 0 {
			return strings.TrimSpace(rest)
		}
		return strings.TrimSpace(rest[:end])
	}
	if req.IDFranqueado == "" {
		req.IDFranqueado = pick("ID_Franqueado =")
	}
	if req.IDCliente == "" {
		req.IDCliente = pick("ID_Cliente =")
	}
	if req.DataInicio == "" {
		req.DataInicio = pick("DataOperacao >=")
	}
	if req.DataFim == "" {
		req.DataFim = pick("DataOperacao <=")
	}
}
