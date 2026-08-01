package eventoV4

import (
	connV4 "api/src/V4/conexao"
	paginacaoV4 "api/src/V4/paginacaoV4"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Grupos exibidos no dashboard de eventos (ultimos 7 dias).
var gruposDashboardEvento = []string{
	"ALARME", "ARME", "DESARME", "EMERGENCIA", "FALHAS", "GERAL",
	"MEDICO", "PANICO", "RESTAURE", "SETUP", "TESTE",
}

type ContagemEventoPeriodo struct {
	Total int `json:"total"`
}

type ContagemEventosGrupo map[string]int

// JOIN direto (substitui subquery IN) — permite uso de indices em cliente, processo e evento.
const sqlJoinEventoFranqueado = `
	INNER JOIN processo ON evento.ID_Processo = processo.ID_Processo
	INNER JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
	INNER JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
`

const sqlWhereFranqueadoPeriodo = `
	WHERE cliente.ID_Franqueado = ?
	AND processo.DataCriacao >= ?
	AND processo.DataCriacao <= ?
`

func (ef *EvtFiltro) ContarByIdFranqueadoPeriodo(total *int) error {
	if ef.IdFranqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}
	if ef.DataInicio == "" || ef.DataFim == "" {
		return errors.New("data inicial e final devem ser informadas")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlContagem := `
		SELECT COUNT(*)
		FROM evento
	` + sqlJoinEventoFranqueado + sqlWhereFranqueadoPeriodo

	var qtd sql.NullInt64
	if err := db.QueryRow(sqlContagem, ef.IdFranqueado, ef.DataInicio, ef.DataFim).Scan(&qtd); err != nil {
		return err
	}
	*total = int(qtd.Int64)
	return nil
}

func (ef *EvtFiltro) ContarByIdFranqueadoGrupo(contagem *ContagemEventosGrupo) error {
	if ef.IdFranqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}
	if ef.DataInicio == "" || ef.DataFim == "" {
		return errors.New("data inicial e final devem ser informadas")
	}

	*contagem = make(ContagemEventosGrupo)
	for _, g := range gruposDashboardEvento {
		(*contagem)[g] = 0
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlGrupos := `
		SELECT 
			UPPER(COALESCE(NULLIF(ctiPersonalizado.Grupo, ''), ctiPadrao.Grupo)) AS grupo,
			COUNT(*) AS total
		FROM evento
	` + sqlJoinEventoFranqueado + `
		LEFT JOIN contactId AS ctiPadrao
			ON evento.Codigo = ctiPadrao.Codigo AND ctiPadrao.ID_Vinculo = 'CENTRAL'
		LEFT JOIN contactId AS ctiPersonalizado
			ON evento.Codigo = ctiPersonalizado.Codigo AND ctiPersonalizado.ID_Vinculo = cliente.ID_Franqueado
	` + sqlWhereFranqueadoPeriodo + `
		GROUP BY UPPER(COALESCE(NULLIF(ctiPersonalizado.Grupo, ''), ctiPadrao.Grupo))
	`

	tab, err := db.Query(sqlGrupos, ef.IdFranqueado, ef.DataInicio, ef.DataFim)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var grupo sql.NullString
		var total sql.NullInt64
		if err := tab.Scan(&grupo, &total); err != nil {
			return err
		}
		g := strings.ToUpper(strings.TrimSpace(grupo.String))
		if g != "" {
			(*contagem)[g] = int(total.Int64)
		}
	}
	return nil
}

func (ef *EvtFiltro) ListarByIdFranqueadoStartEndGrupo(lista *[]EvtFiltro) (int, error) {
	if ef.IdFranqueado == "" {
		return 0, errors.New("um id de franqueado deve ser informado")
	}
	if ef.DataInicio == "" || ef.DataFim == "" {
		return 0, errors.New("data inicial e final devem ser informadas")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	args := []interface{}{ef.IdFranqueado, ef.DataInicio, ef.DataFim}
	filtroGrupo := ""
	if strings.TrimSpace(ef.Grupo) != "" {
		filtroGrupo = `
			AND UPPER(COALESCE(NULLIF(ctiPersonalizado.Grupo, ''), ctiPadrao.Grupo)) = UPPER(?)
		`
		args = append(args, strings.TrimSpace(ef.Grupo))
	}

	fromWhere := `
		FROM evento
	` + sqlJoinEventoFranqueado + `
		LEFT JOIN franqueado ON cliente.ID_Franqueado = franqueado.ID_Franqueado
		LEFT JOIN contactId AS ctiPadrao
			ON evento.Codigo = ctiPadrao.Codigo AND ctiPadrao.ID_Vinculo = 'CENTRAL'
		LEFT JOIN contactId AS ctiPersonalizado
			ON evento.Codigo = ctiPersonalizado.Codigo AND ctiPersonalizado.ID_Vinculo = cliente.ID_Franqueado
		LEFT JOIN usuariosAlarme
			ON dispositivo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo
			AND evento.ZonaUser = usuariosAlarme.Codigo
		LEFT JOIN setorAlarme
			ON dispositivo.ID_Dispositivo = setorAlarme.ID_Dispositivo
			AND evento.ZonaUser = setorAlarme.Numero
	` + sqlWhereFranqueadoPeriodo + filtroGrupo

	total := 0
	if ef.Limit > 0 {
		sqlCount := `SELECT COUNT(*) ` + fromWhere
		var qtd sql.NullInt64
		if err := db.QueryRow(sqlCount, args...).Scan(&qtd); err != nil {
			return 0, err
		}
		total = int(qtd.Int64)
	}

	txtSql := fmt.Sprintf(`
		SELECT 
			evento.DataEntrada,
			evento.Codigo,
			evento.ZonaUser,
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			dispositivo.Conta,
			dispositivo.DataUltimoEvento,
			cliente.ID_Cliente,
			cliente.Nome,
			franqueado.RazaoSocial,
			ctiPadrao.Descricao,
			ctiPadrao.Grupo,
			ctiPersonalizado.Descricao,
			ctiPersonalizado.Grupo,
			usuariosAlarme.Nome,
			usuariosAlarme.Codigo,
			setorAlarme.Nome,
			setorAlarme.Numero
		%s
		ORDER BY evento.DataEntrada DESC
		%s
	`, fromWhere, paginacaoV4.Clausula(ef.Limit, ef.Offset))

	tab, err := db.Query(txtSql, args...)
	if err != nil {
		return 0, err
	}
	defer tab.Close()

	for tab.Next() {
		item, err := scanEvtFiltroDashboard(tab)
		if err != nil {
			return 0, err
		}
		*lista = append(*lista, item)
	}

	if ef.Limit <= 0 {
		total = len(*lista)
	}
	return total, nil
}

func scanEvtFiltroDashboard(tab *sql.Rows) (EvtFiltro, error) {
	var (
		dataEntrada       sql.NullTime
		codigo            sql.NullString
		zonaUser          sql.NullString
		idDispositivo     sql.NullString
		dispNome          sql.NullString
		dispConta         sql.NullString
		dispDataUltimoEvt sql.NullTime
		idCliente         sql.NullString
		cliNome           sql.NullString
		idFranqueado      sql.NullString
		padDescricao      sql.NullString
		padGrupo          sql.NullString
		perDescricao      sql.NullString
		perGrupo          sql.NullString
		usuNome           sql.NullString
		usuCodigo         sql.NullString
		setNome           sql.NullString
		setNumero         sql.NullString
	)

	if err := tab.Scan(
		&dataEntrada, &codigo, &zonaUser,
		&idDispositivo, &dispNome, &dispConta, &dispDataUltimoEvt,
		&idCliente, &cliNome, &idFranqueado,
		&padDescricao, &padGrupo, &perDescricao, &perGrupo,
		&usuNome, &usuCodigo, &setNome, &setNumero,
	); err != nil {
		return EvtFiltro{}, err
	}

	var item EvtFiltro
	item.DataEntrada = dataEntrada.Time.Format("02/01/2006 15:04:05")
	item.Codigo = codigo.String
	item.DispId = idDispositivo.String
	item.DispNome = dispNome.String
	item.DispConta = dispConta.String
	item.DispDataUltimoEvt = dispDataUltimoEvt.Time.Format("02/01/2006 15:04:05")
	item.CliId = idCliente.String
	item.CliNome = cliNome.String
	item.FraId = idFranqueado.String

	if perDescricao.Valid {
		item.CtiDescricao = perDescricao.String
		item.CtiGrupo = perGrupo.String
	} else {
		item.CtiDescricao = padDescricao.String
		item.CtiGrupo = padGrupo.String
	}

	if perGrupo.String == "ARME" || padGrupo.String == "ARME" ||
		perGrupo.String == "DESARME" || padGrupo.String == "DESARME" {
		item.ZonaUser = usuCodigo.String
		item.ZonaUserDescricao = strings.ToUpper(usuNome.String)
		if item.ZonaUser == "" {
			item.ZonaUser = zonaUser.String
		}
	} else {
		item.ZonaUser = setNumero.String
		item.ZonaUserDescricao = strings.ToUpper(setNome.String)
		if item.ZonaUser == "" {
			item.ZonaUser = zonaUser.String
		}
	}

	if item.ZonaUser == "" {
		item.ZonaUser = "NÃO CADASTRADO"
	}
	if item.ZonaUserDescricao == "" {
		item.ZonaUserDescricao = "NÃO CADASTRADO"
	}

	return item, nil
}
