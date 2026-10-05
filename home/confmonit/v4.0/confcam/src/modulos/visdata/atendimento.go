package visdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	RecursoIA       = "inteligencia_artificial"
	RecursoAutofim  = "finalizacao_automatica"
	RecursoParceiro = "parceiro"
	RecursoEmail    = "email"

	ModoTodos      = "todos"
	ModoTodosMenos = "todos_menos"
	ModoSomente    = "somente"
)

type PoliticaAtendimento struct {
	IDFranqueado            string   `json:"id_franqueado"`
	InteligenciaArtificial  bool     `json:"inteligencia_artificial"`
	IAModo                  string   `json:"ia_modo"`
	FinalizacaoAutomatica   bool     `json:"finalizacao_automatica"`
	AutofimModo             string   `json:"autofim_modo"`
	ParceiroMonitoramento   bool     `json:"parceiro_monitoramento"`
	ParceiroModo            string   `json:"parceiro_modo"`
	EmailAtivo              bool     `json:"email_ativo"`
	EmailModo               string   `json:"email_modo"`
	CoberturaHoraria        string   `json:"cobertura_horaria"`
	IACoberturaHoraria      string   `json:"ia_cobertura_horaria"`
	ParceiroCoberturaHoraria string  `json:"parceiro_cobertura_horaria"`
	IDParceiro              string   `json:"id_parceiro"`
	GruposEventoParceiro    []string `json:"grupos_evento_parceiro"`
}

type ClienteAtendimentoConfig struct {
	ID                      int      `json:"id"`
	IDFranqueado            string   `json:"id_franqueado"`
	IDCliente               string   `json:"id_cliente"`
	InteligenciaArtificial  bool     `json:"inteligencia_artificial"`
	FinalizacaoAutomatica   bool     `json:"finalizacao_automatica"`
	ParceiroMonitoramento   bool     `json:"parceiro_monitoramento"`
	EmailAtivo              bool     `json:"email_ativo"`
	CoberturaHoraria        string   `json:"cobertura_horaria"`
	IDParceiro              string   `json:"id_parceiro"`
	GruposEventoParceiro    []string `json:"grupos_evento_parceiro"`
	Ativo                   bool     `json:"ativo"`
}

type AtendimentoListaItem struct {
	ID          int    `json:"id"`
	IDFranqueado string `json:"id_franqueado"`
	Recurso     string `json:"recurso"`
	IDCliente   string `json:"id_cliente"`
	NomeCliente string `json:"nome_cliente"`
}

type IABloqueioCliente struct {
	ID          int    `json:"id"`
	IDFranqueado string `json:"id_franqueado"`
	IDCliente   string `json:"id_cliente"`
	NomeCliente string `json:"nome_cliente"`
	Motivo      string `json:"motivo"`
}

type CreditoSaldo struct {
	IDFranqueado  string  `json:"id_franqueado"`
	Canal         string  `json:"canal"`
	Saldo         float64 `json:"saldo"`
	SaldoInicial  float64 `json:"saldo_inicial"`
	PercentualRestante float64 `json:"percentual_restante"`
}

type CreditoMovimento struct {
	ID           int64   `json:"id"`
	CreatedAt    string  `json:"created_at"`
	IDFranqueado string  `json:"id_franqueado"`
	Canal        string  `json:"canal"`
	Tipo         string  `json:"tipo"`
	Quantidade   float64 `json:"quantidade"`
	ValorUnitario *float64 `json:"valor_unitario"`
	ValorTotal   *float64 `json:"valor_total"`
	Observacao   string  `json:"observacao"`
}

func GetPoliticaAtendimento(ctx context.Context, idFranqueado string) (PoliticaAtendimento, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	out := PoliticaAtendimento{
		IDFranqueado: idFranqueado,
		IAModo:       ModoTodos,
		AutofimModo:  ModoTodos,
		ParceiroModo: ModoTodos,
		EmailModo:    ModoTodos,
		CoberturaHoraria: "24h",
		IACoberturaHoraria: "24h",
		ParceiroCoberturaHoraria: "24h",
	}
	if idFranqueado == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return out, err
	}
	var idParceiro sql.NullString
	var gruposJSON sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT inteligencia_artificial, ia_modo,
       finalizacao_automatica, autofim_modo,
       parceiro_monitoramento, parceiro_modo,
       email_ativo, email_modo,
       cobertura_horaria, COALESCE(ia_cobertura_horaria, '24h'), COALESCE(parceiro_cobertura_horaria, '24h'),
       id_parceiro,
       COALESCE(array_to_json(grupos_evento_parceiro)::text, '[]')
FROM ops_franqueado_atendimento_politica
WHERE id_franqueado = $1`, idFranqueado).Scan(
		&out.InteligenciaArtificial, &out.IAModo,
		&out.FinalizacaoAutomatica, &out.AutofimModo,
		&out.ParceiroMonitoramento, &out.ParceiroModo,
		&out.EmailAtivo, &out.EmailModo,
		&out.CoberturaHoraria, &out.IACoberturaHoraria, &out.ParceiroCoberturaHoraria, &idParceiro, &gruposJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, nil
		}
		return out, err
	}
	if idParceiro.Valid {
		out.IDParceiro = idParceiro.String
	}
	out.GruposEventoParceiro, err = decodeGruposJSON(gruposJSON)
	return out, err
}

func SavePoliticaAtendimento(ctx context.Context, p PoliticaAtendimento) error {
	p.IDFranqueado = strings.TrimSpace(p.IDFranqueado)
	if p.IDFranqueado == "" {
		return errors.New("id_franqueado obrigatorio")
	}
	normModo := func(m string) string {
		m = strings.TrimSpace(m)
		switch m {
		case ModoTodosMenos, ModoSomente:
			return m
		default:
			return ModoTodos
		}
	}
	p.IAModo = normModo(p.IAModo)
	p.AutofimModo = normModo(p.AutofimModo)
	p.ParceiroModo = normModo(p.ParceiroModo)
	p.EmailModo = normModo(p.EmailModo)
	if p.CoberturaHoraria == "" {
		p.CoberturaHoraria = "24h"
	}
	if p.IACoberturaHoraria == "" {
		p.IACoberturaHoraria = p.CoberturaHoraria
	}
	if p.ParceiroCoberturaHoraria == "" {
		p.ParceiroCoberturaHoraria = p.CoberturaHoraria
	}
	gruposJSON, err := encodeGruposJSON(p.GruposEventoParceiro)
	if err != nil {
		return err
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_franqueado_atendimento_politica (
    id_franqueado, inteligencia_artificial, ia_modo,
    finalizacao_automatica, autofim_modo,
    parceiro_monitoramento, parceiro_modo,
    email_ativo, email_modo,
    cobertura_horaria, ia_cobertura_horaria, parceiro_cobertura_horaria,
    id_parceiro, grupos_evento_parceiro, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
    COALESCE(ARRAY(SELECT json_array_elements_text($14::json)), ARRAY[]::text[]), NOW())
ON CONFLICT (id_franqueado) DO UPDATE SET
    inteligencia_artificial = EXCLUDED.inteligencia_artificial,
    ia_modo = EXCLUDED.ia_modo,
    finalizacao_automatica = EXCLUDED.finalizacao_automatica,
    autofim_modo = EXCLUDED.autofim_modo,
    parceiro_monitoramento = EXCLUDED.parceiro_monitoramento,
    parceiro_modo = EXCLUDED.parceiro_modo,
    email_ativo = EXCLUDED.email_ativo,
    email_modo = EXCLUDED.email_modo,
    cobertura_horaria = EXCLUDED.cobertura_horaria,
    ia_cobertura_horaria = EXCLUDED.ia_cobertura_horaria,
    parceiro_cobertura_horaria = EXCLUDED.parceiro_cobertura_horaria,
    id_parceiro = EXCLUDED.id_parceiro,
    grupos_evento_parceiro = EXCLUDED.grupos_evento_parceiro,
    updated_at = NOW()`,
		p.IDFranqueado, p.InteligenciaArtificial, p.IAModo,
		p.FinalizacaoAutomatica, p.AutofimModo,
		p.ParceiroMonitoramento, p.ParceiroModo,
		p.EmailAtivo, p.EmailModo,
		p.CoberturaHoraria, p.IACoberturaHoraria, p.ParceiroCoberturaHoraria,
		strings.TrimSpace(p.IDParceiro), gruposJSON,
	)
	return err
}

func ListAtendimentoLista(ctx context.Context, idFranqueado, recurso string) ([]AtendimentoListaItem, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	recurso = strings.TrimSpace(recurso)
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, id_franqueado, recurso, id_cliente, COALESCE(nome_cliente,'')
FROM ops_franqueado_atendimento_lista
WHERE id_franqueado = $1 AND recurso = $2
ORDER BY nome_cliente, id_cliente`, idFranqueado, recurso)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AtendimentoListaItem
	for rows.Next() {
		var it AtendimentoListaItem
		if err := rows.Scan(&it.ID, &it.IDFranqueado, &it.Recurso, &it.IDCliente, &it.NomeCliente); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func AddAtendimentoLista(ctx context.Context, idFranqueado, recurso, idCliente, nomeCliente string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	recurso = strings.TrimSpace(recurso)
	idCliente = strings.TrimSpace(idCliente)
	if idFranqueado == "" || recurso == "" || idCliente == "" {
		return errors.New("id_franqueado, recurso e id_cliente obrigatorios")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_franqueado_atendimento_lista (id_franqueado, recurso, id_cliente, nome_cliente)
VALUES ($1,$2,$3,$4)
ON CONFLICT DO NOTHING`,
		idFranqueado, recurso, idCliente, strings.TrimSpace(nomeCliente),
	)
	return err
}

func RemoveAtendimentoLista(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("id obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `DELETE FROM ops_franqueado_atendimento_lista WHERE id = $1`, id)
	return err
}

func ListIABloqueios(ctx context.Context, idFranqueado string) ([]IABloqueioCliente, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, id_franqueado, id_cliente, COALESCE(nome_cliente,''), COALESCE(motivo,'')
FROM ops_ia_bloqueio_cliente
WHERE id_franqueado = $1
ORDER BY nome_cliente, id_cliente`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []IABloqueioCliente
	for rows.Next() {
		var it IABloqueioCliente
		if err := rows.Scan(&it.ID, &it.IDFranqueado, &it.IDCliente, &it.NomeCliente, &it.Motivo); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func AddIABloqueio(ctx context.Context, idFranqueado, idCliente, nomeCliente, motivo string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idFranqueado == "" || idCliente == "" {
		return errors.New("id_franqueado e id_cliente obrigatorios")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_ia_bloqueio_cliente (id_franqueado, id_cliente, nome_cliente, motivo)
VALUES ($1,$2,$3,$4)
ON CONFLICT (id_franqueado, id_cliente) DO UPDATE SET
    nome_cliente = EXCLUDED.nome_cliente,
    motivo = EXCLUDED.motivo`,
		idFranqueado, idCliente, strings.TrimSpace(nomeCliente), strings.TrimSpace(motivo),
	)
	return err
}

func RemoveIABloqueio(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("id obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `DELETE FROM ops_ia_bloqueio_cliente WHERE id = $1`, id)
	return err
}

func IsIABloqueado(ctx context.Context, idFranqueado, idCliente string) (bool, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idFranqueado == "" || idCliente == "" {
		return false, nil
	}
	db, err := DB()
	if err != nil {
		return false, err
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_ia_bloqueio_cliente
WHERE id_franqueado = $1 AND id_cliente = $2`, idFranqueado, idCliente).Scan(&n)
	return n > 0, err
}

func ListCreditoSaldos(ctx context.Context, idFranqueado string) ([]CreditoSaldo, error) {
	r, err := GetCreditoResumo(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	return []CreditoSaldo{r}, nil
}

func GetCreditoResumo(ctx context.Context, idFranqueado string) (CreditoSaldo, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	out := CreditoSaldo{IDFranqueado: idFranqueado, Canal: CanalCredito, PercentualRestante: 100}
	if idFranqueado == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	saldo, ini, err := getSaldoWallet(ctx, idFranqueado)
	if err != nil {
		return out, err
	}
	out.Saldo = saldo
	out.SaldoInicial = ini
	out.PercentualRestante = percentualSaldo(saldo, ini)
	return out, nil
}

func percentualSaldo(saldo, ini float64) float64 {
	if ini <= 0 {
		if saldo > 0 {
			return 100
		}
		return 0
	}
	pct := (saldo / ini) * 100
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

func getSaldoWallet(ctx context.Context, idFranqueado string) (saldo, saldoIni float64, err error) {
	db, err := DB()
	if err != nil {
		return 0, 0, err
	}
	err = db.QueryRowContext(ctx, `
SELECT COALESCE(saldo,0), COALESCE(saldo_inicial,0)
FROM ops_credito_saldo WHERE id_franqueado = $1 AND canal = $2`,
		idFranqueado, CanalCredito).Scan(&saldo, &saldoIni)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, nil
	}
	return saldo, saldoIni, err
}

func ListCreditoMovimentos(ctx context.Context, idFranqueado, canal, de, ate string, limit int) ([]CreditoMovimento, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	q := `
SELECT id, created_at, id_franqueado, canal, tipo, quantidade,
       valor_unitario, valor_total, COALESCE(observacao,'')
FROM ops_credito_movimento
WHERE id_franqueado = $1`
	args := []any{strings.TrimSpace(idFranqueado)}
	n := 2
	if canal = strings.TrimSpace(canal); canal != "" {
		q += fmt.Sprintf(" AND canal = $%d", n)
		args = append(args, canal)
		n++
	}
	if de = strings.TrimSpace(de); de != "" {
		q += fmt.Sprintf(" AND created_at >= $%d::timestamptz", n)
		args = append(args, de)
		n++
	}
	if ate = strings.TrimSpace(ate); ate != "" {
		q += fmt.Sprintf(" AND created_at <= $%d::timestamptz", n)
		args = append(args, ate)
		n++
	}
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", n)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CreditoMovimento
	for rows.Next() {
		var m CreditoMovimento
		var created time.Time
		var vu, vt sql.NullFloat64
		if err := rows.Scan(&m.ID, &created, &m.IDFranqueado, &m.Canal, &m.Tipo, &m.Quantidade, &vu, &vt, &m.Observacao); err != nil {
			return nil, err
		}
		m.CreatedAt = created.UTC().Format(time.RFC3339)
		if vu.Valid {
			v := vu.Float64
			m.ValorUnitario = &v
		}
		if vt.Valid {
			v := vt.Float64
			m.ValorTotal = &v
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ResolveRecursoAtivo aplica politica + lista + config por cliente.
func ResolveRecursoAtivo(ctx context.Context, idFranqueado, idCliente, recurso string, politica *PoliticaAtendimento) (bool, error) {
	if bloqueado, err := IsIABloqueado(ctx, idFranqueado, idCliente); err != nil {
		return false, err
	} else if bloqueado && recurso == RecursoIA {
		return false, nil
	}

	var p PoliticaAtendimento
	var err error
	if politica != nil {
		p = *politica
	} else {
		p, err = GetPoliticaAtendimento(ctx, idFranqueado)
		if err != nil {
			return false, err
		}
	}

	ativoGlobal := false
	modo := ModoTodos
	switch recurso {
	case RecursoIA:
		ativoGlobal = p.InteligenciaArtificial
		modo = p.IAModo
	case RecursoAutofim:
		ativoGlobal = p.FinalizacaoAutomatica
		modo = p.AutofimModo
	case RecursoParceiro:
		ativoGlobal = p.ParceiroMonitoramento
		modo = p.ParceiroModo
	case RecursoEmail:
		ativoGlobal = p.EmailAtivo
		modo = p.EmailModo
	default:
		return false, nil
	}
	if !ativoGlobal {
		return false, nil
	}
	if modo == ModoTodos {
		return true, nil
	}

	db, err := DB()
	if err != nil {
		return false, err
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_franqueado_atendimento_lista
WHERE id_franqueado = $1 AND recurso = $2 AND id_cliente = $3`,
		idFranqueado, recurso, idCliente).Scan(&n)
	if err != nil {
		return false, err
	}
	naLista := n > 0
	switch modo {
	case ModoSomente:
		return naLista, nil
	case ModoTodosMenos:
		return !naLista, nil
	default:
		return true, nil
	}
}

func franqueadoTemPoliticaAtendimento(ctx context.Context, idFranqueado string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var ok bool
	err = db.QueryRowContext(ctx, `
SELECT EXISTS(SELECT 1 FROM ops_franqueado_atendimento_politica WHERE id_franqueado = $1)`,
		idFranqueado).Scan(&ok)
	return ok, err
}

func clienteTemConfigAtendimento(ctx context.Context, idFranqueado, idCliente string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var ok bool
	err = db.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1 FROM ops_cliente_atendimento_config
  WHERE id_franqueado = $1 AND id_cliente = $2 AND ativo = TRUE
)`, idFranqueado, idCliente).Scan(&ok)
	return ok, err
}

func resolveFromClienteConfig(ctx context.Context, idFranqueado, idCliente, recurso string) (bool, error) {
	if bloqueado, err := IsIABloqueado(ctx, idFranqueado, idCliente); err != nil {
		return false, err
	} else if bloqueado && recurso == RecursoIA {
		return false, nil
	}
	db, err := DB()
	if err != nil {
		return false, err
	}
	col := ""
	switch recurso {
	case RecursoIA:
		col = "inteligencia_artificial"
	case RecursoAutofim:
		col = "finalizacao_automatica"
	case RecursoParceiro:
		col = "parceiro_monitoramento"
	case RecursoEmail:
		col = "email_ativo"
	default:
		return false, nil
	}
	var ativo bool
	q := fmt.Sprintf(`
SELECT %s FROM ops_cliente_atendimento_config
WHERE id_franqueado = $1 AND id_cliente = $2 AND ativo = TRUE`, col)
	err = db.QueryRowContext(ctx, q, idFranqueado, idCliente).Scan(&ativo)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	return ativo, err
}

// ResolveRecursoAtivoCompat fail-open sem politica/config (compat legado #22 e sidecar).
func ResolveRecursoAtivoCompat(ctx context.Context, idFranqueado, idCliente, recurso string) (ativo bool, failOpen bool, err error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	recurso = strings.TrimSpace(recurso)
	if idFranqueado == "" || idCliente == "" || recurso == "" {
		return true, true, nil
	}

	hasPol, err := franqueadoTemPoliticaAtendimento(ctx, idFranqueado)
	if err != nil {
		return true, true, err
	}
	hasCli, err := clienteTemConfigAtendimento(ctx, idFranqueado, idCliente)
	if err != nil {
		return true, true, err
	}
	if !hasPol && !hasCli {
		return true, true, nil
	}
	if !hasPol {
		ok, err := resolveFromClienteConfig(ctx, idFranqueado, idCliente, recurso)
		return ok, false, err
	}
	ok, err := ResolveRecursoAtivo(ctx, idFranqueado, idCliente, recurso, nil)
	return ok, false, err
}

const (
	TipoMovBonificacao = "bonificacao"
	TipoMovRecarga     = "recarga"
	TipoMovDebitoUso   = "debito_uso"

	CanalCredito   = "credito"
	CanalLigacao   = "ligacao"
	CanalSMS       = "sms"
	CanalWhatsApp  = "whatsapp"
	CanalEmail     = "email"
)

type TarifaOperacional struct {
	IDCentral      string  `json:"id_central"`
	Canal          string  `json:"canal"`
	ValorTentativa float64 `json:"valor_tentativa"`
	ValorMinuto    float64 `json:"valor_minuto"`
	ValorUnidade   float64 `json:"valor_unidade"`
}

func GetTarifaOperacional(ctx context.Context, idCentral, canal string) (TarifaOperacional, error) {
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		idCentral = "CENTRAL"
	}
	canal = strings.TrimSpace(canal)
	if canal == "" {
		canal = CanalLigacao
	}
	out := TarifaOperacional{IDCentral: idCentral, Canal: canal}
	db, err := DB()
	if err != nil {
		return out, err
	}
	err = db.QueryRowContext(ctx, `
SELECT COALESCE(valor_tentativa,0), COALESCE(valor_minuto,0), COALESCE(valor_unidade,0)
FROM ops_tarifa_operacional WHERE id_central = $1 AND canal = $2`,
		idCentral, canal).Scan(&out.ValorTentativa, &out.ValorMinuto, &out.ValorUnidade)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	return out, err
}

func SaveTarifaOperacional(ctx context.Context, t TarifaOperacional) error {
	db, err := DB()
	if err != nil {
		return err
	}
	idCentral := strings.TrimSpace(t.IDCentral)
	if idCentral == "" {
		idCentral = "CENTRAL"
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_tarifa_operacional (id_central, canal, valor_tentativa, valor_minuto, valor_unidade, updated_at)
VALUES ($1,$2,$3,$4,$5,NOW())
ON CONFLICT (id_central, canal) DO UPDATE SET
  valor_tentativa = EXCLUDED.valor_tentativa,
  valor_minuto = EXCLUDED.valor_minuto,
  valor_unidade = EXCLUDED.valor_unidade,
  updated_at = NOW()`, idCentral, t.Canal, t.ValorTentativa, t.ValorMinuto, t.ValorUnidade)
	return err
}

func GetSaldoCanal(ctx context.Context, idFranqueado, _ string) (float64, error) {
	saldo, _, err := getSaldoWallet(ctx, idFranqueado)
	return saldo, err
}

func CustoMinimoCanal(ctx context.Context, idCentral, canal string) (float64, error) {
	t, err := GetTarifaOperacional(ctx, idCentral, canal)
	if err != nil {
		return 0, err
	}
	min := t.ValorTentativa
	if min <= 0 && t.ValorUnidade > 0 {
		min = t.ValorUnidade
	}
	if min <= 0 {
		min = 0.01
	}
	return min, nil
}

// PodeUsarCanal verifica saldo total >= custo minimo do servico.
func PodeUsarCanal(ctx context.Context, idFranqueado, idCentral, canal string) (bool, float64, float64, error) {
	saldo, _, err := getSaldoWallet(ctx, idFranqueado)
	if err != nil {
		return false, 0, 0, err
	}
	min, err := CustoMinimoCanal(ctx, idCentral, canal)
	if err != nil {
		return false, saldo, 0, err
	}
	return saldo >= min, saldo, min, nil
}

func movimentoCanal(servicoCanal, tipo string) string {
	servicoCanal = strings.TrimSpace(servicoCanal)
	tipo = strings.TrimSpace(tipo)
	if tipo == TipoMovRecarga || tipo == TipoMovBonificacao {
		return CanalCredito
	}
	if servicoCanal == "" || servicoCanal == CanalCredito {
		return CanalCredito
	}
	return servicoCanal
}

func CreditarSaldo(ctx context.Context, idFranqueado, servicoCanal, tipo string, valor float64, idFatura *int64, obs, criadoPor string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	tipo = strings.TrimSpace(tipo)
	if idFranqueado == "" || tipo == "" {
		return errors.New("id_franqueado e tipo obrigatorios")
	}
	if valor <= 0 {
		return errors.New("valor deve ser positivo")
	}
	movCanal := movimentoCanal(servicoCanal, tipo)
	db, err := DB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO ops_credito_saldo (id_franqueado, canal, saldo, saldo_inicial, updated_at)
VALUES ($1,$2,$3,$3,NOW())
ON CONFLICT (id_franqueado, canal) DO UPDATE SET
  saldo = ops_credito_saldo.saldo + EXCLUDED.saldo,
  updated_at = NOW()`, idFranqueado, CanalCredito, valor)
	if err != nil {
		return err
	}

	if tipo == TipoMovRecarga || tipo == TipoMovBonificacao {
		_, err = tx.ExecContext(ctx, `
UPDATE ops_credito_saldo SET saldo_inicial = saldo, updated_at = NOW()
WHERE id_franqueado = $1 AND canal = $2`, idFranqueado, CanalCredito)
		if err != nil {
			return err
		}
	}

	var idFat any
	if idFatura != nil && *idFatura > 0 {
		idFat = *idFatura
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO ops_credito_movimento
  (id_franqueado, canal, tipo, quantidade, valor_total, id_fatura, observacao, criado_por)
VALUES ($1,$2,$3,$4,$4,$5,$6,$7)`,
		idFranqueado, movCanal, tipo, valor, idFat, obs, criadoPor)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func DebitarSaldo(ctx context.Context, idFranqueado, servicoCanal string, valor float64, idProcesso, obs string) (jaDebitado bool, err error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	servicoCanal = strings.TrimSpace(servicoCanal)
	idProcesso = strings.TrimSpace(idProcesso)
	if idFranqueado == "" {
		return false, errors.New("id_franqueado obrigatorio")
	}
	if servicoCanal == "" {
		servicoCanal = CanalLigacao
	}
	if valor <= 0 {
		return false, errors.New("valor deve ser positivo")
	}
	db, err := DB()
	if err != nil {
		return false, err
	}
	if idProcesso != "" {
		var n int
		err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_credito_movimento
WHERE id_processo = $1 AND tipo = $2`, idProcesso, TipoMovDebitoUso).Scan(&n)
		if err != nil {
			return false, err
		}
		if n > 0 {
			return true, nil
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var saldo float64
	err = tx.QueryRowContext(ctx, `
SELECT COALESCE(saldo,0) FROM ops_credito_saldo
WHERE id_franqueado = $1 AND canal = $2 FOR UPDATE`, idFranqueado, CanalCredito).Scan(&saldo)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("sem saldo de credito")
	}
	if err != nil {
		return false, err
	}
	if saldo < valor {
		return false, fmt.Errorf("saldo insuficiente: %.4f < %.4f", saldo, valor)
	}
	_, err = tx.ExecContext(ctx, `
UPDATE ops_credito_saldo SET saldo = saldo - $3, updated_at = NOW()
WHERE id_franqueado = $1 AND canal = $2`, idFranqueado, CanalCredito, valor)
	if err != nil {
		return false, err
	}
	var idProc any
	if idProcesso != "" {
		idProc = idProcesso
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO ops_credito_movimento
  (id_franqueado, canal, tipo, quantidade, valor_total, id_processo, observacao, criado_por)
VALUES ($1,$2,$3,$4,$4,$5,$6,'sistema')`,
		idFranqueado, servicoCanal, TipoMovDebitoUso, valor, idProc, obs)
	if err != nil {
		return false, err
	}
	return false, tx.Commit()
}

func InicializarSaldoZero(ctx context.Context, idFranqueado string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil
	}
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_credito_saldo (id_franqueado, canal, saldo, saldo_inicial, updated_at)
VALUES ($1,$2,0,0,NOW())
ON CONFLICT (id_franqueado, canal) DO NOTHING`, idFranqueado, CanalCredito)
	return err
}
