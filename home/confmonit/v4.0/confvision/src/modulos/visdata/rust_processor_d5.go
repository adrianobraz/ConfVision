package visdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// D5 — control plane: capacity-report dos rust-processors + assign worker_id.

const d5DefaultProcessorURLs = "https://foxpro-rust-pilot.rkr351.easypanel.host,https://foxpro-rust-pilot-b.rkr351.easypanel.host"

type ProcessorCapacityRow struct {
	ServidorID           string  `json:"servidor_id,omitempty"`
	BaseURL              string  `json:"base_url"`
	Reachable            bool    `json:"reachable"`
	Error                string  `json:"error,omitempty"`
	WorkerID             string  `json:"worker_id,omitempty"`
	ProcessorID          string  `json:"processor_id,omitempty"`
	CapacityState        string  `json:"capacity_state,omitempty"`
	LimitingResource     string  `json:"limiting_resource,omitempty"`
	EstimatedAvailable   int     `json:"estimated_available_cameras"`
	CamerasOnline        int     `json:"cameras_online"`
	CamerasTotal         int     `json:"cameras_total"`
	AllowNewCamera       bool    `json:"allow_new_camera"`
	LoadAdvisory         string  `json:"load_advisory,omitempty"`
	Rtsp404Count         int     `json:"rtsp_404_count"`
	AssignScore          float64 `json:"assign_score"`
	AssignEligible       bool    `json:"assign_eligible"`
}

func D5AutoAssignEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("D5_AUTO_ASSIGN_ENABLED")))
	if v == "0" || v == "false" || v == "off" {
		return false
	}
	return true
}

type rustProcessorRegistryEntry struct {
	ServidorID string
	BaseURL    string
}

// Formato RUST_PROCESSOR_BASE_URLS (vírgula):
//   https://rust-a.example
//   srv-confvision-042|https://rust-a.example
func parseRustProcessorRegistry() []rustProcessorRegistryEntry {
	raw := strings.TrimSpace(os.Getenv("RUST_PROCESSOR_BASE_URLS"))
	if raw == "" {
		raw = d5DefaultProcessorURLs
	}
	parts := strings.Split(raw, ",")
	out := make([]rustProcessorRegistryEntry, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		servidorID := ""
		base := p
		if i := strings.Index(p, "|"); i >= 0 {
			servidorID = strings.TrimSpace(p[:i])
			base = strings.TrimSpace(p[i+1:])
		}
		base = strings.TrimRight(base, "/")
		if base == "" {
			continue
		}
		out = append(out, rustProcessorRegistryEntry{ServidorID: servidorID, BaseURL: base})
	}
	return out
}

func RustProcessorBaseURLs() []string {
	reg := parseRustProcessorRegistry()
	out := make([]string, 0, len(reg))
	for _, e := range reg {
		out = append(out, e.BaseURL)
	}
	return out
}

func filterRegistryByServidor(reg []rustProcessorRegistryEntry, servidorID string) []rustProcessorRegistryEntry {
	servidorID = strings.TrimSpace(servidorID)
	if servidorID == "" {
		return reg
	}
	out := make([]rustProcessorRegistryEntry, 0, len(reg))
	for _, e := range reg {
		if e.ServidorID == "" || strings.EqualFold(e.ServidorID, servidorID) {
			out = append(out, e)
		}
	}
	return out
}

func FetchAllProcessorCapacity(ctx context.Context) ([]ProcessorCapacityRow, error) {
	return FetchProcessorCapacityScoped(ctx, "")
}

func FetchProcessorCapacityScoped(ctx context.Context, servidorID string) ([]ProcessorCapacityRow, error) {
	reg := filterRegistryByServidor(parseRustProcessorRegistry(), servidorID)
	if len(reg) == 0 {
		if strings.TrimSpace(servidorID) != "" {
			return nil, fmt.Errorf("nenhum processor para servidor_id=%s", servidorID)
		}
		return nil, fmt.Errorf("RUST_PROCESSOR_BASE_URLS vazio")
	}
	out := make([]ProcessorCapacityRow, 0, len(reg))
	for _, e := range reg {
		row, _ := fetchOneProcessorCapacity(ctx, e.BaseURL)
		row.ServidorID = e.ServidorID
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].AssignScore > out[j].AssignScore
	})
	return out, nil
}

func fetchOneProcessorCapacity(ctx context.Context, base string) (ProcessorCapacityRow, error) {
	row := ProcessorCapacityRow{BaseURL: base}
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/capacity-report", nil)
	if err != nil {
		row.Error = err.Error()
		return row, err
	}
	resp, err := client.Do(req)
	if err != nil {
		row.Error = err.Error()
		return row, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if resp.StatusCode != http.StatusOK {
		row.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return row, fmt.Errorf("%s", row.Error)
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		row.Error = err.Error()
		return row, err
	}
	row.Reachable = true
	if id, ok := doc["identity"].(map[string]any); ok {
		row.WorkerID = trimAny(id["worker_id"])
		row.ProcessorID = trimAny(id["processor_id"])
	}
	if sum, ok := doc["summary"].(map[string]any); ok {
		row.CamerasOnline = intFromAny(sum["cameras_online"])
		row.CamerasTotal = intFromAny(sum["cameras_total"])
		row.Rtsp404Count = intFromAny(sum["rtsp_404_count"])
	}
	if cap, ok := doc["capacity"].(map[string]any); ok {
		row.CapacityState = trimAny(cap["state"])
		row.LimitingResource = trimAny(cap["limiting_resource"])
		row.EstimatedAvailable = intFromAny(cap["estimated_available_cameras"])
	}
	if load, ok := doc["load"].(map[string]any); ok {
		row.AllowNewCamera = boolFromAny(load["allow_new_camera"])
	}
	if rt, ok := doc["runtime"].(map[string]any); ok {
		row.LoadAdvisory = trimAny(rt["load_advisory"])
	}
	if row.LoadAdvisory == "" {
		row.LoadAdvisory = trimAny(doc["load_advisory"])
	}
	row.AssignScore, row.AssignEligible = scoreProcessorForAssign(row)
	return row, nil
}

func boolFromAny(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true") || t == "1"
	default:
		return false
	}
}

func scoreProcessorForAssign(row ProcessorCapacityRow) (float64, bool) {
	if !row.Reachable || row.WorkerID == "" {
		return -1e9, false
	}
	if !row.AllowNewCamera {
		return -1e6, false
	}
	if strings.Contains(strings.ToLower(row.LoadAdvisory), "reject") {
		return -1e6, false
	}
	score := float64(row.EstimatedAvailable) * 10.0
	switch strings.ToLower(row.CapacityState) {
	case "healthy", "ok", "normal":
		score += 5
	case "warning", "elevated":
		score += 2
	case "critical":
		score -= 5
	}
	score -= float64(row.CamerasOnline) * 0.5
	if row.Rtsp404Count > 0 {
		score -= float64(row.Rtsp404Count) * 2
	}
	eligible := score > 0 && row.EstimatedAvailable > 0
	if row.CapacityState == "critical" && row.EstimatedAvailable <= 0 {
		eligible = false
	}
	return score, eligible
}

func PickBestProcessor(rows []ProcessorCapacityRow) (ProcessorCapacityRow, error) {
	var best *ProcessorCapacityRow
	for i := range rows {
		r := &rows[i]
		if !r.AssignEligible {
			continue
		}
		if best == nil || r.AssignScore > best.AssignScore {
			best = r
		}
	}
	if best == nil {
		return ProcessorCapacityRow{}, fmt.Errorf("nenhum processor elegivel para assign (capacity/admission)")
	}
	return *best, nil
}

func AssignCameraWorkerID(ctx context.Context, cameraID int, workerID string) (map[string]any, error) {
	if cameraID <= 0 {
		return nil, fmt.Errorf("camera_id invalido")
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return nil, fmt.Errorf("worker_id obrigatorio")
	}
	return UpdateCamera(ctx, cameraID, map[string]any{"worker_id": workerID})
}

func AutoAssignCameraWorkerScoped(ctx context.Context, cameraID int, servidorID string) (string, map[string]any, error) {
	cam, err := GetCameraByID(ctx, cameraID)
	if err != nil {
		return "", nil, err
	}
	if !cameraNeedsWorkerAssign(cam) {
		w := trimAny(cam["worker_id"])
		return w, map[string]any{"skipped": true, "reason": "worker_id_ja_definido", "worker_id": w}, nil
	}
	rows, err := FetchProcessorCapacityScoped(ctx, servidorID)
	if err != nil {
		return "", nil, err
	}
	best, err := PickBestProcessor(rows)
	if err != nil {
		return "", nil, err
	}
	if _, err := AssignCameraWorkerID(ctx, cameraID, best.WorkerID); err != nil {
		return "", nil, err
	}
	meta := map[string]any{
		"camera_id":    cameraID,
		"worker_id":    best.WorkerID,
		"processor_id": best.ProcessorID,
		"base_url":     best.BaseURL,
		"servidor_id":  best.ServidorID,
		"assign_score": best.AssignScore,
		"reason":       "d5_auto_capacity",
	}
	if strings.TrimSpace(servidorID) != "" {
		meta["assign_scope"] = servidorID
	}
	return best.WorkerID, meta, nil
}

func AutoAssignCameraWorker(ctx context.Context, cameraID int) (string, map[string]any, error) {
	return AutoAssignCameraWorkerScoped(ctx, cameraID, "")
}

func cameraNeedsWorkerAssign(cam map[string]any) bool {
	if cam == nil {
		return false
	}
	if w := strings.TrimSpace(trimAny(cam["worker_id"])); w != "" {
		return false
	}
	if !boolFromAny(cam["ativo"]) {
		return false
	}
	if boolFromAny(cam["analitico_pausado"]) {
		return false
	}
	return boolFromAny(cam["deteccao_humano"])
}

func ListUnassignedAnaliticas(ctx context.Context, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 200
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT c.id, c.nome, c.id_franqueado, c.worker_id, c.deteccao_humano, c.analitico_pausado
FROM vis_camera c
WHERE c.ativo = TRUE
  AND c.deteccao_humano = TRUE
  AND (c.analitico_pausado IS NOT TRUE)
  AND NULLIF(BTRIM(c.worker_id), '') IS NULL
ORDER BY c.id ASC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int
		var nome, idFra, worker sql.NullString
		var det, pausa bool
		if err := rows.Scan(&id, &nome, &idFra, &worker, &det, &pausa); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id": id, "nome": nullStr(nome), "id_franqueado": nullStr(idFra),
			"deteccao_humano": det, "analitico_pausado": pausa,
		})
	}
	return out, rows.Err()
}

func AutoAssignUnassignedAnaliticas(ctx context.Context, dryRun bool, limit int) (map[string]any, error) {
	list, err := ListUnassignedAnaliticas(ctx, limit)
	if err != nil {
		return nil, err
	}
	assigned := []map[string]any{}
	skipped := []map[string]any{}
	errors := []map[string]any{}
	for _, cam := range list {
		cid := intFromAny(cam["id"])
		if dryRun {
			rows, _ := FetchAllProcessorCapacity(ctx)
			best, pickErr := PickBestProcessor(rows)
			if pickErr != nil {
				errors = append(errors, map[string]any{"camera_id": cid, "erro": pickErr.Error()})
				continue
			}
			assigned = append(assigned, map[string]any{
				"camera_id": cid, "worker_id": best.WorkerID, "dry_run": true,
			})
			continue
		}
		wid, meta, aerr := AutoAssignCameraWorker(ctx, cid)
		if aerr != nil {
			errors = append(errors, map[string]any{"camera_id": cid, "erro": aerr.Error()})
			continue
		}
		if skip, _ := meta["skipped"].(bool); skip {
			skipped = append(skipped, meta)
			continue
		}
		meta["camera_id"] = cid
		meta["worker_id"] = wid
		assigned = append(assigned, meta)
	}
	return map[string]any{
		"ok": true, "dry_run": dryRun,
		"total_unassigned": len(list),
		"assigned":         assigned,
		"skipped":          skipped,
		"errors":           errors,
	}, nil
}

func MaybeAutoAssignAfterCreate(ctx context.Context, cam map[string]any) map[string]any {
	if !D5AutoAssignEnabled() || cam == nil {
		return nil
	}
	if !cameraNeedsWorkerAssign(cam) {
		return nil
	}
	cid := intFromAny(cam["id"])
	if cid <= 0 {
		return nil
	}
	_, meta, err := AutoAssignCameraWorker(ctx, cid)
	if err != nil {
		return map[string]any{"camera_id": cid, "erro": err.Error()}
	}
	return meta
}
