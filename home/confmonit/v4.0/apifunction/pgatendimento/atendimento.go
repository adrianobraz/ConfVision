package pgatendimento

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"apifunction/pgcredito"
	fin "apifunction/service/financeiro"
)

const (
	RecursoIA       = "inteligencia_artificial"
	RecursoAutofim  = "finalizacao_automatica"
	RecursoParceiro = "parceiro"
	RecursoEmail    = "email"

	ModoTodos      = "todos"
	ModoTodosMenos = "todos_menos"
	ModoSomente    = "somente"

	CanalCredito = "credito"
	CanalLigacao = "ligacao"
)

type Politica struct {
	IDFranqueado             string   `json:"id_franqueado"`
	InteligenciaArtificial   bool     `json:"inteligencia_artificial"`
	IAModo                   string   `json:"ia_modo"`
	FinalizacaoAutomatica    bool     `json:"finalizacao_automatica"`
	AutofimModo              string   `json:"autofim_modo"`
	ParceiroMonitoramento    bool     `json:"parceiro_monitoramento"`
	ParceiroModo             string   `json:"parceiro_modo"`
	EmailAtivo               bool     `json:"email_ativo"`
	EmailModo                string   `json:"email_modo"`
	CoberturaHoraria         string   `json:"cobertura_horaria"`
	IACoberturaHoraria       string   `json:"ia_cobertura_horaria"`
	ParceiroCoberturaHoraria string   `json:"parceiro_cobertura_horaria"`
	ParceiroDuplaComunicacao bool     `json:"parceiro_dupla_comunicacao"`
	IDParceiro               string   `json:"id_parceiro"`
	GruposEventoParceiro     []string `json:"grupos_evento_parceiro"`
}

func db(ctx context.Context) (*sql.DB, error) {
	return pgcredito.DB()
}

func GetPolitica(ctx context.Context, idFranqueado string) (Politica, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	out := Politica{
		IDFranqueado: idFranqueado, IAModo: ModoTodos, AutofimModo: ModoTodos,
		ParceiroModo: ModoTodos, EmailModo: ModoTodos, CoberturaHoraria: "24h",
		IACoberturaHoraria: "24h", ParceiroCoberturaHoraria: "24h",
	}
	if idFranqueado == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}
	d, err := db(ctx)
	if err != nil {
		return out, err
	}
	var idParceiro sql.NullString
	var gruposJSON sql.NullString
	err = d.QueryRowContext(ctx, `
SELECT inteligencia_artificial, ia_modo, finalizacao_automatica, autofim_modo,
       parceiro_monitoramento, parceiro_modo, email_ativo, email_modo,
       cobertura_horaria, COALESCE(ia_cobertura_horaria, '24h'), COALESCE(parceiro_cobertura_horaria, '24h'),
       COALESCE(parceiro_dupla_comunicacao, FALSE), id_parceiro,
       COALESCE(array_to_json(grupos_evento_parceiro)::text, '[]')
FROM ops_franqueado_atendimento_politica WHERE id_franqueado = $1`, idFranqueado).Scan(
		&out.InteligenciaArtificial, &out.IAModo, &out.FinalizacaoAutomatica, &out.AutofimModo,
		&out.ParceiroMonitoramento, &out.ParceiroModo, &out.EmailAtivo, &out.EmailModo,
		&out.CoberturaHoraria, &out.IACoberturaHoraria, &out.ParceiroCoberturaHoraria,
		&out.ParceiroDuplaComunicacao, &idParceiro, &gruposJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	if idParceiro.Valid {
		out.IDParceiro = idParceiro.String
	}
	out.GruposEventoParceiro, err = decodeGruposJSON(gruposJSON)
	return out, err
}

func IsIABloqueado(ctx context.Context, idFranqueado, idCliente string) (bool, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	if idFranqueado == "" || idCliente == "" {
		return false, nil
	}
	d, err := db(ctx)
	if err != nil {
		return false, err
	}
	var n int
	err = d.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_ia_bloqueio_cliente
WHERE id_franqueado = $1 AND id_cliente = $2`, idFranqueado, idCliente).Scan(&n)
	return n > 0, err
}

func ResolveRecursoAtivo(ctx context.Context, idFranqueado, idCliente, recurso string, politica *Politica) (bool, error) {
	if bloqueado, err := IsIABloqueado(ctx, idFranqueado, idCliente); err != nil {
		return false, err
	} else if bloqueado && recurso == RecursoIA {
		return false, nil
	}

	var p Politica
	var err error
	if politica != nil {
		p = *politica
	} else {
		p, err = GetPolitica(ctx, idFranqueado)
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

	d, err := db(ctx)
	if err != nil {
		return false, err
	}
	var n int
	err = d.QueryRowContext(ctx, `
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

func franqueadoTemPolitica(ctx context.Context, idFranqueado string) (bool, error) {
	d, err := db(ctx)
	if err != nil {
		return false, err
	}
	var ok bool
	err = d.QueryRowContext(ctx, `
SELECT EXISTS(SELECT 1 FROM ops_franqueado_atendimento_politica WHERE id_franqueado = $1)`,
		idFranqueado).Scan(&ok)
	return ok, err
}

func clienteTemConfig(ctx context.Context, idFranqueado, idCliente string) (bool, error) {
	d, err := db(ctx)
	if err != nil {
		return false, err
	}
	var ok bool
	err = d.QueryRowContext(ctx, `
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
	d, err := db(ctx)
	if err != nil {
		return false, err
	}
	var ativo bool
	q := fmt.Sprintf(`
SELECT %s FROM ops_cliente_atendimento_config
WHERE id_franqueado = $1 AND id_cliente = $2 AND ativo = TRUE`, col)
	err = d.QueryRowContext(ctx, q, idFranqueado, idCliente).Scan(&ativo)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	return ativo, err
}

func ResolveCompat(ctx context.Context, idFranqueado, idCliente, recurso, ctiGrupo string) (ativo bool, failOpen bool, err error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	recurso = strings.TrimSpace(recurso)
	if idFranqueado == "" || idCliente == "" || recurso == "" {
		return true, true, nil
	}

	hasPol, err := franqueadoTemPolitica(ctx, idFranqueado)
	if err != nil {
		return true, true, err
	}
	hasCli, err := clienteTemConfig(ctx, idFranqueado, idCliente)
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
	ok, err := ResolveComHorario(ctx, idFranqueado, idCliente, recurso, ctiGrupo, nil)
	return ok, false, err
}

func getSaldoWallet(ctx context.Context, idFranqueado string) (saldo, saldoIni float64, err error) {
	d, err := db(ctx)
	if err != nil {
		return 0, 0, err
	}
	err = d.QueryRowContext(ctx, `
SELECT COALESCE(saldo,0), COALESCE(saldo_inicial,0)
FROM ops_credito_saldo WHERE id_franqueado = $1 AND canal = $2`,
		idFranqueado, CanalCredito).Scan(&saldo, &saldoIni)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, nil
	}
	return saldo, saldoIni, err
}

func custoMinimoCanal(ctx context.Context, idCentral, canal string) (float64, error) {
	idCentral = strings.TrimSpace(idCentral)
	canal = strings.TrimSpace(canal)
	if canal == "" {
		canal = CanalLigacao
	}
	d, err := db(ctx)
	if err != nil {
		return 0, err
	}

	keys := map[string]struct{}{}
	if idCentral != "" {
		expanded, err := fin.ChavesCentral(ctx, idCentral)
		if err != nil {
			return 0, err
		}
		keys = expanded
	}
	if len(keys) == 0 && idCentral != "" {
		keys[idCentral] = struct{}{}
	}

	for key := range keys {
		var tentativa, minuto, unidade float64
		err = d.QueryRowContext(ctx, `
SELECT COALESCE(valor_tentativa,0), COALESCE(valor_minuto,0), COALESCE(valor_unidade,0)
FROM ops_tarifa_operacional WHERE id_central = $1 AND canal = $2`, key, canal).Scan(&tentativa, &minuto, &unidade)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return 0, err
		}
		min := tentativa
		if min <= 0 && unidade > 0 {
			min = unidade
		}
		if min <= 0 {
			min = 0.01
		}
		return min, nil
	}
	return 0.01, nil
}

func PodeLigar(ctx context.Context, idFranqueado, idCentral string) (bool, float64, float64, error) {
	saldo, _, err := getSaldoWallet(ctx, idFranqueado)
	if err != nil {
		return false, 0, 0, err
	}
	min, err := custoMinimoCanal(ctx, idCentral, CanalLigacao)
	if err != nil {
		return false, saldo, 0, err
	}
	return saldo >= min, saldo, min, nil
}

func SaveTarifa(ctx context.Context, idCentral, canal string, tentativa, minuto, unidade float64) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	idCentral = strings.TrimSpace(idCentral)
	if idCentral == "" {
		return errors.New("id_central obrigatorio (IDCentralUUID)")
	}
	keys, err := fin.ChavesCentral(ctx, idCentral)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		keys = map[string]struct{}{idCentral: {}}
	}
	for key := range keys {
		_, err = d.ExecContext(ctx, `
INSERT INTO ops_tarifa_operacional (id_central, canal, valor_tentativa, valor_minuto, valor_unidade, updated_at)
VALUES ($1,$2,$3,$4,$5,NOW())
ON CONFLICT (id_central, canal) DO UPDATE SET
  valor_tentativa = EXCLUDED.valor_tentativa,
  valor_minuto = EXCLUDED.valor_minuto,
  valor_unidade = EXCLUDED.valor_unidade,
  updated_at = NOW()`, key, canal, tentativa, minuto, unidade)
		if err != nil {
			return err
		}
	}
	return nil
}

func InicializarSaldoZero(ctx context.Context, idFranqueado string) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	_, err = d.ExecContext(ctx, `
INSERT INTO ops_credito_saldo (id_franqueado, canal, saldo, saldo_inicial, updated_at)
VALUES ($1,$2,0,0,NOW())
ON CONFLICT (id_franqueado, canal) DO NOTHING`, idFranqueado, CanalCredito)
	return err
}

func Creditar(ctx context.Context, idFranqueado, servico, tipo string, valor float64, idFatura *int64, obs, criadoPor string) error {
	return pgcredito.CreditarSaldo(ctx, idFranqueado, servico, tipo, valor, idFatura, obs, criadoPor)
}

func Debitar(ctx context.Context, idFranqueado, servico string, valor float64, idProcesso, obs string) (pgcredito.DebitarResult, error) {
	return pgcredito.DebitarSaldo(ctx, idFranqueado, servico, valor, idProcesso, obs)
}

func Configurado() bool {
	return pgcredito.Configurado()
}

// ResolveResult resposta compativel com ConfVision / #22.
type ResolveResult struct {
	Ativo      bool   `json:"ativo"`
	IABloqueado bool  `json:"ia_bloqueado"`
	Recurso    string `json:"recurso"`
	FailOpen   bool   `json:"fail_open"`
	CtiGrupo   string `json:"cti_grupo"`
	Emergencia bool   `json:"emergencia"`
}

func Resolve(ctx context.Context, idFranqueado, idCliente, recurso, ctiGrupo string) (ResolveResult, error) {
	ok, failOpen, err := ResolveCompat(ctx, idFranqueado, idCliente, recurso, ctiGrupo)
	if err != nil {
		return ResolveResult{}, err
	}
	bloq, _ := IsIABloqueado(ctx, idFranqueado, idCliente)
	return ResolveResult{
		Ativo: ok, IABloqueado: bloq, Recurso: recurso, FailOpen: failOpen,
		CtiGrupo: ctiGrupo, Emergencia: IsGrupoEmergencia(ctiGrupo),
	}, nil
}
