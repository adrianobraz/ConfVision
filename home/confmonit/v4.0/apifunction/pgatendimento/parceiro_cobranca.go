package pgatendimento

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	mysqldb "apifunction/db"
	"apifunction/config"
	"apifunction/pgcobranca"
	"apifunction/xano"
)

const (
	RefTipoParceiroConfig = "parceiro_config"

	StatusAtivPendente  = "pendente"
	StatusAtivPaga      = "paga"
	StatusAtivCancelada = "cancelada"

	MotivoSemCobranca = "sem_cobranca"
)

type AtivacaoRegistro struct {
	ID            int64   `json:"id"`
	IDFranqueado  string  `json:"id_franqueado"`
	IDParceiro    string  `json:"id_parceiro"`
	Status        string  `json:"status"`
	QtdClientes   int     `json:"qtd_clientes"`
	PrecoUnitario float64 `json:"preco_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	FPFaturaID    int     `json:"fp_fatura_id,omitempty"`
	CicloRef      string  `json:"ciclo_ref,omitempty"`
	PeriodoInicio string  `json:"periodo_inicio,omitempty"`
	PeriodoFim    string  `json:"periodo_fim,omitempty"`
	PagoAte       string  `json:"pago_ate,omitempty"`
}

type AtivacaoStatus struct {
	Liberado  bool              `json:"liberado"`
	Motivo    string            `json:"motivo"`
	Pendente  *AtivacaoRegistro `json:"pendente,omitempty"`
	PagoAte   string            `json:"pago_ate,omitempty"`
	QtdClientes int             `json:"qtd_clientes,omitempty"`
	ValorTotal  float64         `json:"valor_total,omitempty"`
}

type PrecoParceiroFranqueado struct {
	PrecoFranq    float64 `json:"preco_franqueado"`
	PisoParceiro  float64 `json:"piso_parceiro"`
	MargemCentral float64 `json:"margem_central"`
	MargemRep     float64 `json:"margem_rep"`
}

func dateOnlyBR(t time.Time) time.Time {
	loc := time.Local
	if sp, err := time.LoadLocation("America/Sao_Paulo"); err == nil && sp != nil {
		loc = sp
	}
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func roundMoney(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func franqueadoMeta(idFranqueado string) (idCentral, idRep string, err error) {
	if mysqldb.Conn == nil {
		return "", "", errors.New("mysql indisponivel")
	}
	var rep, central sql.NullString
	err = mysqldb.Conn.QueryRow(`
SELECT COALESCE(f.ID_Representante,''),
       COALESCE(NULLIF(TRIM(c.IDCentralUUID), ''), NULLIF(TRIM(c.ID_Central), ''), '')
FROM franqueado f
LEFT JOIN representante r ON f.ID_Representante = r.ID_Representante
LEFT JOIN central c ON r.IDCentralUUID = c.IDCentralUUID OR c.ID_Central = r.IDCentralUUID
WHERE f.ID_Franqueado = ?
LIMIT 1`, idFranqueado).Scan(&rep, &central)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(central.String), strings.TrimSpace(rep.String), nil
}

func lookupPrecoParceiro(ctx context.Context, idFranqueado, idParceiro string) (PrecoParceiroFranqueado, error) {
	out := PrecoParceiroFranqueado{}
	base := strings.TrimRight(strings.TrimSpace(config.ConfServiceURL), "/")
	if base == "" {
		return out, errors.New("CONFSERVICE_URL nao configurado")
	}
	q := url.Values{}
	q.Set("idFranqueado", idFranqueado)
	rawURL := base + "/internal/parceiros/catalogo?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("confservice HTTP %d", resp.StatusCode)
	}
	var parsed struct {
		Dados []map[string]any `json:"dados"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return out, err
	}
	for _, row := range parsed.Dados {
		id, _ := row["id"].(string)
		if strings.TrimSpace(id) != strings.TrimSpace(idParceiro) {
			continue
		}
		out.PrecoFranq = floatFromAny(row["precoClienteQuinzena"])
		out.PisoParceiro = floatFromAny(row["precoPisoParceiro"])
		out.MargemCentral = floatFromAny(row["margemCentral"])
		out.MargemRep = floatFromAny(row["margemRep"])
		if out.PrecoFranq <= 0 {
			return out, errors.New("preco parceiro invalido")
		}
		return out, nil
	}
	return out, errors.New("parceiro nao liberado para este franqueado")
}

func floatFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

// listClientesAtivosFranqueado lista clientes habilitados (mesma regra da tela Gerenciar Cliente):
// sem bloqueio em listaBloqueio (cliente, franqueado ou representante) e sem cancelamento.
func listClientesAtivosFranqueado(idFranqueado string) ([]string, error) {
	if mysqldb.Conn == nil {
		return nil, errors.New("mysql indisponivel")
	}
	rows, err := mysqldb.Conn.Query(`
SELECT c.ID_Cliente
FROM cliente c
INNER JOIN franqueado f ON f.ID_Franqueado = c.ID_Franqueado
LEFT JOIN listaBloqueio bloqCli ON bloqCli.ID_Alvo = c.ID_Cliente
LEFT JOIN listaBloqueio bloqFra ON bloqFra.ID_Alvo = f.ID_Franqueado
LEFT JOIN listaBloqueio bloqRep ON bloqRep.ID_Alvo = f.ID_Representante
WHERE c.ID_Franqueado = ?
  AND c.DataCancelamento IS NULL
  AND bloqCli.ID_Alvo IS NULL
  AND bloqFra.ID_Alvo IS NULL
  AND bloqRep.ID_Alvo IS NULL`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, rows.Err()
}

func CountClientesElegiveis(ctx context.Context, pol Politica) (int, error) {
	if !pol.ParceiroMonitoramento {
		return 0, nil
	}
	ids, err := listClientesAtivosFranqueado(pol.IDFranqueado)
	if err != nil {
		return 0, err
	}
	if pol.ParceiroModo == ModoTodos {
		return len(ids), nil
	}
	n := 0
	for _, idCli := range ids {
		ok, err := ResolveRecursoAtivo(ctx, pol.IDFranqueado, idCli, RecursoParceiro, &pol)
		if err != nil {
			return 0, err
		}
		if ok {
			n++
		}
	}
	return n, nil
}

func ClienteMonitoramentoAtivo(idFranqueado, idCliente string) (bool, error) {
	if mysqldb.Conn == nil {
		return true, nil
	}
	var bloqCli, bloqFra, bloqRep sql.NullString
	err := mysqldb.Conn.QueryRow(`
SELECT bloqCli.ID_Alvo, bloqFra.ID_Alvo, bloqRep.ID_Alvo
FROM cliente c
INNER JOIN franqueado f ON f.ID_Franqueado = c.ID_Franqueado
LEFT JOIN listaBloqueio bloqCli ON bloqCli.ID_Alvo = c.ID_Cliente
LEFT JOIN listaBloqueio bloqFra ON bloqFra.ID_Alvo = f.ID_Franqueado
LEFT JOIN listaBloqueio bloqRep ON bloqRep.ID_Alvo = f.ID_Representante
WHERE c.ID_Franqueado = ? AND c.ID_Cliente = ?
  AND c.DataCancelamento IS NULL
LIMIT 1`, idFranqueado, idCliente).Scan(&bloqCli, &bloqFra, &bloqRep)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !bloqCli.Valid && !bloqFra.Valid && !bloqRep.Valid, nil
}

func retornoSemCobranca(qtd int) AtivacaoStatus {
	return AtivacaoStatus{
		Liberado:    true,
		Motivo:      MotivoSemCobranca,
		QtdClientes: qtd,
	}
}

func scanAtivacao(row *sql.Row) (*AtivacaoRegistro, error) {
	var rec AtivacaoRegistro
	var fp sql.NullInt64
	var ini, fim, pago sql.NullTime
	err := row.Scan(
		&rec.ID, &rec.IDFranqueado, &rec.IDParceiro, &rec.Status,
		&rec.QtdClientes, &rec.PrecoUnitario, &rec.ValorTotal, &fp,
		&rec.CicloRef, &ini, &fim, &pago,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if fp.Valid {
		rec.FPFaturaID = int(fp.Int64)
	}
	if ini.Valid {
		rec.PeriodoInicio = ini.Time.Format("2006-01-02")
	}
	if fim.Valid {
		rec.PeriodoFim = fim.Time.Format("2006-01-02")
	}
	if pago.Valid {
		rec.PagoAte = pago.Time.Format("2006-01-02")
	}
	return &rec, nil
}

func GetAtivacaoPendente(ctx context.Context, idFranqueado string) (*AtivacaoRegistro, error) {
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	pend, err := scanAtivacao(d.QueryRowContext(ctx, `
SELECT id, id_franqueado, id_parceiro, status, qtd_clientes, preco_unitario, valor_total,
       fp_fatura_id, ciclo_ref, periodo_inicio, periodo_fim, pago_ate
FROM ops_parceiro_ativacao
WHERE id_franqueado = $1 AND status = $2
ORDER BY id DESC LIMIT 1`, idFranqueado, StatusAtivPendente))
	if err != nil || pend == nil {
		return pend, err
	}
	return reconciliarPendenteObsoleto(ctx, pend)
}

const metaTableFPFatura = 121

func lookupFaturaStatusXano(ctx context.Context, faturaID int) (string, error) {
	if faturaID <= 0 {
		return "", nil
	}
	cli := xano.NewMeta(config.XanoMetaBaseURL, config.XanoMetaAccessToken, config.XanoMetaWorkspaceID)
	if cli == nil || !cli.Enabled() {
		return "", nil
	}
	row, ok, err := cli.GetRecordByID(metaTableFPFatura, int64(faturaID))
	if err != nil || !ok {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(fmt.Sprint(row["status"]))), nil
}

func reconciliarPendenteObsoleto(ctx context.Context, pend *AtivacaoRegistro) (*AtivacaoRegistro, error) {
	if pend == nil || pend.FPFaturaID <= 0 {
		return pend, nil
	}
	st, err := lookupFaturaStatusXano(ctx, pend.FPFaturaID)
	if err != nil {
		return pend, err
	}
	if st == "" || st == "aberta" {
		return pend, nil
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	_, err = d.ExecContext(ctx, `
UPDATE ops_parceiro_ativacao SET status = $2, updated_at = NOW()
WHERE id = $1 AND status = $3`, pend.ID, StatusAtivCancelada, StatusAtivPendente)
	if err != nil {
		return pend, err
	}
	return nil, nil
}

func cancelarPendentes(ctx context.Context, idFranqueado string) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	_, err = d.ExecContext(ctx, `
UPDATE ops_parceiro_ativacao SET status = $2, updated_at = NOW()
WHERE id_franqueado = $1 AND status = $3`, idFranqueado, StatusAtivCancelada, StatusAtivPendente)
	return err
}

func servicoParceiroLiberado(ctx context.Context, idFranqueado, idParceiro string) (bool, string, error) {
	if idParceiro == "" {
		return false, "sem_parceiro", nil
	}
	srv, err := pgcobranca.GetServicoPorRef(ctx, idFranqueado, RefTipoParceiroConfig, idParceiro)
	if errors.Is(err, sql.ErrNoRows) {
		return true, "legado_sem_cobranca", nil
	}
	if err != nil {
		return false, "", err
	}
	if !srv.Ativo {
		return false, "servico_inativo", nil
	}
	if srv.PagoAte == nil {
		return false, "aguardando_pagamento", nil
	}
	hoje := dateOnlyBR(time.Now())
	if dateOnlyBR(*srv.PagoAte).Before(hoje) {
		return false, "periodo_expirado", nil
	}
	return true, "pago", nil
}

func ParceiroFinanceiroLiberado(ctx context.Context, idFranqueado string) (bool, string, error) {
	pend, err := GetAtivacaoPendente(ctx, idFranqueado)
	if err != nil {
		return false, "", err
	}
	if pend != nil {
		return false, "fatura_pendente", nil
	}
	pol, err := GetPolitica(ctx, idFranqueado)
	if err != nil {
		return false, "", err
	}
	if !pol.ParceiroMonitoramento {
		return false, "parceiro_desativado", nil
	}
	return servicoParceiroLiberado(ctx, idFranqueado, strings.TrimSpace(pol.IDParceiro))
}

func GetStatusAtivacao(ctx context.Context, idFranqueado string) (AtivacaoStatus, error) {
	out := AtivacaoStatus{}
	pol, err := GetPolitica(ctx, idFranqueado)
	if err != nil {
		return out, err
	}
	if !pol.ParceiroMonitoramento {
		out.Liberado = false
		out.Motivo = "parceiro_desativado"
		return out, nil
	}
	pend, err := GetAtivacaoPendente(ctx, idFranqueado)
	if err != nil {
		return out, err
	}
	if pend != nil {
		out.Liberado = false
		out.Motivo = "fatura_pendente"
		out.Pendente = pend
		out.QtdClientes = pend.QtdClientes
		out.ValorTotal = pend.ValorTotal
		return out, nil
	}
	ok, motivo, err := servicoParceiroLiberado(ctx, idFranqueado, strings.TrimSpace(pol.IDParceiro))
	if err != nil {
		return out, err
	}
	out.Liberado = ok
	out.Motivo = motivo
	srv, err := pgcobranca.GetServicoPorRef(ctx, idFranqueado, RefTipoParceiroConfig, strings.TrimSpace(pol.IDParceiro))
	if err == nil && srv != nil && srv.PagoAte != nil {
		out.PagoAte = srv.PagoAte.Format("2006-01-02")
	}
	qtd, _ := CountClientesElegiveis(ctx, pol)
	out.QtdClientes = qtd
	return out, nil
}

func syncFaturaAtivacao(ctx context.Context, idFranqueado, idCentral, idRep, cicloRef string, vencimento time.Time, qtd int, preco PrecoParceiroFranqueado, valorTotal float64, idParceiro string) (int, error) {
	cli := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	if cli == nil || !cli.Enabled() {
		return 0, nil
	}
	desc := fmt.Sprintf("Parceiro monitoramento — %d cliente(s) × R$ %.2f (quinzena)", qtd, preco.PrecoFranq)
	item := xano.OrdemItemSync{
		Descricao:     desc,
		Quantidade:    qtd,
		ValorUnitario: preco.PrecoFranq,
		ValorTotal:    valorTotal,
		RefTipo:       RefTipoParceiroConfig,
		RefID:         idParceiro,
		ValorPiso:     preco.PisoParceiro,
		MargemCentral: preco.MargemCentral,
		MargemRep:     preco.MargemRep,
	}
	return cli.SyncFaturaOrdem(xano.SyncOrdemInput{
		IDFranqueado:    idFranqueado,
		IDCentral:       idCentral,
		IDRepresentante: idRep,
		CicloRef:        cicloRef,
		VencimentoEm:    vencimento.Format("2006-01-02"),
		ValorTotal:      valorTotal,
		Itens:           []xano.OrdemItemSync{item},
	})
}

func ProporAtivacao(ctx context.Context, pol Politica) (AtivacaoStatus, error) {
	out := AtivacaoStatus{}
	pol.IDFranqueado = strings.TrimSpace(pol.IDFranqueado)
	if pol.IDFranqueado == "" {
		return out, errors.New("id_franqueado obrigatorio")
	}

	if !pol.ParceiroMonitoramento {
		_ = cancelarPendentes(ctx, pol.IDFranqueado)
		if idParceiro := strings.TrimSpace(pol.IDParceiro); idParceiro != "" {
			_ = pgcobranca.DesativarServicoPorRef(ctx, pol.IDFranqueado, RefTipoParceiroConfig, idParceiro)
		}
		out.Liberado = false
		out.Motivo = "parceiro_desativado"
		return out, nil
	}

	idParceiro := strings.TrimSpace(pol.IDParceiro)
	if idParceiro == "" {
		return out, errors.New("id_parceiro obrigatorio com parceiro ativo")
	}

	ok, motivo, err := servicoParceiroLiberado(ctx, pol.IDFranqueado, idParceiro)
	if err != nil {
		return out, err
	}
	if ok && motivo == "pago" {
		st, err := GetStatusAtivacao(ctx, pol.IDFranqueado)
		return st, err
	}

	pend, err := GetAtivacaoPendente(ctx, pol.IDFranqueado)
	if err != nil {
		return out, err
	}
	if pend != nil && pend.IDParceiro == idParceiro {
		out.Liberado = false
		out.Motivo = "fatura_pendente"
		out.Pendente = pend
		out.QtdClientes = pend.QtdClientes
		out.ValorTotal = pend.ValorTotal
		return out, nil
	}

	qtd, err := CountClientesElegiveis(ctx, pol)
	if err != nil {
		return out, err
	}
	if qtd <= 0 {
		return retornoSemCobranca(0), nil
	}

	preco, err := lookupPrecoParceiro(ctx, pol.IDFranqueado, idParceiro)
	if err != nil {
		return out, err
	}
	valorTotal := roundMoney(float64(qtd) * preco.PrecoFranq)
	if valorTotal <= 0 {
		return retornoSemCobranca(qtd), nil
	}

	idCentral, idRep, err := franqueadoMeta(pol.IDFranqueado)
	if err != nil {
		return out, err
	}

	hoje := dateOnlyBR(time.Now())
	fim := hoje.AddDate(0, 0, 14)
	cicloRef := fmt.Sprintf("PARCEIRO-%s-%s", pol.IDFranqueado, hoje.Format("20060102"))

	_ = cancelarPendentes(ctx, pol.IDFranqueado)

	d, err := db(ctx)
	if err != nil {
		return out, err
	}
	detalhe, _ := json.Marshal(map[string]any{
		"parceiro_modo": pol.ParceiroModo,
		"grupos":        pol.GruposEventoParceiro,
	})
	var ativID int64
	err = d.QueryRowContext(ctx, `
INSERT INTO ops_parceiro_ativacao (
    id_franqueado, id_parceiro, status, qtd_clientes, preco_unitario, valor_total,
    ciclo_ref, periodo_inicio, periodo_fim, detalhe, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,NOW())
RETURNING id`,
		pol.IDFranqueado, idParceiro, StatusAtivPendente, qtd, preco.PrecoFranq, valorTotal,
		cicloRef, hoje, fim, string(detalhe),
	).Scan(&ativID)
	if err != nil {
		return out, err
	}

	_, _ = pgcobranca.UpsertServico(ctx, pgcobranca.Servico{
		IDFranqueado:  pol.IDFranqueado,
		RefTipo:       RefTipoParceiroConfig,
		RefID:         idParceiro,
		Descricao:     "Parceiro de monitoramento (config)",
		DataInicio:    hoje,
		Periodicidade: pgcobranca.PeriodicidadeQuinzenal,
		ValorCiclo:    preco.PrecoFranq,
		DiasCiclo:     15,
		StatusCiclo:   pgcobranca.StatusCicloEmAjuste,
		ValorPiso:     preco.PisoParceiro,
		MargemCentral: preco.MargemCentral,
		MargemRep:     preco.MargemRep,
		Ativo:         true,
	})

	fpID, err := syncFaturaAtivacao(ctx, pol.IDFranqueado, idCentral, idRep, cicloRef, hoje, qtd, preco, valorTotal, idParceiro)
	if err != nil {
		return out, fmt.Errorf("falha ao gerar fatura: %w", err)
	}
	if fpID > 0 {
		_, _ = d.ExecContext(ctx, `
UPDATE ops_parceiro_ativacao SET fp_fatura_id = $2, updated_at = NOW() WHERE id = $1`, ativID, fpID)
	}

	rec := &AtivacaoRegistro{
		ID: ativID, IDFranqueado: pol.IDFranqueado, IDParceiro: idParceiro,
		Status: StatusAtivPendente, QtdClientes: qtd, PrecoUnitario: preco.PrecoFranq,
		ValorTotal: valorTotal, FPFaturaID: fpID, CicloRef: cicloRef,
		PeriodoInicio: hoje.Format("2006-01-02"), PeriodoFim: fim.Format("2006-01-02"),
	}
	out.Liberado = false
	out.Motivo = "fatura_pendente"
	out.Pendente = rec
	out.QtdClientes = qtd
	out.ValorTotal = valorTotal
	return out, nil
}

func CancelarAtivacaoPendente(ctx context.Context, idFranqueado string) error {
	pol, err := GetPolitica(ctx, idFranqueado)
	if err != nil {
		return err
	}
	_ = cancelarPendentes(ctx, idFranqueado)
	if idParceiro := strings.TrimSpace(pol.IDParceiro); idParceiro != "" {
		_ = pgcobranca.DesativarServicoPorRef(ctx, idFranqueado, RefTipoParceiroConfig, idParceiro)
	}
	return nil
}

func MarcarAtivacaoPaga(ctx context.Context, idFranqueado, refID string, faturaID int, pagoAte time.Time) error {
	_, err := MarcarAtivacaoPagaFatura(ctx, idFranqueado, faturaID, refID, pagoAte)
	return err
}

func LookupFranqueadoNomes(ids []string) map[string]string {
	out := map[string]string{}
	if mysqldb.Conn == nil || len(ids) == 0 {
		return out
	}
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		var nome sql.NullString
		if err := mysqldb.Conn.QueryRow(`SELECT COALESCE(NomeFantasia, RazaoSocial, '') FROM franqueado WHERE ID_Franqueado = ?`, id).Scan(&nome); err == nil {
			n := strings.TrimSpace(nome.String)
			if n != "" {
				out[id] = n
			}
		}
	}
	return out
}
