package autofim

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	Interval      time.Duration
	Concurrency   int
	PendingURL    string
	ClaimURL      string
	ContextURL    string
	ProcessEndURL string
	LogURL        string
	SuccessURL    string
	FailURL       string
}

var running int32

func StartLoop(client *http.Client, cfg Config, shutdown <-chan struct{}) {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	runOnce(client, cfg)
	for {
		select {
		case <-shutdown:
			log.Printf("autofim: loop encerrado")
			return
		case <-ticker.C:
			runOnce(client, cfg)
		}
	}
}

func runOnce(client *http.Client, cfg Config) {
	if !atomic.CompareAndSwapInt32(&running, 0, 1) {
		log.Printf("autofim: ciclo anterior ainda rodando, pulando")
		return
	}
	defer atomic.StoreInt32(&running, 0)

	if err := process(client, cfg); err != nil {
		log.Printf("autofim: erro no ciclo: %v", err)
	}
}

func process(client *http.Client, cfg Config) error {
	limit := cfg.Concurrency * 20
	if limit < 20 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := getPending(client, cfg.PendingURL, limit)
	if err != nil {
		return fmt.Errorf("pending: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for _, row := range rows {
		row := row
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			processRow(client, cfg, row)
		}()
	}
	wg.Wait()
	return nil
}

func processRow(client *http.Client, cfg Config, row pendingRow) {
	lockToken := fmt.Sprintf("go-autofim-%d-%d", row.ID, time.Now().UnixNano())

	claimResp, err := claim(client, cfg.ClaimURL, row.ID, lockToken)
	if err != nil {
		log.Printf("autofim: claim erro id=%d: %v", row.ID, err)
		_ = markFail(client, cfg.FailURL, row.ID, lockToken, "claim erro: "+err.Error(), 15)
		return
	}
	if !claimResp.Dados.Claimed {
		return
	}

	contextResp, err := getContext(client, cfg.ContextURL, claimResp.Dados.Row.AlarmEventsID)
	if err != nil {
		log.Printf("autofim: context erro id=%d: %v", row.ID, err)
		_ = markFail(client, cfg.FailURL, row.ID, lockToken, "context erro: "+err.Error(), 30)
		return
	}

	dec := evaluateBotFinalizaEventoAuto(contextResp.Dados.EventosProcesso, contextResp.Dados.HistAgg, contextResp.Dados.ProcessoJaFinalizado)

	var retornoProcessoEnd map[string]any
	if dec.FinalizarAgora {
		procEnd, procErr := callProcessEnd(client, cfg.ProcessEndURL, contextResp.Dados.Evento.IDProcesso, dec.Motivo, dec.PerfilLocal)
		if procErr != nil {
			log.Printf("autofim: processo-end erro id=%d: %v", row.ID, procErr)
			_ = markFail(client, cfg.FailURL, row.ID, lockToken, "processo-end erro: "+procErr.Error(), 60)
			return
		}
		dec.Acao = procEnd.Dados.Acao
		retornoProcessoEnd = procEnd.Dados.RetornoProcessoEnd
	}

	if dec.Acao == "FINALIZOU" {
		logErr := sendLog(client, cfg.LogURL, map[string]any{
			"idEvento":            claimResp.Dados.Row.AlarmEventsID,
			"idProcesso":          contextResp.Dados.Evento.IDProcesso,
			"idDispositivo":       contextResp.Dados.Evento.IDDispositivo,
			"motivo":              dec.Motivo,
			"acao":                dec.Acao,
			"qtdCiclos3mDiffProc": dec.QtdCiclos3mDiffProc,
			"temFalhas":           dec.TemFalhas,
			"temAlarme":           dec.TemAlarme,
			"temDesarme":          dec.TemDesarme,
			"temRestaure":         dec.TemRestaure,
			"temParAlarmeRest50":  dec.TemParAlarmeRest50,
			"retornoProcessoEnd":  retornoProcessoEnd,
			"regraVersao":         "vGo_autofim_worker",
			"erro":                "",
		})
		if logErr != nil {
			log.Printf("autofim: log erro id=%d: %v", row.ID, logErr)
			_ = markFail(client, cfg.FailURL, row.ID, lockToken, "log erro: "+logErr.Error(), 60)
			return
		}
	}

	payloadResultado := map[string]any{
		"idEvento":             claimResp.Dados.Row.AlarmEventsID,
		"idProcesso":           contextResp.Dados.Evento.IDProcesso,
		"idDispositivo":        contextResp.Dados.Evento.IDDispositivo,
		"perfilLocal":          dec.PerfilLocal,
		"motivo":               dec.Motivo,
		"acao":                 dec.Acao,
		"temAlarme":            dec.TemAlarme,
		"temRestaure":          dec.TemRestaure,
		"temDesarme":           dec.TemDesarme,
		"temArme":              dec.TemArme,
		"temFalhas":            dec.TemFalhas,
		"temParAlarmeRest50":   dec.TemParAlarmeRest50,
		"qtdCiclos3mDiffProc":  dec.QtdCiclos3mDiffProc,
		"retornoProcessoEnd":   retornoProcessoEnd,
		"processoJaFinalizado": contextResp.Dados.ProcessoJaFinalizado,
		"dataAtenFim":          contextResp.Dados.DataAtenFim,
	}

	if err := markSuccess(client, cfg.SuccessURL, row.ID, lockToken, dec.Motivo, dec.Acao, payloadResultado); err != nil {
		log.Printf("autofim: success erro id=%d: %v", row.ID, err)
		_ = markFail(client, cfg.FailURL, row.ID, lockToken, "success erro: "+err.Error(), 60)
		return
	}
}
