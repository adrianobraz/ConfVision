package pgatendimento

import (
	"context"
	"strings"

	"apifunction/config"
	"apifunction/pgcobranca"
	"apifunction/xano"
)

// RemoverVinculoOrdemConsolidada tira linha cs_parceiro_vinculo de faturas abertas (ordem consolidada).
func RemoverVinculoOrdemConsolidada(ctx context.Context, idFranqueado, idVinculo string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idVinculo = strings.TrimSpace(idVinculo)
	if idFranqueado == "" || idVinculo == "" {
		return nil
	}
	_ = pgcobranca.DesativarServicoPorRef(ctx, idFranqueado, "cs_parceiro_vinculo", idVinculo)
	cli := xano.New(config.XanoAPIFinanceiro, config.WorkerSecret)
	if cli == nil || !cli.Enabled() {
		return nil
	}
	_, err := cli.RemoverItensOrdemConsolidada(idFranqueado, []xano.RemoverOrdemItem{{
		RefTipo: "cs_parceiro_vinculo",
		RefID:   idVinculo,
	}})
	return err
}
