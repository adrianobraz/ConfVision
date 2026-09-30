package pgcobranca

import (
	"context"
	"fmt"
	"time"

	"apifunction/config"
	"apifunction/xano"
)

type Engine struct {
	xano *xano.Client
}

func NovoEngine() *Engine {
	return &Engine{xano: xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)}
}

func (e *Engine) Simular(ctx context.Context, idFranqueado string, anchor time.Time) (*OrdemSimulada, error) {
	cfg, err := GetConfig(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	servicos, err := ListServicos(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	return MontarOrdem(idFranqueado, cfg, anchor, servicos), nil
}

func (e *Engine) GerarOrdens(ctx context.Context, anchor time.Time, idFranqueadoFiltro string) (GerarOrdensResult, error) {
	res := GerarOrdensResult{Vencimento: dateOnly(anchor).Format("2006-01-02")}
	var franqueados []string
	var err error
	if idFranqueadoFiltro != "" {
		franqueados = []string{idFranqueadoFiltro}
	} else {
		franqueados, err = ListFranqueadosConsolidados(ctx)
		if err != nil {
			return res, err
		}
	}

	for _, idFra := range franqueados {
		res.Processados++
		cfg, err := GetConfig(ctx, idFra)
		if err != nil {
			res.Erros = append(res.Erros, err.Error())
			continue
		}
		if cfg.Modo == ModoOriginal {
			continue
		}
		if !anchorMatches(cfg, anchor) {
			continue
		}
		ord, err := e.Simular(ctx, idFra, anchor)
		if err != nil {
			res.Erros = append(res.Erros, err.Error())
			continue
		}
		if ord == nil || ord.ValorTotal <= 0 {
			continue
		}
		ja, err := OrdemJaGerada(ctx, idFra, ord.CicloRef)
		if err != nil {
			res.Erros = append(res.Erros, err.Error())
			continue
		}
		if ja {
			continue
		}

		fpID := 0
		if e.xano != nil && e.xano.Enabled() {
			fpID, err = e.syncXano(cfg, ord)
			if err != nil {
				res.Erros = append(res.Erros, fmt.Sprintf("%s: %v", idFra, err))
				continue
			}
		}

		_ = LogOrdem(ctx, idFra, ord.CicloRef, ord.Vencimento, ord.ValorTotal, ord.TotalAjustes, len(ord.Linhas), fpID)
		var ids []int64
		for _, l := range ord.Linhas {
			ids = append(ids, l.ServicoID)
		}
		_ = MarcarServicosNormal(ctx, ids)
		_ = AvancarPagoAteLinhas(ctx, idFra, ord.Linhas)
		res.Ordens++
		if fpID > 0 {
			res.FaturasXano = append(res.FaturasXano, fpID)
		}
	}
	return res, nil
}

func (e *Engine) syncXano(cfg Config, ord *OrdemSimulada) (int, error) {
	itens := make([]xano.OrdemItemSync, 0, len(ord.Linhas))
	for _, l := range ord.Linhas {
		itens = append(itens, xano.OrdemItemSync{
			Descricao:     l.Descricao,
			Quantidade:    1,
			ValorUnitario: l.ValorLiquido,
			ValorTotal:    l.ValorLiquido,
			RefTipo:       l.RefTipo,
			RefID:         l.RefID,
			ValorPiso:     l.ValorPiso,
			MargemCentral: l.MargemCentral,
			MargemRep:     l.MargemRep,
		})
	}
	return e.xano.SyncFaturaOrdem(xano.SyncOrdemInput{
		IDFranqueado:    ord.IDFranqueado,
		IDCentral:       cfg.IDCentral,
		IDRepresentante: cfg.IDRepresentante,
		CicloRef:        ord.CicloRef,
		VencimentoEm:    ord.Vencimento.Format("2006-01-02"),
		ValorTotal:      ord.ValorTotal,
		Itens:           itens,
	})
}
