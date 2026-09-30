# ConfVision — Arquitetura de Storage

Auditoria de código: 2026-09-30.

---

## 1. Visão alvo vs atual

```text
ATUAL (simplificado)

PostgreSQL ──► metadata (URLs, keys, segmentos, eventos)
Redis      ──► fila + cache sync (não mídia)
Disco node ──► temp + MediaMTX record + motion/timelapse build
S3 Contabo ──► mídia durável (eventos + DVR + motion/timelapse upload)

ALVO INCREMENTAL (Fase 5)

Call sites ──► MediaStorage (Rust) / storage.py + gravacao_storage.py (Python)
            ──► backend S3 hoje; NAS/MinIO futuro sem mudar metadata PG
```

---

## 2. PostgreSQL

### 2.1 Papel

Fonte de verdade para **configuração** e **metadata** de mídia. Não armazena blobs de vídeo (sem `bytea` encontrado no repo).

### 2.2 Tabelas relevantes (amostra via Go `visdata`)

| Tabela | Uso | Crescimento esperado |
|--------|-----|----------------------|
| `vis_camera` | config câmera, `worker_id`, retenção plano | baixo |
| `vis_evento` | eventos analíticos, `snapshot_url`, `video_url` | **alto** |
| `vis_evento_clip` | clips adicionais | médio |
| `vis_gravacao_segmento` | índice DVR (`s3_key`, `s3_url`, janelas tempo) | **muito alto** |
| `vis_gravacao_storage` | bucket/credenciais por franqueado | baixo |
| `vis_worker` | heartbeat workers | baixo |
| `vis_sistema_health` / `vis_sistema_metric` | coleta operacional | médio (retenção **NÃO DEFINIDO** no código) |

### 2.3 Queries críticas

- Sync Rust: listagem câmeras ativas por `worker_id` (`rust_processor_d5.go`, sync handlers).
- Eventos: INSERT/UPDATE `vis_evento`, listagens paginadas (`eventos.go`).
- DVR: INSERT `vis_gravacao_segmento`, list by camera/franqueado (`gravacao.go`).

### 2.4 Volume / tamanho

**Não medido** nesta sessão (requer `POSTGRES_URL` + `pg_total_relation_size`). Script operacional existente: `home/confmonit/v4.0/confvision/scripts/pg-audit/main.go`.

### 2.5 Migrations no repo

- `sql/migrations/20260326_vis_camera_stream_policy.sql`
- `sql/migrations/20260928_vis_coleta_operacional.sql`

---

## 3. Redis

| Key / padrão | Tipo | TTL | Produtor | Consumidor | Classificação |
|--------------|------|-----|----------|------------|---------------|
| `confvision:eventos` | LIST | none | Rust/Python publish | Rust BRPOP capture | **QUEUE** |
| `confvision:eventos:dlq` | LIST | none | Rust on full/error | manual/ops | **QUEUE** |
| `confvision:yolo:queue` | LIST | none | scheduler Python | detector | **QUEUE** |
| `confvision:sync:{kind}:n{node}:w{worker}` | STRING (JSON) | `CONFIG_CACHE_TTL_SEC` | sync-agent | workers Python | **CACHE** |
| `confvision:sync:…:version` | STRING | idem | sync-agent | workers | **CACHE** |
| `confvision:rtmp_auth:*` | cache auth | config | sync-agent | publicadores | **CACHE** |

**Fonte de verdade:** assignment e config de câmera = **Postgres**, não Redis (ADR-013).

---

## 4. Media por tipo

### 4.1 Snapshot / clip de evento (analítico)

| Campo | Valor |
|-------|--------|
| **Local temp** | `CAPTURE_DIR` (default `/tmp/confvision`), subdir `{evento_id}/` Rust; Python `capture.py` work dirs |
| **Formato** | JPEG snapshot, MP4 clip (FFmpeg pontual Rust) |
| **Durável** | S3 keys `eventos/{id_franqueado}/{evento_id}/snapshot.jpg`, `clip_001.mp4` |
| **Metadata** | `vis_evento.snapshot_url`, `video_url` |
| **Upload Rust** | `media/upload.rs` HTTP PUT; wrapper `media/storage.rs` (`MediaStorage`) |
| **Upload Python** | `storage.py` boto3 + `upload_queue.py` retries |
| **Retenção objeto S3** | **NÃO DEFINIDO** no código (lifecycle bucket = ops) |
| **Retenção metadata** | plano `vis_camera.retencao_dias` / licença — **limpeza automática PG+S3 não auditada** |
| **Falha storage** | Rust: warn + `finalizar_evento` sem URL (`capture/process.rs`) |

### 4.2 DVR contínuo

| Campo | Valor |
|-------|--------|
| **Processo** | `dvr_main.py` + `dvr_watcher.py` + `dvr_segment.py` |
| **Origem** | MediaMTX grava em `DVR_RECORD_DIR` (default `/recordings`) por path câmera |
| **Estabilidade** | `DVR_STABLE_SEC` antes de processar |
| **Upload** | `gravacao_storage.py` boto3, credenciais por franqueado via API |
| **Key S3** | `{id_franqueado}/camera-{id}/{filename}` |
| **Metadata** | POST `/vis_gravacao_segmento` → `vis_gravacao_segmento` |
| **Pós-upload** | rename local para `.uploaded` (não apaga raw imediatamente) |
| **Retry** | `DVR_UPLOAD_RETRIES` com backoff sleep |
| **Retenção** | `retencao_dias` em licença/câmera — **job de purge global NÃO DEFINIDO** no código DVR |

### 4.3 Timelapse / motion clip (Python)

| Campo | Valor |
|-------|--------|
| **Local** | `MOTION_RECORD_DIR/timelapse/{camera_id}/` (default base `/tmp/confvision/motion`) |
| **Montagem** | FFmpeg concat → MP4 |
| **Upload** | mesmo stack Contabo + registro segmento tipo `timelapse` / movimento |
| **Retenção** | plano timelapse 7/15/30d em `plano.go` — enforcement storage **NÃO DEFINIDO** |

### 4.4 URLs

- **Públicas:** URLs Contabo path-style montadas em Python/Rust após upload.
- **Internas:** RTSP `MEDIAMTX_RTSP_BASE`, API Go `CONFVISION_API_URL` / Xano bridge.

---

## 5. Abstração `MediaStorage` (Rust)

Arquivo: `src/media/storage.rs`

- `MediaStorage::put` — falha explícita se S3 ausente.
- `put_or_log` — usado no pipeline de captura; **não aborta** evento.

Python equivalente conceitual: `storage.upload_file` + `gravacao_storage.upload_segment_file` (dois clientes S3: global eventos vs per-franqueado DVR).

**Não feito nesta fase:** backend local-only, multipart, delete, exists, listagem orphan.

---

## 6. Retenção e limpeza

| Tipo | Política no código | Status |
|------|-------------------|--------|
| Eventos PG | campos plano/licença | parcial |
| S3 eventos | — | **NÃO DEFINIDO** |
| DVR segmentos PG | insert only + queries | purge **NÃO DEFINIDO** |
| Disco `.uploaded` / `/recordings` | rename only | **NÃO DEFINIDO** |
| `CAPTURE_DIR` | `remove_dir_all` após captura Rust | temp OK |
| `cleanup_work_dir` Python | após event_capture | temp OK |

**Limpeza desejada (futuro):** idade + retenção + cliente + tipo + estado evento; **sem** cron delete indiscriminado.

---

## 7. Orphan / missing media

### 7.1 Definições

- **ORPHAN:** objeto existe no S3 ou disco, sem row correspondente em PG.
- **MISSING MEDIA:** row PG (`vis_evento`, `vis_gravacao_segmento`) com URL/key, objeto ausente.

### 7.2 Mecanismo Fase 5

1. **Detect** — relatório SQL + listagem bucket prefix (script ops, não automático).
2. **Report** — métrica proposta `orphan_media`, `missing_media`.
3. **Validate** — amostragem manual.
4. **Cleanup** — fase posterior, com audit log.

---

## 8. Retry policies (existentes)

| Caminho | Retries | Backoff |
|---------|---------|---------|
| Rust Redis publish | `queue_publish_retries` | 50ms × attempt |
| Python upload | `UPLOAD_RETRIES` / `DVR_UPLOAD_RETRIES` | exponential cap 10s |
| Rust S3 PUT | **single try** | **pendente** alinhar com Python |

---

## 9. Media job (avaliação)

Hoje: **evento + captura inline** no worker Redis (Rust). Upload S3 síncrono dentro de `processar_deteccao`.

**Recomendação futura:** fila `media_jobs` desacoplada — **não implementado** (evita scope creep).

---

## 10. Backup / restore

| Ativo | Estado auditoria |
|-------|------------------|
| PostgreSQL | **Não verificado** — assumir backup infra host; **restore test pendente** |
| Redis | fila efêmera — RPO fila = segundos; rebuild via redetecção **não garantido** |
| S3 | durabilidade provider; versionamento **NÃO DEFINIDO** |
| Config `.env` | manual por node |

---

## 11. Projeção capacidade storage

Sem métricas de produção nesta sessão. Fórmula para ops:

```text
DVR ≈ câmeras_gravando × (86400 / segmento_seg) × tamanho_segmento × retenção_dias
Eventos ≈ detecções/dia × (snapshot_kb + clip_mb)
```

Marcar números como **ESTIMATIVA** até coleta `vis_sistema_metric` / bucket billing.

---

## 12. Métricas propostas (implementação parcial)

Já expostas: profundidade fila Redis, health `event_queue_redis_ok`.

Pendentes: `storage_bytes`, `media_failed`, `orphan_media`, `missing_media`, disk usage node.
