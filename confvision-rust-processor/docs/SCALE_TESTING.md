# ConfVision — Scale testing & benchmarks

Reprodutibilidade: registrar **hardware, env, duração, resultado**.

**Plano 40 fases (status):** [docs/PLANO_TESTES_ESCALABILIDADE_STATUS.md](../../docs/PLANO_TESTES_ESCALABILIDADE_STATUS.md) · inventário [AUDITORIA_ARQUITETURA.md](../../docs/AUDITORIA_ARQUITETURA.md) · relatório [RELATORIO_ESCALABILIDADE_CONFVISION.md](../../docs/RELATORIO_ESCALABILIDADE_CONFVISION.md) · template Fase 2 [BENCHMARK_PROCESSOR.md](./BENCHMARK_PROCESSOR.md).

---

## Status Fase 6

| Suite | Status |
|-------|--------|
| Rust unit (`cargo test`) | ✅ 136 tests (local dev) |
| d3-online-test quick/full | ⚠️ requer FoxPro/credenciais |
| Load 10→1000 câmeras | **NÃO EXECUTADO** |
| API 100–1000 rps | **NÃO EXECUTADO** |
| YOLO multi-camera benchmark | **NÃO EXECUTADO** |
| Failure / chaos | ver Fase 5 — **NÃO EXECUTADO** |

---

## 1. Smoke (sempre)

```powershell
Set-Location c:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor
cargo test
```

```text
curl -s https://<rust-host>/health | jq .
curl -s https://<rust-host>/capacity-report | jq .
node confvision-rust-processor/scripts/d3-online-test.mjs --quick
```

---

## 2. Capacity benchmark (Rust)

**Input:** N câmeras reais ou simuladas RTSP no mesmo worker.  
**Duration:** ≥ 30 min (alinhado `capacity_min_sample_sec`).  
**Collect:** snapshots `/capacity-report` a cada 5 min.

**Record:**

| Field | Example |
|-------|---------|
| hardware | CPU model, GPU, RAM |
| N cameras | |
| yolo_* env | |
| avg fps | from report |
| estimated_available | |
| frames_dropped rate | |

---

## 3. Event throughput

**Measure:**

- Redis `LLEN confvision:eventos` over time
- Rust metrics `events_published`, capture latency logs
- PG `COUNT(*) FROM vis_evento WHERE created_at > now()-interval '1 hour'`

**Do not** extrapolate to 100k cameras without linear test points.

---

## 4. API load (Go)

Ferramenta sugerida: `hey`, `k6`, ou `vegeta` contra endpoints read-heavy (`/vis_camera_sync_ativas`) e write (`/vis_worker_ping`).

Registrar p50/p95/p99 e error rate. **Não executado** na Fase 6 doc pass.

---

## 5. PostgreSQL / Redis micro-benchmarks

- PG: `pgbench` no host isolado (não confundir com app queries).
- Redis: `redis-benchmark` baseline; app usa LIST BRPOP/LPUSH.

---

## 6. Load test ladder (câmeras)

| Step | Action |
|------|--------|
| 10 | Baseline capacity-report |
| 50 | Watch GPU/CPU limiting_resource |
| 100 | Admission reject? shedding? |
| 500+ | Simulação parcial se hardware insuficiente — marcar **SIMULADO** |

---

## 7. Failure tests (controlado)

Repetir matriz Fase 5 em **staging** only:

Node kill, Redis stop, PG stop, disk full, MTX restart, Rust crash.

Template de registro:

```text
Cenário:
Detecção (tempo):
Impacto:
Recovery (tempo):
Dados perdidos:
```

---

## 8. Result archive

Salvar em ops (fora do repo): `benchmarks/YYYY-MM-DD-<host>.json` com capacity-report series + env redacted.
