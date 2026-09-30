# ConfVision — Capacity Planning

Separar **MEDIDO** vs **ESTIMADO**. Valores MEDIDO só entram após benchmark no ambiente anotado.

---

## Modelo por componente

| Componente | Driver principal | Métrica chave |
|------------|------------------|---------------|
| Rust Node | decode + YOLO + motion | CPU/GPU %, FPS agregado, `frames_dropped` |
| MediaMTX | bitrate × câmeras | rede + disco record |
| Redis | eventos/s | `LLEN`, publish errors |
| PostgreSQL | INSERT evento + segmento DVR | write TPS, tamanho tabelas |
| S3 | GB/dia upload | billing + segmento count |
| Go API | req/s sync/ping | latência p95 |

---

## Cenários (projeção)

### 100 câmeras analíticas

| Recurso | ESTIMADO | Notas |
|---------|----------|--------|
| Rust nodes | 2–4 workers | Depende GPU; usar D5 assign |
| CPU/RAM/node | MEDIDO pendente | `/capacity-report` |
| Redis | 1 instância | fila << 1000 default max |
| PG | instância atual | eventos ~ detecções/dia |
| Network | Σ bitrate | **medir** |

### 1.000 câmeras

| Recurso | ESTIMADO |
|---------|----------|
| Rust nodes | 10–40 | **forte dependência GPU** |
| PG | índices `vis_evento`, archiving **avaliar** |
| Redis | monitorar lag capture |
| CP Go | 2+ instâncias LB **recomendado** |

### 10.000 / 100.000+

Arquitetura **teórica:** shards por região/pool, read replicas PG, Redis cluster, object storage multi-bucket.

**TESTED LIMIT atual:** não comprovado nesta fase.

---

## Fórmulas úteis (ESTIMADO)

```text
events_day ≈ cameras × detections_per_camera_day
dvr_storage_day ≈ cameras_dvr × (86400/segment_sec) × avg_segment_mb
redis_memory ≈ queue_depth × avg_job_bytes + cache JSON
```

---

## Como obter MEDIDO

1. `GET /capacity-report` em cada processor (salvar JSON horário pico).
2. `go run ./scripts/pg-audit` + `pg_total_relation_size`.
3. Redis `INFO memory`, `LLEN confvision:eventos`.
4. Coleta `vis_sistema_metric` (`rust_processors`, `vis_camera_resumo`).
5. Load test progressivo — ver `SCALE_TESTING.md`.

---

## Headroom recomendado (ops)

Manter `capacity_state` ≠ `critical` e `allow_new_camera=true` antes de assign em massa.

Não planejar 100% do `estimated_available_cameras` — margem **A DEFINIR** por negócio (sugestão ops: 70–80% do estimador após MEDIDO).
