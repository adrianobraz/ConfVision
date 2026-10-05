package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const workerTipoAnalitico = "analitico"

// CameraNeedsWorkerID indica se a câmera precisa de worker analítico no cadastro.
func CameraNeedsWorkerID(plano string, deteccaoHumano, capturaAnalitico bool) bool {
	if strings.EqualFold(strings.TrimSpace(plano), "online") {
		return false
	}
	return deteccaoHumano || capturaAnalitico
}

// ValidateFranqueadoInfraPool exige pool MTX e worker analitico configurados (obrigatorio por franqueado).
func ValidateFranqueadoInfraPool(ctx context.Context, idFranqueado string) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return fmt.Errorf("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	var nMtx int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_franqueado_mediamtx_pool p
JOIN vis_mediamtx_node n ON n.id = p.vis_mediamtx_node_id
WHERE p.id_franqueado = $1 AND p.ativo IS TRUE AND COALESCE(n.status, 'ativo') = 'ativo'`, idFranqueado).Scan(&nMtx)
	if err != nil {
		return fmt.Errorf("Infra MediaMTX: configure o pool no Administrator (migration 022 aplicada?)")
	}
	if nMtx < 1 {
		return fmt.Errorf("Infra MediaMTX nao configurada para este franqueado — cadastre ao menos um no MediaMTX no Administrator")
	}
	var nWrk int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vis_franqueado_worker_pool
WHERE id_franqueado = $1 AND ativo IS TRUE AND worker_tipo = $2`, idFranqueado, workerTipoAnalitico).Scan(&nWrk)
	if err != nil {
		return fmt.Errorf("Infra worker: configure o pool no Administrator")
	}
	if nWrk < 1 {
		return fmt.Errorf("Infra worker analitico nao configurada para este franqueado — cadastre ao menos um worker no Administrator")
	}
	return nil
}

type infraMtxPick struct {
	NodeID      int
	PoolMax     int
	NodeMax     int
	FranqueadoU int
	GlobalU     int
}

// pickMediamtxForFranqueadoTx escolhe nó MTX do pool (menor carga, pula cheios).
func pickMediamtxForFranqueadoTx(ctx context.Context, tx *sql.Tx, idFranqueado string) (*MediamtxNode, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	rows, err := tx.QueryContext(ctx, `
SELECT p.vis_mediamtx_node_id,
       COALESCE(p.max_cameras, n.max_cameras, 200),
       COALESCE(n.max_cameras, 200),
       COALESCE(p.ordem, 1)
FROM vis_franqueado_mediamtx_pool p
JOIN vis_mediamtx_node n ON n.id = p.vis_mediamtx_node_id
WHERE p.id_franqueado = $1 AND p.ativo IS TRUE AND COALESCE(n.status, 'ativo') = 'ativo'
ORDER BY p.ordem ASC, p.id ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type mtxRow struct {
		nodeID, poolMax, nodeMax, ordem int
	}
	var poolRows []mtxRow
	for rows.Next() {
		var r mtxRow
		if err := rows.Scan(&r.nodeID, &r.poolMax, &r.nodeMax, &r.ordem); err != nil {
			return nil, err
		}
		poolRows = append(poolRows, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	var candidates []infraMtxPick
	for _, r := range poolRows {
		var frU, glU int
		err = tx.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*)::int FROM vis_camera c
   WHERE c.id_franqueado = $1 AND c.vis_mediamtx_node_id = $2 AND c.vis_licenca_id IS NOT NULL),
  (SELECT COUNT(*)::int FROM vis_camera c WHERE c.vis_mediamtx_node_id = $2 AND c.vis_licenca_id IS NOT NULL)`,
			idFranqueado, r.nodeID).Scan(&frU, &glU)
		if err != nil {
			return nil, err
		}
		effectiveMax := r.poolMax
		if r.nodeMax < effectiveMax {
			effectiveMax = r.nodeMax
		}
		if glU >= r.nodeMax {
			continue
		}
		if frU >= effectiveMax {
			continue
		}
		candidates = append(candidates, infraMtxPick{
			NodeID: r.nodeID, PoolMax: effectiveMax, NodeMax: r.nodeMax, FranqueadoU: frU, GlobalU: glU,
		})
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("MediaMTX cheio para este franqueado — libere uma camera ou solicite novo servidor/capacidade no Administrator")
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.FranqueadoU < best.FranqueadoU || (c.FranqueadoU == best.FranqueadoU && c.NodeID < best.NodeID) {
			best = c
		}
	}
	var n MediamtxNode
	var total int
	err = tx.QueryRowContext(ctx, `
SELECT n.id, n.nome, n.rtmp_public, n.hls_public, n.rtsp_internal, n.rtsp_public,
       COALESCE(n.max_cameras, 200), COALESCE(n.ordem, 1), COALESCE(n.status, 'ativo'),
       (SELECT COUNT(*)::int FROM vis_camera c WHERE c.vis_mediamtx_node_id = n.id)
FROM vis_mediamtx_node n WHERE n.id = $1`, best.NodeID).Scan(
		&n.ID, &n.Nome, &n.RtmpPublic, &n.HlsPublic, &n.RtspInternal, &n.RtspPublic,
		&n.MaxCameras, &n.Ordem, &n.Status, &total,
	)
	if err != nil {
		return nil, err
	}
	n.CamerasAtribuidas = total
	return &n, nil
}

type infraWorkerPick struct {
	WorkerID    string
	MaxCameras  int
	FranqueadoU int
}

// pickWorkerForFranqueadoTx escolhe worker analítico do pool (menor carga, pula cheios).
func pickWorkerForFranqueadoTx(ctx context.Context, tx *sql.Tx, idFranqueado string) (string, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	rows, err := tx.QueryContext(ctx, `
SELECT worker_id, COALESCE(max_cameras, 30), COALESCE(ordem, 1)
FROM vis_franqueado_worker_pool
WHERE id_franqueado = $1 AND ativo IS TRUE AND worker_tipo = $2
ORDER BY ordem ASC, id ASC`, idFranqueado, workerTipoAnalitico)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type workerRow struct {
		wid           string
		maxCam, ordem int
	}
	var poolRows []workerRow
	for rows.Next() {
		var r workerRow
		if err := rows.Scan(&r.wid, &r.maxCam, &r.ordem); err != nil {
			return "", err
		}
		r.wid = strings.TrimSpace(r.wid)
		if r.wid == "" {
			continue
		}
		poolRows = append(poolRows, r)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if err := rows.Close(); err != nil {
		return "", err
	}

	var candidates []infraWorkerPick
	for _, r := range poolRows {
		var frU int
		err = tx.QueryRowContext(ctx, `
SELECT COUNT(*)::int FROM vis_camera c
WHERE c.id_franqueado = $1 AND c.worker_id = $2 AND c.vis_licenca_id IS NOT NULL
  AND (c.deteccao_humano IS TRUE OR c.captura_analitico IS TRUE)`, idFranqueado, r.wid).Scan(&frU)
		if err != nil {
			return "", err
		}
		if frU >= r.maxCam {
			continue
		}
		candidates = append(candidates, infraWorkerPick{WorkerID: r.wid, MaxCameras: r.maxCam, FranqueadoU: frU})
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("Workers de processamento cheios para este franqueado — libere uma camera ou aumente capacidade no Administrator")
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if c.FranqueadoU < best.FranqueadoU || (c.FranqueadoU == best.FranqueadoU && c.WorkerID < best.WorkerID) {
			best = c
		}
	}
	return best.WorkerID, nil
}

func infraPoolHasActiveMtx(rows []map[string]any) bool {
	for _, row := range rows {
		if v, ok := row["ativo"].(bool); ok && v {
			return true
		}
	}
	return false
}

func infraPoolHasActiveWorkerAnalitico(rows []map[string]any) bool {
	for _, row := range rows {
		if v, ok := row["ativo"].(bool); !ok || !v {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(fmt.Sprint(row["worker_tipo"])), workerTipoAnalitico) {
			return true
		}
	}
	return false
}

// InfraPoolStatus resumo para Administrator.
func InfraPoolStatus(ctx context.Context, idFranqueado string) (map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	mtx, err := listFranqueadoMediamtxPool(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	wrk, err := listFranqueadoWorkerPool(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	okMtx := infraPoolHasActiveMtx(mtx)
	okWrk := infraPoolHasActiveWorkerAnalitico(wrk)
	return map[string]any{
		"id_franqueado":     idFranqueado,
		"infra_completa":    okMtx && okWrk,
		"tem_mediamtx_pool": okMtx,
		"tem_worker_pool":   okWrk,
		"mediamtx":          mtx,
		"workers":           wrk,
	}, nil
}

func listFranqueadoMediamtxPool(ctx context.Context, idFranqueado string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT p.id, p.vis_mediamtx_node_id, n.nome,
       COALESCE(p.max_cameras, n.max_cameras, 200),
       COALESCE(n.max_cameras, 200),
       p.ordem, p.ativo,
       (SELECT COUNT(*)::int FROM vis_camera c
        WHERE c.id_franqueado = $1 AND c.vis_mediamtx_node_id = p.vis_mediamtx_node_id AND c.vis_licenca_id IS NOT NULL),
       (SELECT COUNT(*)::int FROM vis_camera c WHERE c.vis_mediamtx_node_id = p.vis_mediamtx_node_id AND c.vis_licenca_id IS NOT NULL)
FROM vis_franqueado_mediamtx_pool p
JOIN vis_mediamtx_node n ON n.id = p.vis_mediamtx_node_id
WHERE p.id_franqueado = $1
ORDER BY p.ordem ASC, p.id ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var pid, nodeID, poolMax, nodeMax, ordem int
		var nome string
		var ativo bool
		var frU, glU int
		if err := rows.Scan(&pid, &nodeID, &nome, &poolMax, &nodeMax, &ordem, &ativo, &frU, &glU); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id": pid, "vis_mediamtx_node_id": nodeID, "node_nome": nome,
			"max_cameras": poolMax, "node_max_cameras": nodeMax, "ordem": ordem, "ativo": ativo,
			"uso_franqueado": frU, "uso_global_no": glU,
		})
	}
	return out, rows.Err()
}

func listFranqueadoWorkerPool(ctx context.Context, idFranqueado string) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT p.id, p.worker_id, p.worker_tipo, p.max_cameras, p.ordem, p.ativo,
       (SELECT COUNT(*)::int FROM vis_camera c
        WHERE c.id_franqueado = $1 AND c.worker_id = p.worker_id AND c.vis_licenca_id IS NOT NULL
          AND (c.deteccao_humano IS TRUE OR c.captura_analitico IS TRUE))
FROM vis_franqueado_worker_pool p
WHERE p.id_franqueado = $1
ORDER BY p.ordem ASC, p.id ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var pid, maxCam, ordem int
		var wid, tipo string
		var ativo bool
		var uso int
		if err := rows.Scan(&pid, &wid, &tipo, &maxCam, &ordem, &ativo, &uso); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id": pid, "worker_id": wid, "worker_tipo": tipo, "max_cameras": maxCam,
			"ordem": ordem, "ativo": ativo, "uso_franqueado": uso,
		})
	}
	return out, rows.Err()
}

// UpsertFranqueadoMediamtxPool insere ou atualiza linha do pool MTX.
func UpsertFranqueadoMediamtxPool(ctx context.Context, input map[string]any) (map[string]any, error) {
	idFra := strings.TrimSpace(strVal(input, "id_franqueado"))
	nodeID := intVal(input, "vis_mediamtx_node_id")
	if idFra == "" || nodeID <= 0 {
		return nil, fmt.Errorf("id_franqueado e vis_mediamtx_node_id obrigatorios")
	}
	maxCamVal := intVal(input, "max_cameras")
	var maxCam any
	if maxCamVal > 0 {
		maxCam = maxCamVal
	}
	ordem := intVal(input, "ordem")
	if ordem <= 0 {
		ordem = 1
	}
	ativo := true
	if v := boolVal(input, "ativo"); v != nil {
		ativo = *v
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_franqueado_mediamtx_pool (id_franqueado, vis_mediamtx_node_id, max_cameras, ordem, ativo, updated_at)
VALUES ($1,$2,$3,$4,$5,NOW())
ON CONFLICT (id_franqueado, vis_mediamtx_node_id) DO UPDATE SET
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    ativo = EXCLUDED.ativo,
    updated_at = NOW()
RETURNING id`, idFra, nodeID, maxCam, ordem, ativo).Scan(&id)
	if err != nil {
		return nil, humanizeDBErr(err)
	}
	return map[string]any{"id": id, "id_franqueado": idFra, "vis_mediamtx_node_id": nodeID}, nil
}

// UpsertFranqueadoWorkerPool insere ou atualiza linha do pool worker.
func UpsertFranqueadoWorkerPool(ctx context.Context, input map[string]any) (map[string]any, error) {
	idFra := strings.TrimSpace(strVal(input, "id_franqueado"))
	wid := strings.TrimSpace(strVal(input, "worker_id"))
	if idFra == "" || wid == "" {
		return nil, fmt.Errorf("id_franqueado e worker_id obrigatorios")
	}
	tipo := strings.TrimSpace(strVal(input, "worker_tipo"))
	if tipo == "" {
		tipo = workerTipoAnalitico
	}
	maxCam := intVal(input, "max_cameras")
	if maxCam <= 0 {
		maxCam = 30
	}
	ordem := intVal(input, "ordem")
	if ordem <= 0 {
		ordem = 1
	}
	ativo := true
	if v := boolVal(input, "ativo"); v != nil {
		ativo = *v
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_franqueado_worker_pool (id_franqueado, worker_id, worker_tipo, max_cameras, ordem, ativo, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (id_franqueado, worker_id, worker_tipo) DO UPDATE SET
    max_cameras = EXCLUDED.max_cameras,
    ordem = EXCLUDED.ordem,
    ativo = EXCLUDED.ativo,
    updated_at = NOW()
RETURNING id`, idFra, wid, tipo, maxCam, ordem, ativo).Scan(&id)
	if err != nil {
		return nil, humanizeDBErr(err)
	}
	return map[string]any{"id": id, "id_franqueado": idFra, "worker_id": wid}, nil
}

// ListMediamtxNodesAtivos lista nós para o Administrator montar selects.
// ListCameraIDsNeedingInfraAssign lista câmeras do franqueado sem nó ou sem worker (analítico).
func ListCameraIDsNeedingInfraAssign(ctx context.Context, idFranqueado string) ([]int, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id FROM vis_camera
WHERE id_franqueado = $1 AND vis_licenca_id IS NOT NULL
  AND (
    vis_mediamtx_node_id IS NULL OR vis_mediamtx_node_id <= 0
    OR (
      (deteccao_humano IS TRUE OR captura_analitico IS TRUE)
      AND COALESCE(NULLIF(TRIM(worker_id), ''), '') = ''
    )
  )
ORDER BY id ASC`, strings.TrimSpace(idFranqueado))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// AssignCameraInfraFromPool preenche vis_mediamtx_node_id e worker_id usando o pool (backfill).
func AssignCameraInfraFromPool(ctx context.Context, cameraID int, dryRun bool) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var idFra, plano, workerCur sql.NullString
	var nodeCur sql.NullInt64
	var detHum, capAn sql.NullBool
	err = tx.QueryRowContext(ctx, `
SELECT id_franqueado, plano, worker_id, vis_mediamtx_node_id, deteccao_humano, captura_analitico
FROM vis_camera WHERE id = $1 FOR UPDATE`, cameraID).Scan(
		&idFra, &plano, &workerCur, &nodeCur, &detHum, &capAn)
	if err != nil {
		return nil, err
	}
	idFranqueado := strings.TrimSpace(idFra.String)
	if idFranqueado == "" {
		return nil, fmt.Errorf("camera sem id_franqueado")
	}

	nodeID := int(nodeCur.Int64)
	if nodeID <= 0 {
		node, err := pickMediamtxForFranqueadoTx(ctx, tx, idFranqueado)
		if err != nil {
			return nil, err
		}
		nodeID = node.ID
	}
	needsWorker := CameraNeedsWorkerID(plano.String, detHum.Bool, capAn.Bool)
	workerID := strings.TrimSpace(workerCur.String)
	if needsWorker && workerID == "" {
		workerID, err = pickWorkerForFranqueadoTx(ctx, tx, idFranqueado)
		if err != nil {
			return nil, err
		}
	}
	if !needsWorker {
		workerID = ""
	}

	out := map[string]any{
		"vis_mediamtx_node_id": nodeID,
		"worker_id":            workerID,
		"dry_run":              dryRun,
	}
	if dryRun {
		return out, tx.Rollback()
	}
	_, err = tx.ExecContext(ctx, `
UPDATE vis_camera SET vis_mediamtx_node_id = $2, worker_id = $3 WHERE id = $1`,
		cameraID, nodeID, nullStrOrNil(workerID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	_ = SyncMediamtxNode(ctx, nodeID)
	_, _, _ = EnsureRtspURLSecForCamera(ctx, cameraID)
	return out, nil
}

func ListMediamtxNodesAtivos(ctx context.Context) ([]map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, nome, COALESCE(max_cameras, 200), COALESCE(status, 'ativo')
FROM vis_mediamtx_node
ORDER BY ordem ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, maxCam int
		var nome, status string
		if err := rows.Scan(&id, &nome, &maxCam, &status); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id": id, "nome": nome, "max_cameras": maxCam, "status": status,
		})
	}
	return out, rows.Err()
}
