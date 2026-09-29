package visdata

import (
	"context"
	"strings"
)

// maybeScheduleIntegracaoDispatch dispara integrações externas após evento com snapshot.
// Implementação completa (Moni/ConfMonit) fica em integracao_*.go quando habilitada.
func maybeScheduleIntegracaoDispatch(ctx context.Context, eventoID int, snapshotURL string) {
	_ = ctx
	_ = eventoID
	_ = snapshotURL
}

// integracaoSistemaPermiteDispatch regra mínima compartilhada com testes e futuro dispatch Moni.
func integracaoSistemaPermiteDispatch(found bool, sistema string, ativo bool) bool {
	if !found || !ativo {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(sistema))
	if s == "" || s == "nenhum" {
		return false
	}
	return true
}
