package pgcobranca

import (
	"context"
	"time"
)

type PosPagamentoInput struct {
	IDFranqueado string
	FaturaID     int
	PagoEm       time.Time
	Itens        []PosPagamentoItem
}

type PosPagamentoItem struct {
	RefTipo   string
	RefID     string
	ValorPiso float64
	PagoAte   time.Time
}

type PosPagamentoResult struct {
	PagoAteAtualizados int `json:"pago_ate_atualizados"`
	ItensParceiro      int `json:"itens_parceiro"`
}

func PosPagamento(ctx context.Context, in PosPagamentoInput) (PosPagamentoResult, error) {
	res := PosPagamentoResult{}
	pagoEm := in.PagoEm
	if pagoEm.IsZero() {
		pagoEm = time.Now()
	}
	for _, it := range in.Itens {
		if it.RefTipo == "" || it.RefID == "" {
			continue
		}
		ate := it.PagoAte
		if ate.IsZero() {
			ate = pagoEm
		}
		if err := AtualizarPagoAtePorRef(ctx, in.IDFranqueado, it.RefTipo, it.RefID, ate); err == nil {
			res.PagoAteAtualizados++
		}
		if it.RefTipo == "cs_parceiro_vinculo" && it.ValorPiso > 0 {
			res.ItensParceiro++
		}
	}
	return res, nil
}

func AvancarPagoAteLinhas(ctx context.Context, idFranqueado string, linhas []LinhaOrdem) error {
	for _, l := range linhas {
		if l.RefID == "" || l.RefTipo == "" {
			continue
		}
		fim := dateOnly(l.PeriodoFim)
		_ = AtualizarPagoAtePorRef(ctx, idFranqueado, l.RefTipo, l.RefID, fim)
	}
	return nil
}
