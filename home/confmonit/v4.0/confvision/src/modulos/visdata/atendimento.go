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
	CoberturaHoraria          string   `json:"cobertura_horaria"`
	IACoberturaHoraria        string   `json:"ia_cobertura_horaria"`
	ParceiroCoberturaHoraria  string   `json:"parceiro_cobertura_horaria"`
	IDParceiro                string   `json:"id_parceiro"`
	GruposEventoParceiro      []string `json:"grupos_evento_parceiro"`
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
	}
	if idFranqueado == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return out, err
	}
	var idParceiro sql.NullString
	var grupos []string
	err = db.QueryRowContext(ctx, `
SELECT inteligencia_artificial, ia_modo,
       finalizacao_automatica, autofim_modo,
       parceiro_monitoramento, parceiro_modo,
       email_ativo, email_modo,
       cobertura_horaria, id_parceiro, grupos_evento_parceiro
FROM ops_franqueado_atendimento_politica
WHERE id_franqueado = $1`, idFranqueado).Scan(
		&out.InteligenciaArtificial, &out.IAModo,
		&out.FinalizacaoAutomatica, &out.AutofimModo,
		&out.ParceiroMonitoramento, &out.ParceiroModo,
		&out.EmailAtivo, &out.EmailModo,
		&out.CoberturaHoraria, &idParceiro, &grupos,
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
	out.GruposEventoParceiro = grupos
	return out, nil
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
	if p.GruposEventoParceiro == nil {
		p.GruposEventoParceiro = []string{}
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
    cobertura_horaria, id_parceiro, grupos_evento_parceiro, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULLIF($11,''),$12,NOW())
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
    id_parceiro = EXCLUDED.id_parceiro,
    grupos_evento_parceiro = EXCLUDED.grupos_evento_parceiro,
    updated_at = NOW()`,
		p.IDFranqueado, p.InteligenciaArtificial, p.IAModo,
		p.FinalizacaoAutomatica, p.AutofimModo,
		p.ParceiroMonitoramento, p.ParceiroModo,
		p.EmailAtivo, p.EmailModo,
		p.CoberturaHoraria, strings.TrimSpace(p.IDParceiro), p.GruposEventoParceiro,
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
	idFranqueado = strings.TrimSpace(idFranqueado)
	db, err := DB()
	if err != nil {
		return nil, err
	}
	canais := []string{"ligacao", "sms", "whatsapp", "email"}
	var out []CreditoSaldo
	for _, canal := range canais {
		var saldo, saldoIni float64
		err := db.QueryRowContext(ctx, `
SELECT COALESCE(saldo,0), COALESCE(saldo_inicial,0)
FROM ops_credito_saldo WHERE id_franqueado = $1 AND canal = $2`,
			idFranqueado, canal).Scan(&saldo, &saldoIni)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		pct := float64(100)
		if saldoIni > 0 {
			pct = (saldo / saldoIni) * 100
			if pct < 0 {
				pct = 0
			}
		}
		out = append(out, CreditoSaldo{
			IDFranqueado: idFranqueado, Canal: canal,
			Saldo: saldo, SaldoInicial: saldoIni, PercentualRestante: pct,
		})
	}
	return out, nil
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
