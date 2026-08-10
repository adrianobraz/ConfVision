package visapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const cameraSelectCols = `
c.id, c.created_at, c.ativo, c.bloqueado, c.nome, c.id_franqueado, c.id_cliente,
c.id_dispositivo, c.conta, c.particao, c.canal, c.setor, c.protocolo, c.rtsp_url_sec,
c.onvif_host, c.onvif_porta, c.onvif_usuario, c.onvif_senha, c.confianca_min, c.cooldown_seg,
c.somente_armado, c.deteccao_humano, c.deteccao_veiculo, c.status, c.ultimo_evento_em,
c.worker_id, c.ultimo_ping_em, c.zonauser, c.captura_sensor, c.captura_analitico,
c.analitico_pausado, c.id_setor, c.snapshot_url, c.vis_licenca_id, c.plano, c.ativado_em,
c.evento_grava_foto, c.evento_grava_video, c.vis_licenca_gravacao_id, c.grava_continua,
c.retencao_dias, c.gravacao_ativada_em, c.gravacao_status, c.grava_movimento,
c.grava_timelapse, c.gravacao_flush_pedido, c.modo_deteccao, c.vis_mediamtx_node_id,
COALESCE(n.rtsp_internal, '') AS mediamtx_rtsp_base
`

func (r *Repository) ListCamerasAnaliticas(workerID string, nodeID int) ([]map[string]any, error) {
	query := `
SELECT ` + cameraSelectCols + `
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE c.ativo = TRUE
  AND c.deteccao_humano = TRUE
  AND (c.analitico_pausado IS NOT TRUE)
`
	args := []any{}
	n := 1
	if workerID != "" {
		query += fmt.Sprintf(" AND c.worker_id = $%d", n)
		args = append(args, workerID)
		n++
	}
	if nodeID > 0 {
		query += fmt.Sprintf(" AND c.vis_mediamtx_node_id = $%d", n)
		args = append(args, nodeID)
	}
	query += " ORDER BY c.id ASC"

	return r.queryCameras(query, args...)
}

func (r *Repository) ListGravacaoCameras(workerID string, nodeID int) ([]map[string]any, error) {
	query := `
SELECT ` + cameraSelectCols + `
FROM vis_camera c
LEFT JOIN vis_mediamtx_node n ON n.id = c.vis_mediamtx_node_id
WHERE (c.grava_continua = TRUE OR c.grava_movimento = TRUE OR c.grava_timelapse = TRUE)
  AND (c.bloqueado IS NOT TRUE)
`
	args := []any{}
	n := 1
	if workerID != "" {
		query += fmt.Sprintf(" AND c.worker_id = $%d", n)
		args = append(args, workerID)
		n++
	}
	if nodeID > 0 {
		query += fmt.Sprintf(" AND c.vis_mediamtx_node_id = $%d", n)
		args = append(args, nodeID)
	}
	query += " ORDER BY c.id ASC"

	return r.queryCameras(query, args...)
}

func (r *Repository) ListAreasAtivas() ([]map[string]any, error) {
	rows, err := r.db.Query(`
SELECT id, created_at, vis_camera_id, nome, ativo, poligono_json, cor
FROM vis_camera_area
WHERE ativo = TRUE
ORDER BY vis_camera_id, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		m, err := scanAreaRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Repository) GetCameraRTMPAuth(cameraID int) (map[string]any, error) {
	row := r.db.QueryRow(`
SELECT id, id_franqueado, ativo, bloqueado, protocolo, plano
FROM vis_camera WHERE id = $1`, cameraID)

	var id int
	var idFranqueado, protocolo, plano sql.NullString
	var ativo, bloqueado sql.NullBool

	if err := row.Scan(&id, &idFranqueado, &ativo, &bloqueado, &protocolo, &plano); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":            id,
		"id_franqueado": nullStr(idFranqueado),
		"ativo":         nullBool(ativo),
		"bloqueado":     nullBool(bloqueado),
		"protocolo":     nullStr(protocolo),
		"plano":         nullStr(plano),
	}, nil
}

type WorkerPingInput struct {
	WorkerID          string   `json:"worker_id"`
	WorkerTipo        string   `json:"worker_tipo"`
	Hostname          string   `json:"hostname"`
	Versao            string   `json:"versao"`
	CamerasAtivas     int      `json:"cameras_ativas"`
	UltimoPingEm      string   `json:"ultimo_ping_em"`
	Ativo             *bool    `json:"ativo"`
	ShardIndex        *int     `json:"shard_index"`
	ShardTotal        *int     `json:"shard_total"`
	MaxCameras        *int     `json:"max_cameras"`
	YoloDevice        string   `json:"yolo_device"`
	QueueBackend      string   `json:"queue_backend"`
	VisMediamtxNodeID *int     `json:"vis_mediamtx_node_id"`
	CPUPercent        *float64 `json:"cpu_percent"`
	MemPercent        *float64 `json:"mem_percent"`
	Load1m            *float64 `json:"load_1m"`
}

func (r *Repository) UpsertWorkerPing(in WorkerPingInput) (map[string]any, error) {
	tipo := strings.TrimSpace(in.WorkerTipo)
	if tipo == "" {
		tipo = "analitico"
	}

	pingAt := time.Now().UTC()
	if t := strings.TrimSpace(in.UltimoPingEm); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			pingAt = parsed
		}
	}

	ativo := true
	if in.Ativo != nil {
		ativo = *in.Ativo
	}

	nodeID := 0
	if in.VisMediamtxNodeID != nil && *in.VisMediamtxNodeID > 0 {
		nodeID = *in.VisMediamtxNodeID
	}

	var shardIndex, shardTotal, maxCameras sql.NullInt64
	if in.ShardIndex != nil {
		shardIndex = sql.NullInt64{Int64: int64(*in.ShardIndex), Valid: true}
	}
	if in.ShardTotal != nil {
		shardTotal = sql.NullInt64{Int64: int64(*in.ShardTotal), Valid: true}
	}
	if in.MaxCameras != nil {
		maxCameras = sql.NullInt64{Int64: int64(*in.MaxCameras), Valid: true}
	}

	var cpu, mem, load sql.NullFloat64
	if in.CPUPercent != nil {
		cpu = sql.NullFloat64{Float64: *in.CPUPercent, Valid: true}
	}
	if in.MemPercent != nil {
		mem = sql.NullFloat64{Float64: *in.MemPercent, Valid: true}
	}
	if in.Load1m != nil {
		load = sql.NullFloat64{Float64: *in.Load1m, Valid: true}
	}

	row := r.db.QueryRow(`
INSERT INTO vis_worker (
    worker_id, worker_tipo, hostname, versao, cameras_ativas, ultimo_ping_em, ativo,
    shard_index, shard_total, max_cameras, yolo_device, queue_backend,
    vis_mediamtx_node_id, cpu_percent, mem_percent, load_1m
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
ON CONFLICT (worker_id, worker_tipo, vis_mediamtx_node_id)
DO UPDATE SET
    hostname = EXCLUDED.hostname,
    versao = EXCLUDED.versao,
    cameras_ativas = EXCLUDED.cameras_ativas,
    ultimo_ping_em = EXCLUDED.ultimo_ping_em,
    ativo = EXCLUDED.ativo,
    shard_index = EXCLUDED.shard_index,
    shard_total = EXCLUDED.shard_total,
    max_cameras = EXCLUDED.max_cameras,
    yolo_device = EXCLUDED.yolo_device,
    queue_backend = EXCLUDED.queue_backend,
    cpu_percent = EXCLUDED.cpu_percent,
    mem_percent = EXCLUDED.mem_percent,
    load_1m = EXCLUDED.load_1m
RETURNING id, created_at, worker_id, worker_tipo, hostname, versao, cameras_ativas,
    ultimo_ping_em, ativo, shard_index, shard_total, max_cameras, yolo_device,
    queue_backend, vis_mediamtx_node_id, cpu_percent, mem_percent, load_1m`,
		in.WorkerID, tipo, in.Hostname, in.Versao, in.CamerasAtivas, pingAt, ativo,
		shardIndex, shardTotal, maxCameras, in.YoloDevice, in.QueueBackend,
		nodeID, cpu, mem, load,
	)

	out, err := scanWorkerRow(row)
	if err != nil {
		return nil, err
	}

	if nodeID > 0 {
		_ = r.UpdateNodeMetrics(nodeID, cpu, mem, pingAt)
		_ = r.RecalcNodePoints(nodeID)
	}

	return out, nil
}

func (r *Repository) RecalcNodePoints(nodeID int) error {
	if nodeID <= 0 {
		return nil
	}
	_, err := r.db.Exec(`SELECT vis_mediamtx_node_recalc_pontos($1)`, nodeID)
	return err
}

func (r *Repository) UpdateNodeMetrics(nodeID int, cpu, mem sql.NullFloat64, pingAt time.Time) error {
	_, err := r.db.Exec(`
UPDATE vis_mediamtx_node
SET ultimo_ping_em = $2,
    cpu_percent = $3,
    mem_percent = $4
WHERE id = $1`, nodeID, pingAt, cpu, mem)
	return err
}

func (r *Repository) queryCameras(query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		m, err := scanCameraRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func scanCameraRow(rows *sql.Rows) (map[string]any, error) {
	cols, _ := rows.Columns()
	dest := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range dest {
		ptrs[i] = &dest[i]
	}
	if err := rows.Scan(ptrs...); err != nil {
		return nil, err
	}

	m := make(map[string]any, len(cols))
	for i, col := range cols {
		m[col] = normalizeValue(dest[i])
	}
	return m, nil
}

func scanAreaRow(rows *sql.Rows) (map[string]any, error) {
	var id, cameraID int
	var createdAt time.Time
	var nome, poligono, cor sql.NullString
	var ativo sql.NullBool

	if err := rows.Scan(&id, &createdAt, &cameraID, &nome, &ativo, &poligono, &cor); err != nil {
		return nil, err
	}

	return map[string]any{
		"id":             id,
		"created_at":     createdAt.UTC().Format(time.RFC3339),
		"vis_camera_id":  cameraID,
		"nome":           nullStr(nome),
		"ativo":          nullBool(ativo),
		"poligono_json":  nullStr(poligono),
		"cor":            nullStr(cor),
	}, nil
}

func scanWorkerRow(row *sql.Row) (map[string]any, error) {
	var id, camerasAtivas int
	var createdAt, pingAt time.Time
	var workerID, workerTipo, hostname, versao, yolo, queue sql.NullString
	var ativo sql.NullBool
	var shardIndex, shardTotal, maxCameras, nodeID sql.NullInt64
	var cpu, mem, load sql.NullFloat64

	err := row.Scan(
		&id, &createdAt, &workerID, &workerTipo, &hostname, &versao, &camerasAtivas,
		&pingAt, &ativo, &shardIndex, &shardTotal, &maxCameras, &yolo, &queue,
		&nodeID, &cpu, &mem, &load,
	)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":                   id,
		"created_at":           createdAt.UTC().Format(time.RFC3339),
		"worker_id":            nullStr(workerID),
		"worker_tipo":          nullStr(workerTipo),
		"hostname":             nullStr(hostname),
		"versao":               nullStr(versao),
		"cameras_ativas":       camerasAtivas,
		"ultimo_ping_em":       pingAt.UTC().Format(time.RFC3339),
		"ativo":                nullBool(ativo),
		"shard_index":          nullInt(shardIndex),
		"shard_total":          nullInt(shardTotal),
		"max_cameras":          nullInt(maxCameras),
		"yolo_device":          nullStr(yolo),
		"queue_backend":        nullStr(queue),
		"vis_mediamtx_node_id": nullInt(nodeID),
		"cpu_percent":          nullFloat(cpu),
		"mem_percent":          nullFloat(mem),
		"load_1m":              nullFloat(load),
	}, nil
}

func normalizeValue(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(t)
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	case bool, float64, int64, string:
		return t
	default:
		return t
	}
}

func nullStr(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func nullBool(v sql.NullBool) any {
	if v.Valid {
		return v.Bool
	}
	return nil
}

func nullInt(v sql.NullInt64) any {
	if v.Valid {
		return int(v.Int64)
	}
	return nil
}

func nullFloat(v sql.NullFloat64) any {
	if v.Valid {
		return v.Float64
	}
	return nil
}

func buildConfigVersion(cameras []map[string]any, areas []map[string]any) string {
	b := strings.Builder{}
	b.WriteString(strconv.Itoa(len(cameras)))
	b.WriteString("c-")
	b.WriteString(strconv.Itoa(len(areas)))
	b.WriteString("a")
	for _, c := range cameras {
		id := cameraIDFromMap(c)
		b.WriteString("-")
		b.WriteString(strconv.FormatInt(id, 10))
	}
	return b.String()
}

func cameraIDFromMap(c map[string]any) int64 {
	switch v := c["id"].(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		i, _ := v.Int64()
		return i
	default:
		return 0
	}
}
