package visdata

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"confvision/src/config"
)

var (
	comunitiaDispatchMu       sync.Mutex
	comunitiaDispatchInFlight = map[int]bool{}
)

func scheduleComunitiaWebhook(eventoID int) {
	if eventoID <= 0 || !config.ComunitiaWebhookEnabled() {
		return
	}
	comunitiaDispatchMu.Lock()
	if comunitiaDispatchInFlight[eventoID] {
		comunitiaDispatchMu.Unlock()
		return
	}
	comunitiaDispatchInFlight[eventoID] = true
	comunitiaDispatchMu.Unlock()

	go func() {
		defer func() {
			comunitiaDispatchMu.Lock()
			delete(comunitiaDispatchInFlight, eventoID)
			comunitiaDispatchMu.Unlock()
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := dispatchComunitiaWebhook(ctx, eventoID); err != nil {
			log.Printf("[COMUNITIA] evento=%d falha: %v", eventoID, err)
		}
	}()
}

func dispatchComunitiaWebhook(ctx context.Context, eventoID int) error {
	row, err := loadEventoComunitiaRow(ctx, eventoID)
	if err != nil {
		return err
	}
	if row == nil {
		return nil
	}
	if strings.TrimSpace(row.SnapshotURL) == "" {
		return nil
	}

	idCamera := row.IDCamera
	if idCamera == "" {
		idCamera = row.IDDispositivo
	}
	if idCamera == "" {
		return fmt.Errorf("evento sem id_camera")
	}

	texto := strings.TrimSpace(row.TipoDeteccao)
	if texto == "" {
		texto = "Evento"
	}
	if n := strings.TrimSpace(row.CameraNome); n != "" {
		texto = texto + " · " + n
	}

	payload := map[string]string{
		"id_evento_ext": strconv.Itoa(eventoID),
		"id_camera":     idCamera,
		"id_cliente":    row.IDCliente,
		"texto":         texto,
		"midia_url":     row.SnapshotURL,
		"tipo_evento":   row.TipoDeteccao,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.ComunitiaWebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", config.ComunitiaWebhookSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("comunitia webhook %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	log.Printf("[COMUNITIA] evento=%d enviado camera=%s cliente=%s", eventoID, idCamera, row.IDCliente)
	return nil
}

type eventoComunitiaRow struct {
	ID            int
	IDCliente     string
	IDDispositivo string
	IDCamera      string
	TipoDeteccao  string
	SnapshotURL   string
	CameraNome    string
}

func loadEventoComunitiaRow(ctx context.Context, eventoID int) (*eventoComunitiaRow, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var row eventoComunitiaRow
	var idCli, idDisp, tipo, snap sql.NullString
	var visCamID sql.NullInt64
	var camNome sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT e.id, e.id_cliente, e.id_dispositivo, e.vis_camera_id, e.tipo_deteccao, e.snapshot_url,
       COALESCE(NULLIF(TRIM(c.nome), ''), '') AS camera_nome
FROM vis_evento e
LEFT JOIN vis_camera c ON c.id = e.vis_camera_id
WHERE e.id = $1`, eventoID).Scan(
		&row.ID, &idCli, &idDisp, &visCamID, &tipo, &snap, &camNome,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	row.IDCliente = sqlString(idCli)
	row.IDDispositivo = sqlString(idDisp)
	if visCamID.Valid && visCamID.Int64 > 0 {
		row.IDCamera = strconv.FormatInt(visCamID.Int64, 10)
	}
	row.TipoDeteccao = sqlString(tipo)
	row.SnapshotURL = sqlString(snap)
	row.CameraNome = sqlString(camNome)
	return &row, nil
}
