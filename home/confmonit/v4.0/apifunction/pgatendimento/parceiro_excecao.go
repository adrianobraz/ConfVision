package pgatendimento

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"apifunction/config"
	"apifunction/pgcobranca"
	"apifunction/xano"
)

const RefTipoParceiroExcecao = "parceiro_excecao"

type ExcecaoRegistro struct {
	ID                 int64   `json:"id"`
	IDFranqueado       string  `json:"id_franqueado"`
	IDCliente          string  `json:"id_cliente"`
	NomeCliente        string  `json:"nome_cliente,omitempty"`
	IDParceiro         string  `json:"id_parceiro"`
	IDParceiroAnterior string  `json:"id_parceiro_anterior,omitempty"`
	IDVinculoAnterior  string  `json:"id_vinculo_anterior,omitempty"`
	Status             string  `json:"status"`
	PrecoUnitario      float64 `json:"preco_unitario"`
	ValorTotal         float64 `json:"valor_total"`
	FPFaturaID         int     `json:"fp_fatura_id,omitempty"`
	CicloRef           string  `json:"ciclo_ref,omitempty"`
	PeriodoInicio      string  `json:"periodo_inicio,omitempty"`
	PeriodoFim         string  `json:"periodo_fim,omitempty"`
	PagoAte            string  `json:"pago_ate,omitempty"`
	IDVinculoAtivo     string  `json:"id_vinculo_ativo,omitempty"`
}

type ExcecaoStatus struct {
	TemPendente bool              `json:"tem_pendente"`
	Pendentes   []ExcecaoRegistro `json:"pendentes,omitempty"`
}

type vinculoClienteCS struct {
	ID         string
	IDParceiro string
}

func lookupVinculoClienteCS(ctx context.Context, idFranqueado, idCliente string) (*vinculoClienteCS, error) {
	base := strings.TrimRight(strings.TrimSpace(config.ConfServiceURL), "/")
	if base == "" {
		return nil, nil
	}
	q := url.Values{}
	q.Set("idFranqueado", idFranqueado)
	q.Set("idCliente", idCliente)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/internal/vinculo/cliente?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("confservice status=%d", resp.StatusCode)
	}
	var parsed struct {
		Vinculo *struct {
			ID         string `json:"id"`
			IDParceiro string `json:"idParceiro"`
		} `json:"vinculo"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.Vinculo == nil || strings.TrimSpace(parsed.Vinculo.ID) == "" {
		return nil, nil
	}
	return &vinculoClienteCS{
		ID:         strings.TrimSpace(parsed.Vinculo.ID),
		IDParceiro: strings.TrimSpace(parsed.Vinculo.IDParceiro),
	}, nil
}

func scanExcecao(row *sql.Row) (*ExcecaoRegistro, error) {
	var rec ExcecaoRegistro
	var fp sql.NullInt64
	var ini, fim, pago sql.NullTime
	var parAnt, vincAnt, vincAtivo sql.NullString
	err := row.Scan(
		&rec.ID, &rec.IDFranqueado, &rec.IDCliente, &rec.NomeCliente, &rec.IDParceiro,
		&parAnt, &vincAnt, &rec.Status, &rec.PrecoUnitario, &rec.ValorTotal, &fp,
		&rec.CicloRef, &ini, &fim, &pago, &vincAtivo,
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
	rec.IDParceiroAnterior = strings.TrimSpace(parAnt.String)
	rec.IDVinculoAnterior = strings.TrimSpace(vincAnt.String)
	rec.IDVinculoAtivo = strings.TrimSpace(vincAtivo.String)
	return &rec, nil
}

const excecaoSelectCols = `
SELECT id, id_franqueado, id_cliente, nome_cliente, id_parceiro,
       id_parceiro_anterior, id_vinculo_anterior, status, preco_unitario, valor_total,
       fp_fatura_id, ciclo_ref, periodo_inicio, periodo_fim, pago_ate, id_vinculo_ativo`

func reconciliarExcecaoPendente(ctx context.Context, pend *ExcecaoRegistro) (*ExcecaoRegistro, error) {
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
	if st == "paga" {
		_ = AtivarExcecaoPaga(ctx, pend.IDFranqueado, strconv.FormatInt(pend.ID, 10), pend.FPFaturaID, time.Now())
		return nil, nil
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	_, err = d.ExecContext(ctx, `
UPDATE ops_parceiro_excecao SET status = $2, updated_at = NOW()
WHERE id = $1 AND status = $3`, pend.ID, StatusAtivCancelada, StatusAtivPendente)
	if err != nil {
		return pend, err
	}
	_ = RegistrarHistoricoParceiro(ctx, HistoricoInput{
		IDFranqueado:   pend.IDFranqueado,
		IDCliente:      pend.IDCliente,
		NomeCliente:    pend.NomeCliente,
		Evento:         HistEventoExcecaoCancelada,
		IDParceiroDe:   pend.IDParceiroAnterior,
		IDParceiroPara: pend.IDParceiro,
		IDVinculoDe:    pend.IDVinculoAnterior,
		FPFaturaID:     pend.FPFaturaID,
		Detalhe:        map[string]any{"motivo": "fatura_" + st},
	})
	return nil, nil
}

// RepararExcecoesPagasPerdidas reativa exceções marcadas cancelada indevidamente com fatura paga.
func RepararExcecoesPagasPerdidas(ctx context.Context, idFranqueado string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil
	}
	d, err := db(ctx)
	if err != nil {
		return err
	}
	rows, err := d.QueryContext(ctx, excecaoSelectCols+`
FROM ops_parceiro_excecao
WHERE id_franqueado = $1 AND status = $2 AND fp_fatura_id IS NOT NULL AND fp_fatura_id > 0
ORDER BY id DESC`, idFranqueado, StatusAtivCancelada)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rec, err := scanExcecaoRow(rows)
		if err != nil || rec == nil {
			continue
		}
		st, err := lookupFaturaStatusXano(ctx, rec.FPFaturaID)
		if err != nil || st != "paga" {
			continue
		}
		_ = AtivarExcecaoPaga(ctx, rec.IDFranqueado, strconv.FormatInt(rec.ID, 10), rec.FPFaturaID, time.Now())
	}
	return rows.Err()
}

func scanExcecaoRow(rows *sql.Rows) (*ExcecaoRegistro, error) {
	var rec ExcecaoRegistro
	var fp sql.NullInt64
	var ini, fim, pago sql.NullTime
	var parAnt, vincAnt, vincAtivo sql.NullString
	if err := rows.Scan(
		&rec.ID, &rec.IDFranqueado, &rec.IDCliente, &rec.NomeCliente, &rec.IDParceiro,
		&parAnt, &vincAnt, &rec.Status, &rec.PrecoUnitario, &rec.ValorTotal, &fp,
		&rec.CicloRef, &ini, &fim, &pago, &vincAtivo,
	); err != nil {
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
	rec.IDParceiroAnterior = strings.TrimSpace(parAnt.String)
	rec.IDVinculoAnterior = strings.TrimSpace(vincAnt.String)
	rec.IDVinculoAtivo = strings.TrimSpace(vincAtivo.String)
	return &rec, nil
}

func GetExcecaoPendenteCliente(ctx context.Context, idFranqueado, idCliente string) (*ExcecaoRegistro, error) {
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	pend, err := scanExcecao(d.QueryRowContext(ctx, excecaoSelectCols+`
FROM ops_parceiro_excecao
WHERE id_franqueado = $1 AND id_cliente = $2 AND status = $3
ORDER BY id DESC LIMIT 1`, idFranqueado, idCliente, StatusAtivPendente))
	if err != nil || pend == nil {
		return pend, err
	}
	return reconciliarExcecaoPendente(ctx, pend)
}

func ListExcecoesPendentes(ctx context.Context, idFranqueado string) ([]ExcecaoRegistro, error) {
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.QueryContext(ctx, excecaoSelectCols+`
FROM ops_parceiro_excecao
WHERE id_franqueado = $1 AND status = $2
ORDER BY id DESC`, idFranqueado, StatusAtivPendente)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ExcecaoRegistro
	for rows.Next() {
		var rec ExcecaoRegistro
		var fp sql.NullInt64
		var ini, fim, pago sql.NullTime
		var parAnt, vincAnt, vincAtivo sql.NullString
		if err := rows.Scan(
			&rec.ID, &rec.IDFranqueado, &rec.IDCliente, &rec.NomeCliente, &rec.IDParceiro,
			&parAnt, &vincAnt, &rec.Status, &rec.PrecoUnitario, &rec.ValorTotal, &fp,
			&rec.CicloRef, &ini, &fim, &pago, &vincAtivo,
		); err != nil {
			continue
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
		rec.IDParceiroAnterior = strings.TrimSpace(parAnt.String)
		rec.IDVinculoAnterior = strings.TrimSpace(vincAnt.String)
		rec.IDVinculoAtivo = strings.TrimSpace(vincAtivo.String)
		fixed, _ := reconciliarExcecaoPendente(ctx, &rec)
		if fixed != nil {
			out = append(out, *fixed)
		}
	}
	return out, rows.Err()
}

// RepararPagoAteExcecoesPagas sincroniza pago_ate em fp_servico_cobranca (cs_parceiro_vinculo)
// para exceções já pagas cujo vínculo ficou sem data (deploy anterior ao hook).
func RepararPagoAteExcecoesPagas(ctx context.Context, idFranqueado string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil
	}
	d, err := db(ctx)
	if err != nil {
		return err
	}
	rows, err := d.QueryContext(ctx, excecaoSelectCols+`
FROM ops_parceiro_excecao
WHERE id_franqueado = $1 AND status = $2
  AND id_vinculo_ativo IS NOT NULL AND TRIM(id_vinculo_ativo) <> ''
ORDER BY id DESC`, idFranqueado, StatusAtivPaga)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		rec, err := scanExcecaoRow(rows)
		if err != nil || rec == nil {
			continue
		}
		_ = syncPagoAteVinculoExcecao(ctx, rec, time.Time{})
	}
	return rows.Err()
}

func parsePagoAteExcecao(rec *ExcecaoRegistro, fallback time.Time) time.Time {
	if rec == nil {
		return dateOnlyBR(fallback)
	}
	if s := strings.TrimSpace(rec.PagoAte); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return dateOnlyBR(t)
		}
	}
	if s := strings.TrimSpace(rec.PeriodoFim); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return dateOnlyBR(t)
		}
	}
	fb := dateOnlyBR(fallback)
	if fb.IsZero() {
		fb = dateOnlyBR(time.Now())
	}
	return fb.AddDate(0, 0, 14)
}

func syncPagoAteVinculoExcecao(ctx context.Context, rec *ExcecaoRegistro, pagoAte time.Time) error {
	if rec == nil {
		return nil
	}
	idVinculo := strings.TrimSpace(rec.IDVinculoAtivo)
	if idVinculo == "" {
		return nil
	}
	if pagoAte.IsZero() {
		pagoAte = parsePagoAteExcecao(rec, time.Now())
	} else {
		pagoAte = dateOnlyBR(pagoAte)
	}
	return pgcobranca.AtualizarPagoAtePorRef(ctx, rec.IDFranqueado, "cs_parceiro_vinculo", idVinculo, pagoAte)
}

func GetExcecaoStatus(ctx context.Context, idFranqueado string) (ExcecaoStatus, error) {
	out := ExcecaoStatus{}
	_ = RepararExcecoesPagasPerdidas(ctx, idFranqueado)
	_ = RepararPagoAteExcecoesPagas(ctx, idFranqueado)
	list, err := ListExcecoesPendentes(ctx, idFranqueado)
	if err != nil {
		return out, err
	}
	out.Pendentes = list
	out.TemPendente = len(list) > 0
	return out, nil
}

func cancelarExcecaoPendenteCliente(ctx context.Context, idFranqueado, idCliente string) error {
	d, err := db(ctx)
	if err != nil {
		return err
	}
	_, err = d.ExecContext(ctx, `
UPDATE ops_parceiro_excecao SET status = $3, updated_at = NOW()
WHERE id_franqueado = $1 AND id_cliente = $2 AND status = $4`,
		idFranqueado, idCliente, StatusAtivCancelada, StatusAtivPendente)
	return err
}

func syncFaturaExcecao(ctx context.Context, idFranqueado, idCentral, idRep, cicloRef string, vencimento time.Time, preco PrecoParceiroFranqueado, valorTotal float64, excecaoID int64, nomeCliente string) (int, error) {
	cli := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	if cli == nil || !cli.Enabled() {
		return 0, nil
	}
	nome := strings.TrimSpace(nomeCliente)
	if nome == "" {
		nome = "cliente"
	}
	desc := fmt.Sprintf("Exceção parceiro — %s (quinzena)", nome)
	item := xano.OrdemItemSync{
		Descricao:     desc,
		Quantidade:    1,
		ValorUnitario: preco.PrecoFranq,
		ValorTotal:    valorTotal,
		RefTipo:       RefTipoParceiroExcecao,
		RefID:         strconv.FormatInt(excecaoID, 10),
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

type ProporExcecaoInput struct {
	IDFranqueado string
	IDCliente    string
	NomeCliente  string
	IDParceiro   string
}

func ProporExcecao(ctx context.Context, in ProporExcecaoInput) (ExcecaoRegistro, error) {
	out := ExcecaoRegistro{}
	in.IDFranqueado = strings.TrimSpace(in.IDFranqueado)
	in.IDCliente = strings.TrimSpace(in.IDCliente)
	in.IDParceiro = strings.TrimSpace(in.IDParceiro)
	in.NomeCliente = strings.TrimSpace(in.NomeCliente)
	if in.IDFranqueado == "" || in.IDCliente == "" || in.IDParceiro == "" {
		return out, errors.New("id_franqueado, id_cliente e id_parceiro obrigatorios")
	}

	liberado, motivo, err := ParceiroFinanceiroLiberado(ctx, in.IDFranqueado)
	if err != nil {
		return out, err
	}
	if !liberado {
		switch motivo {
		case "fatura_pendente":
			return out, errors.New("fatura de ativacao pendente — pague ou cancele em Configuracao de Atendimento")
		case "parceiro_desativado":
			return out, errors.New("ative o parceiro em Configuracao de Atendimento")
		default:
			return out, fmt.Errorf("parceiro nao liberado: %s", motivo)
		}
	}

	pol, err := GetPolitica(ctx, in.IDFranqueado)
	if err != nil {
		return out, err
	}
	if strings.TrimSpace(pol.IDParceiro) == in.IDParceiro {
		return out, errors.New("parceiro igual ao padrao da configuracao — excecao so para parceiro diferente")
	}

	okCli, err := ClienteMonitoramentoAtivo(in.IDFranqueado, in.IDCliente)
	if err != nil {
		return out, err
	}
	if !okCli {
		return out, errors.New("cliente inativo ou bloqueado")
	}

	pend, err := GetExcecaoPendenteCliente(ctx, in.IDFranqueado, in.IDCliente)
	if err != nil {
		return out, err
	}
	if pend != nil {
		out = *pend
		return out, nil
	}

	vincAnt, err := lookupVinculoClienteCS(ctx, in.IDFranqueado, in.IDCliente)
	if err != nil {
		return out, err
	}
	var idVinculoAnt, idParceiroAnt string
	if vincAnt != nil {
		idVinculoAnt = vincAnt.ID
		idParceiroAnt = vincAnt.IDParceiro
		if idParceiroAnt == in.IDParceiro {
			return out, errors.New("cliente ja vinculado a este parceiro")
		}
	}

	preco, err := lookupPrecoParceiro(ctx, in.IDFranqueado, in.IDParceiro)
	if err != nil {
		return out, err
	}
	valorTotal := roundMoney(preco.PrecoFranq)
	if valorTotal <= 0 {
		return out, errors.New("preco da excecao invalido")
	}

	idCentral, idRep, err := franqueadoMeta(in.IDFranqueado)
	if err != nil {
		return out, err
	}

	hoje := dateOnlyBR(time.Now())
	fim := hoje.AddDate(0, 0, 14)
	cicloRef := fmt.Sprintf("EXCECAO-%s-%s-%s", in.IDFranqueado, in.IDCliente, hoje.Format("20060102"))

	_ = cancelarExcecaoPendenteCliente(ctx, in.IDFranqueado, in.IDCliente)

	d, err := db(ctx)
	if err != nil {
		return out, err
	}
	var excID int64
	err = d.QueryRowContext(ctx, `
INSERT INTO ops_parceiro_excecao (
    id_franqueado, id_cliente, nome_cliente, id_parceiro,
    id_parceiro_anterior, id_vinculo_anterior, status,
    preco_unitario, valor_total, ciclo_ref, periodo_inicio, periodo_fim, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
RETURNING id`,
		in.IDFranqueado, in.IDCliente, in.NomeCliente, in.IDParceiro,
		idParceiroAnt, idVinculoAnt, StatusAtivPendente,
		preco.PrecoFranq, valorTotal, cicloRef, hoje, fim,
	).Scan(&excID)
	if err != nil {
		return out, err
	}

	fpID, err := syncFaturaExcecao(ctx, in.IDFranqueado, idCentral, idRep, cicloRef, hoje, preco, valorTotal, excID, in.NomeCliente)
	if err != nil {
		_, _ = d.ExecContext(ctx, `UPDATE ops_parceiro_excecao SET status = $2, updated_at = NOW() WHERE id = $1`, excID, StatusAtivCancelada)
		return out, fmt.Errorf("falha ao gerar fatura: %w", err)
	}
	if fpID > 0 {
		_, _ = d.ExecContext(ctx, `UPDATE ops_parceiro_excecao SET fp_fatura_id = $2, updated_at = NOW() WHERE id = $1`, excID, fpID)
	}

	out = ExcecaoRegistro{
		ID: excID, IDFranqueado: in.IDFranqueado, IDCliente: in.IDCliente,
		NomeCliente: in.NomeCliente, IDParceiro: in.IDParceiro,
		IDParceiroAnterior: idParceiroAnt, IDVinculoAnterior: idVinculoAnt, Status: StatusAtivPendente,
		PrecoUnitario: preco.PrecoFranq, ValorTotal: valorTotal, FPFaturaID: fpID,
		CicloRef: cicloRef, PeriodoInicio: hoje.Format("2006-01-02"), PeriodoFim: fim.Format("2006-01-02"),
	}
	_ = RegistrarHistoricoParceiro(ctx, HistoricoInput{
		IDFranqueado:   in.IDFranqueado,
		IDCliente:      in.IDCliente,
		NomeCliente:    in.NomeCliente,
		Evento:         HistEventoExcecaoProposta,
		IDParceiroDe:   idParceiroAnt,
		IDParceiroPara: in.IDParceiro,
		IDVinculoDe:    idVinculoAnt,
		FPFaturaID:     fpID,
		Detalhe:        map[string]any{"valor_total": valorTotal, "excecao_id": excID},
	})
	return out, nil
}

func ativarVinculoConfService(ctx context.Context, idFranqueado, idCliente, nomeCliente, idParceiro, idVinculoAnterior string) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(config.ConfServiceURL), "/")
	if base == "" {
		return "", errors.New("CONFSERVICE_URL nao configurado")
	}
	payload, _ := json.Marshal(map[string]any{
		"idFranqueado":      idFranqueado,
		"idCliente":         idCliente,
		"nomeCliente":       nomeCliente,
		"idParceiro":        idParceiro,
		"idVinculoAnterior": idVinculoAnterior,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/internal/vinculo/ativar", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if config.ConfServiceAPIKey != "" {
		req.Header.Set("X-Api-Key", config.ConfServiceAPIKey)
	}
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("confservice ativar vinculo HTTP %d: %s", resp.StatusCode, string(body))
	}
	var parsed struct {
		Vinculo *struct {
			ID string `json:"id"`
		} `json:"vinculo"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	if parsed.Vinculo == nil {
		return "", errors.New("resposta confservice sem vinculo")
	}
	return strings.TrimSpace(parsed.Vinculo.ID), nil
}

func AtivarExcecaoPaga(ctx context.Context, idFranqueado, refID string, faturaID int, pagoAte time.Time) error {
	excID, err := strconv.ParseInt(strings.TrimSpace(refID), 10, 64)
	if err != nil || excID <= 0 {
		return fmt.Errorf("ref_id excecao invalido: %s", refID)
	}
	d, err := db(ctx)
	if err != nil {
		return err
	}
	rec, err := scanExcecao(d.QueryRowContext(ctx, excecaoSelectCols+` FROM ops_parceiro_excecao WHERE id = $1`, excID))
	if err != nil {
		return err
	}
	if rec == nil {
		return errors.New("excecao nao encontrada")
	}
	if rec.Status == StatusAtivPaga {
		if strings.TrimSpace(rec.IDVinculoAtivo) != "" {
			if pagoAte.IsZero() {
				pagoAte = parsePagoAteExcecao(rec, time.Now())
			}
			return syncPagoAteVinculoExcecao(ctx, rec, pagoAte)
		}
		// reparo: paga sem vínculo gravado
	} else if rec.Status == StatusAtivCancelada {
		st, err := lookupFaturaStatusXano(ctx, rec.FPFaturaID)
		if err != nil || st != "paga" {
			return fmt.Errorf("excecao cancelada e fatura nao paga")
		}
	} else if rec.Status != StatusAtivPendente {
		return fmt.Errorf("excecao status=%s", rec.Status)
	}

	pagoAte = dateOnlyBR(pagoAte)
	if pagoAte.IsZero() {
		pagoAte = dateOnlyBR(time.Now()).AddDate(0, 0, 14)
	}

	idVinculo, err := ativarVinculoConfService(ctx, rec.IDFranqueado, rec.IDCliente, rec.NomeCliente, rec.IDParceiro, rec.IDVinculoAnterior)
	if err != nil {
		return err
	}
	if vincAnt := strings.TrimSpace(rec.IDVinculoAnterior); vincAnt != "" {
		_ = RemoverVinculoOrdemConsolidada(ctx, rec.IDFranqueado, vincAnt)
	}

	_, err = d.ExecContext(ctx, `
UPDATE ops_parceiro_excecao SET status = $2, fp_fatura_id = COALESCE(fp_fatura_id, $3),
       pago_ate = $4, id_vinculo_ativo = $5, updated_at = NOW()
WHERE id = $1 AND status IN ($6, $7, $8)`,
		excID, StatusAtivPaga, faturaID, pagoAte, idVinculo, StatusAtivPendente, StatusAtivCancelada, StatusAtivPaga)
	if err != nil {
		return err
	}
	_ = syncPagoAteVinculoExcecao(ctx, &ExcecaoRegistro{
		IDFranqueado:   rec.IDFranqueado,
		IDVinculoAtivo: idVinculo,
		PagoAte:        pagoAte.Format("2006-01-02"),
	}, pagoAte)
	_ = RegistrarHistoricoParceiro(ctx, HistoricoInput{
		IDFranqueado:   rec.IDFranqueado,
		IDCliente:      rec.IDCliente,
		NomeCliente:    rec.NomeCliente,
		Evento:         HistEventoExcecaoAtivada,
		IDParceiroDe:   rec.IDParceiroAnterior,
		IDParceiroPara: rec.IDParceiro,
		IDVinculoDe:    rec.IDVinculoAnterior,
		IDVinculoPara:  idVinculo,
		FPFaturaID:     faturaID,
		Detalhe:        map[string]any{"pago_ate": pagoAte.Format("2006-01-02"), "excecao_id": excID},
	})
	return nil
}

// MetaRepasseExcecao retorna vínculo/parceiro/piso para repasse após exceção paga.
func MetaRepasseExcecao(ctx context.Context, refID string, valorPisoHint float64) (idVinculo, idParceiro string, valorPiso float64, err error) {
	excID, err := strconv.ParseInt(strings.TrimSpace(refID), 10, 64)
	if err != nil || excID <= 0 {
		return "", "", 0, fmt.Errorf("ref_id excecao invalido: %s", refID)
	}
	d, err := db(ctx)
	if err != nil {
		return "", "", 0, err
	}
	rec, err := scanExcecao(d.QueryRowContext(ctx, excecaoSelectCols+` FROM ops_parceiro_excecao WHERE id = $1`, excID))
	if err != nil {
		return "", "", 0, err
	}
	if rec == nil {
		return "", "", 0, errors.New("excecao nao encontrada")
	}
	if rec.Status != StatusAtivPaga {
		return "", "", 0, fmt.Errorf("excecao status=%s", rec.Status)
	}
	idVinculo = strings.TrimSpace(rec.IDVinculoAtivo)
	idParceiro = strings.TrimSpace(rec.IDParceiro)
	if idVinculo == "" || idParceiro == "" {
		return "", "", 0, errors.New("excecao sem vinculo ou parceiro")
	}
	valorPiso = valorPisoHint
	if valorPiso <= 0 {
		preco, err := lookupPrecoParceiro(ctx, rec.IDFranqueado, idParceiro)
		if err != nil {
			return "", "", 0, err
		}
		valorPiso = preco.PisoParceiro
	}
	return idVinculo, idParceiro, valorPiso, nil
}

// VinculoAnteriorExcecao retorna id_vinculo_anterior de uma exceção (para ignorar repasse duplicado).
func VinculoAnteriorExcecao(ctx context.Context, refID string) string {
	excID, err := strconv.ParseInt(strings.TrimSpace(refID), 10, 64)
	if err != nil || excID <= 0 {
		return ""
	}
	d, err := db(ctx)
	if err != nil {
		return ""
	}
	var vinc sql.NullString
	_ = d.QueryRowContext(ctx, `SELECT id_vinculo_anterior FROM ops_parceiro_excecao WHERE id = $1`, excID).Scan(&vinc)
	return strings.TrimSpace(vinc.String)
}

func ConsultaVinculoCliente(ctx context.Context, idFranqueado, idCliente string) (map[string]any, error) {
	v, err := lookupVinculoClienteCS(ctx, idFranqueado, idCliente)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return map[string]any{"vinculo": nil}, nil
	}
	return map[string]any{
		"vinculo": map[string]any{
			"id":         v.ID,
			"idParceiro": v.IDParceiro,
		},
	}, nil
}
