package visdata

import (
	"bytes"
	"confvision/src/config"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	dispatchMu      sync.Mutex
	dispatchInFlight = map[int]bool{}
)

type eventoDispatchRow struct {
	ID            int
	IDFranqueado  string
	Conta         string
	Particao      string
	ZonaUser      string
	TipoDeteccao  string
	SnapshotURL   string
	IDEvento      string
}

func scheduleTerminalDispatch(eventoID int) {
	if eventoID <= 0 || !config.TerminalNotifyEnabled {
		return
	}
	if config.ReceptorWebURL == "" || config.ReceptorWebSenha == "" {
		return
	}

	dispatchMu.Lock()
	if dispatchInFlight[eventoID] {
		dispatchMu.Unlock()
		return
	}
	dispatchInFlight[eventoID] = true
	dispatchMu.Unlock()

	go func() {
		defer func() {
			dispatchMu.Lock()
			delete(dispatchInFlight, eventoID)
			dispatchMu.Unlock()
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := dispatchTerminalEvento(ctx, eventoID); err != nil {
			log.Printf("[TERMINAL] evento=%d falha: %v", eventoID, err)
		}
	}()
}

func maybeScheduleTerminalDispatch(ctx context.Context, eventoID int, snapshotURL string) {
	if strings.TrimSpace(snapshotURL) == "" {
		return
	}
	scheduleTerminalDispatch(eventoID)
}

func dispatchTerminalEvento(ctx context.Context, eventoID int) error {
	row, err := loadEventoDispatchRow(ctx, eventoID)
	if err != nil {
		return err
	}
	if row == nil {
		return nil
	}
	if strings.EqualFold(row.TipoDeteccao, "sensor") {
		return nil
	}
	if strings.TrimSpace(row.IDEvento) != "" {
		return nil
	}
	if strings.TrimSpace(row.SnapshotURL) == "" {
		return nil
	}

	temTerminal, err := FranqueadoTemTerminal(ctx, row.IDFranqueado)
	if err != nil {
		return fmt.Errorf("checagem terminal: %w", err)
	}
	if !temTerminal {
		log.Printf("[TERMINAL] evento=%d ignorado: franqueado=%s sem terminal", eventoID, row.IDFranqueado)
		return nil
	}

	conta := padDigits(row.Conta, 4)
	particao := padDigits(row.Particao, 2)
	zona := normalizeZonaUser(row.ZonaUser)
	if row.IDFranqueado == "" || conta == "" || conta == "0000" {
		log.Printf("[TERMINAL] evento=%d ignorado: franqueado/conta ausentes", eventoID)
		return nil
	}

	payload := map[string]any{
		"idFranqueado":  row.IDFranqueado,
		"conta":         conta,
		"particao":      particao,
		"zonauser":      zona,
		"vis_evento_id": eventoID,
		"senha":         config.ReceptorWebSenha,
		"codigo":        config.TerminalContactID,
	}

	idEvento, idProcesso, err := notifyReceptorWithRetry(ctx, payload)
	if err != nil {
		return err
	}
	if idEvento == "" && idProcesso == "" {
		return nil
	}

	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
UPDATE vis_evento
SET id_evento = COALESCE(NULLIF($2, ''), id_evento),
    id_processo = COALESCE(NULLIF($3, ''), id_processo)
WHERE id = $1 AND (id_evento IS NULL OR id_evento = '')`,
		eventoID, idEvento, idProcesso)
	if err != nil {
		return fmt.Errorf("atualizar ids terminal: %w", err)
	}

	log.Printf("[TERMINAL] ok evento=%d idEvento=%s processo=%s conta=%s part=%s zona=%s",
		eventoID, idEvento, idProcesso, conta, particao, zona)
	return nil
}

func loadEventoDispatchRow(ctx context.Context, eventoID int) (*eventoDispatchRow, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var row eventoDispatchRow
	var idFra, conta, part, tipo, snap, idEvt sql.NullString
	var zona sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT e.id, e.id_franqueado, e.conta, e.particao, e.tipo_deteccao, e.snapshot_url, e.id_evento,
       COALESCE(NULLIF(TRIM(c.zonauser), ''), NULLIF(TRIM(e.canal), ''), '001') AS zonauser
FROM vis_evento e
LEFT JOIN vis_camera c ON c.id = e.vis_camera_id
WHERE e.id = $1`, eventoID).Scan(
		&row.ID, &idFra, &conta, &part, &tipo, &snap, &idEvt, &zona,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	row.IDFranqueado = sqlString(idFra)
	row.Conta = sqlString(conta)
	row.Particao = sqlString(part)
	row.TipoDeteccao = sqlString(tipo)
	row.SnapshotURL = sqlString(snap)
	row.IDEvento = sqlString(idEvt)
	row.ZonaUser = sqlString(zona)
	return &row, nil
}

func notifyReceptorWithRetry(ctx context.Context, payload map[string]any) (string, string, error) {
	url := config.ReceptorWebURL + "/recebe-evento-confvision"
	retries := config.TerminalNotifyRetries
	if retries < 1 {
		retries = 1
	}

	var lastErr error
	for attempt := 1; attempt <= retries; attempt++ {
		idEvento, idProcesso, err := postReceptorConfVision(ctx, url, payload)
		if err == nil {
			return idEvento, idProcesso, nil
		}
		lastErr = err
		log.Printf("[TERMINAL] evento=%v tentativa=%d/%d: %v", payload["vis_evento_id"], attempt, retries, err)
		if attempt < retries {
			sleep := time.Duration(min(attempt*attempt, 10)) * time.Second
			select {
			case <-ctx.Done():
				return "", "", ctx.Err()
			case <-time.After(sleep):
			}
		}
	}
	return "", "", lastErr
}

func postReceptorConfVision(ctx context.Context, url string, payload map[string]any) (string, string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	if res.StatusCode >= 400 {
		msg := strings.TrimSpace(string(raw))
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return "", "", fmt.Errorf("HTTP %d: %s", res.StatusCode, msg)
	}

	var data map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &data); err != nil {
			return "", "", err
		}
	}
	dados, _ := data["dados"].(map[string]any)
	if dados == nil {
		dados = data
	}
	idEvento := strings.TrimSpace(fmt.Sprint(dados["idEvento"]))
	idProcesso := strings.TrimSpace(fmt.Sprint(dados["idProcesso"]))
	if idEvento == "<nil>" {
		idEvento = ""
	}
	if idProcesso == "<nil>" {
		idProcesso = ""
	}
	return idEvento, idProcesso, nil
}

func sqlString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func padDigits(value string, size int) string {
	digits := strings.Builder{}
	for _, r := range strings.TrimSpace(value) {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	s := digits.String()
	if s == "" {
		return strings.Repeat("0", size)
	}
	if len(s) >= size {
		return s[len(s)-size:]
	}
	return strings.Repeat("0", size-len(s)) + s
}

func normalizeZonaUser(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "001"
	}
	allDigits := true
	for _, r := range raw {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return padDigits(raw, 3)
	}
	return raw
}
