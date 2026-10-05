package visdata

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"confvision/src/config"
)

var (
	integracaoDispatchMu       sync.Mutex
	integracaoDispatchInFlight = map[int]bool{}
)

func scheduleIntegracaoDispatch(eventoID int) {
	if eventoID <= 0 {
		return
	}

	integracaoDispatchMu.Lock()
	if integracaoDispatchInFlight[eventoID] {
		integracaoDispatchMu.Unlock()
		return
	}
	integracaoDispatchInFlight[eventoID] = true
	integracaoDispatchMu.Unlock()

	go func() {
		defer func() {
			integracaoDispatchMu.Lock()
			delete(integracaoDispatchInFlight, eventoID)
			integracaoDispatchMu.Unlock()
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := dispatchIntegracoesEvento(ctx, eventoID); err != nil {
			log.Printf("[INTEGRACAO] evento=%d falha: %v", eventoID, err)
		}
	}()
}

func maybeScheduleIntegracaoDispatch(ctx context.Context, eventoID int, snapshotURL string) {
	if strings.TrimSpace(snapshotURL) == "" {
		return
	}
	scheduleIntegracaoDispatch(eventoID)
	scheduleComunitiaWebhook(eventoID)
}

// integracaoSistemaPermiteDispatch indica se ha integracao ativa (nao Nenhum) configurada.
func integracaoSistemaPermiteDispatch(found bool, sistema string, ativo bool) bool {
	if !found {
		return false
	}
	s := normalizeIntegracaoSistema(sistema)
	if s == "nenhum" || !ativo {
		return false
	}
	return true
}

func dispatchIntegracoesEvento(ctx context.Context, eventoID int) error {
	if !config.IntegracaoDispatchEnabled {
		return nil
	}

	row, err := loadEventoIntegracaoRow(ctx, eventoID)
	if err != nil {
		return err
	}
	if row == nil || row.IDFranqueado == "" {
		return nil
	}

	cfg, found, err := loadIntegracaoFranqueadoConfig(ctx, row.IDFranqueado)
	if err != nil {
		return err
	}

	sistema := ""
	ativo := false
	if cfg != nil {
		sistema = cfg.Sistema
		ativo = cfg.Ativo
	}
	// Sem registro ou integracao Nenhum/inativa: nao envia a terminal, Moni, etc.
	if !integracaoSistemaPermiteDispatch(found, sistema, ativo) {
		return nil
	}

	var primaryErr error
	switch cfg.Sistema {
	case "confmonit":
		return dispatchConfmonitIntegracao(ctx, eventoID, *cfg)
	case "moni":
		primaryErr = dispatchMoniIntegracao(ctx, eventoID, *cfg)
	case "dguard", "segware":
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, cfg.Sistema, false, 0,
			"integracao em desenvolvimento", "")
	default:
		return nil
	}

	if cfg.DuplaComunicacao && integracaoDuplaComunicacaoHabilitada(cfg.Sistema) {
		if err := dispatchConfmonitIntegracao(ctx, eventoID, *cfg); err != nil {
			log.Printf("[INTEGRACAO] evento=%d dupla confmonit falha: %v", eventoID, err)
			if primaryErr == nil {
				primaryErr = err
			}
		}
	}
	return primaryErr
}
