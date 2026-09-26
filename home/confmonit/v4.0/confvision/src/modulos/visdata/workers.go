package visdata

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

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
	CameraStreamHealth []CameraStreamHealthInput `json:"camera_stream_health"`
}

func UpsertWorkerPing(ctx context.Context, in WorkerPingInput) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
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

	row := db.QueryRowContext(ctx, `
INSERT INTO vis_worker (
    worker_id, worker_tipo, hostname, versao, cameras_ativas, ultimo_ping_em, ativo,
    shard_index, shard_total, max_cameras, yolo_device, queue_backend,
    vis_mediamtx_node_id, cpu_percent, mem_percent, load_1m
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
ON CONFLICT (worker_id, worker_tipo, vis_mediamtx_node_id)
DO UPDATE SET
    hostname = EXCLUDED.hostname, versao = EXCLUDED.versao,
    cameras_ativas = EXCLUDED.cameras_ativas, ultimo_ping_em = EXCLUDED.ultimo_ping_em,
    ativo = EXCLUDED.ativo, shard_index = EXCLUDED.shard_index,
    shard_total = EXCLUDED.shard_total, max_cameras = EXCLUDED.max_cameras,
    yolo_device = EXCLUDED.yolo_device, queue_backend = EXCLUDED.queue_backend,
    cpu_percent = EXCLUDED.cpu_percent, mem_percent = EXCLUDED.mem_percent,
    load_1m = EXCLUDED.load_1m
RETURNING id, created_at, worker_id, worker_tipo, hostname, versao, cameras_ativas,
    ultimo_ping_em, ativo, shard_index, shard_total, max_cameras, yolo_device,
    queue_backend, vis_mediamtx_node_id, cpu_percent, mem_percent, load_1m`,
		in.WorkerID, tipo, in.Hostname, in.Versao, in.CamerasAtivas, pingAt, ativo,
		nullIntPtr(in.ShardIndex), nullIntPtr(in.ShardTotal), nullIntPtr(in.MaxCameras),
		in.YoloDevice, in.QueueBackend, nodeID,
		nullFloatPtr(in.CPUPercent), nullFloatPtr(in.MemPercent), nullFloatPtr(in.Load1m),
	)

	var id, camerasAtivas int
	var createdAt, pingOut time.Time
	var workerID, workerTipo, hostname, versao, yolo, queue sql.NullString
	var ativoOut sql.NullBool
	var shardIndex, shardTotal, maxCameras, nodeOut sql.NullInt64
	var cpu, mem, load sql.NullFloat64
	if err := row.Scan(&id, &createdAt, &workerID, &workerTipo, &hostname, &versao, &camerasAtivas,
		&pingOut, &ativoOut, &shardIndex, &shardTotal, &maxCameras, &yolo, &queue,
		&nodeOut, &cpu, &mem, &load); err != nil {
		return nil, err
	}
	if nodeOut.Valid {
		_ = SyncMediamtxNode(ctx, int(nodeOut.Int64))
	}
	if len(in.CameraStreamHealth) > 0 {
		if err := ApplyCameraStreamHealthBatch(ctx, in.CameraStreamHealth); err != nil {
			return nil, err
		}
	}
	return map[string]any{
		"id": id, "created_at": createdAt.UTC().Format(time.RFC3339),
		"worker_id": nullStr(workerID), "worker_tipo": nullStr(workerTipo),
		"hostname": nullStr(hostname), "versao": nullStr(versao),
		"cameras_ativas": camerasAtivas, "ultimo_ping_em": pingOut.UTC().Format(time.RFC3339),
		"ativo": nullBool(ativoOut), "shard_index": nullInt(shardIndex),
		"shard_total": nullInt(shardTotal), "max_cameras": nullInt(maxCameras),
		"yolo_device": nullStr(yolo), "queue_backend": nullStr(queue),
		"vis_mediamtx_node_id": nullInt(nodeOut), "cpu_percent": nullFloat(cpu),
		"mem_percent": nullFloat(mem), "load_1m": nullFloat(load),
	}, nil
}

func nullIntPtr(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func nullFloatPtr(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}

func buildConfigVersion(cameras, areas []map[string]any) string {
	b := strings.Builder{}
	b.WriteString(strconv.Itoa(len(cameras)))
	b.WriteString("c-")
	b.WriteString(strconv.Itoa(len(areas)))
	b.WriteString("a")
	for _, c := range cameras {
		b.WriteString("-")
		b.WriteString(trimAny(c["id"]))
	}
	return b.String()
}
