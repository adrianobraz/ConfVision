# ConfVision — Operations Runbook

Procedimentos baseados no stack auditado (Go + Rust + Python + MediaMTX).  
Substituir URLs/hosts pelos do ambiente.

---

## Node / worker offline

1. **Heartbeat:** `SELECT worker_id, ultimo_ping_em, ativo, cameras_ativas FROM vis_worker WHERE worker_tipo='rust_processor' ORDER BY ultimo_ping_em DESC;`
2. **Rede:** ping/curl `https://<rust-host>/health` e `/ready`.
3. **Rust logs:** buscar `shutdown`, `redis PING falhou`, `sync falhou`.
4. **MediaMTX:** API `:9997` paths; RTSP `:8554`.
5. **PostgreSQL / API Go:** `curl` endpoint sync usado pelo Rust (`/vis_camera_sync_ativas`).
6. **Reassignment:** se node irrecuperável, `UPDATE vis_camera SET worker_id='<outro>' WHERE …` + D5 `AutoAssignCameraWorker` ou manual. **Sem failover automático.**

Correlação: node offline → muitas câmeras `stream_falhas_consecutivas` → fila Redis estagnada.

---

## Câmera offline

1. **Cadastro:** `vis_camera` RTSP/RTMP, `bloqueado`, `analitico_pausado`.
2. **Stream health:** campos `stream_*`, `ultimo_stream_ok_em` (PG).
3. **MediaMTX:** path `cam/{hash}` publicado?
4. **Rust:** `/health` entrada da câmera; logs RTSP 404/timeout.
5. **Câmera física:** rede/firewall (fora do ConfVision).

Painel: registros `vis_stream_relatorio` (coleta deduplicada 30 min).

---

## Storage cheio

1. **Volumes:** `DVR_RECORD_DIR`, `CAPTURE_DIR`, `MOTION_RECORD_DIR`, disco root do container.
2. **Tipo:** DVR `.uploaded` local vs S3 backlog vs temp capture.
3. **Retenção:** plano `retencao_dias` — job automático **não** confiar sem verificar.
4. **Orphan:** ver `STORAGE_ARCHITECTURE.md` (detect/report only).
5. **Liberar:** conforme política aprovada — **não** delete em massa sem validação.

---

## Redis degradado

1. Rust `/health` → `event_queue_redis_ok`.
2. Profundidade fila + DLQ (`metrics` / Redis `LLEN confvision:eventos`).
3. Impacto: capture para; RTSP pode continuar.
4. Mitigação: restaurar Redis; drenar DLQ manualmente se necessário.

---

## PostgreSQL degradado

1. Go API errors; Rust `create_evento` falha.
2. Pool: uma instância Go = max 25 conexões — escalar **instâncias** Go, não câmeras→PG.
3. Restore: ver `DISASTER_RECOVERY.md`.

---

## Sobrecarga CPU/GPU

1. Rust `/capacity-report` → `capacity_state`, `limiting_resource`, `estimated_available_cameras`.
2. `load_advisory` / shedding ativo (logs `load shedding`).
3. Ações: reduzir câmeras no worker, aumentar stride YOLO, segundo processor D5 assign, pausar analítico (`analitico_pausado`).

---

## Deploy / rollback (rolling manual)

1. **Canary:** atualizar **um** `RUST_PROCESSOR_BASE_URLS` entry; observar `/health` + coleta 30 min.
2. **Expand:** demais nodes; manter CP compatível (API backward).
3. **Rollback:** redeploy imagem/tag anterior; `processor_version` no ping confirma.
4. **Nunca** dois processos com mesmo `WORKER_ID`.

---

## Escalation data to collect

- Timestamp UTC
- `worker_id`, `processor_id`, `camera_id`, `evento_id` se aplicável
- Trecho log (sem RTSP credentials)
- `/capacity-report` JSON
- Resultado `node scripts/d3-online-test.mjs --quick`
