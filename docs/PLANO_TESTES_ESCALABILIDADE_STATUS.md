# ConfVision — Status do plano de testes (Fases 0–40)

**Regra:** não declarar capacidade sem benchmark. **Não alterar arquitetura** durante testes sem autorização.

**Última atualização:** 2026-09-30 · ambiente dev: Windows, `core4` @ `f982038`

---

## Resumo executivo

| Fases | Status |
|-------|--------|
| 0 Inventário | **CONCLUÍDA** → [AUDITORIA_ARQUITETURA.md](./AUDITORIA_ARQUITETURA.md) |
| 1 Build / integridade | **CONCLUÍDA** (clippy strict **FAIL**; Python tests **SKIP**) |
| 2–40 Carga, HA, sintético | **NÃO EXECUTADAS** (requerem staging/hardware/scripts) |

---

## Fase 0 — Inventário

**STATUS:** CONCLUÍDA  
**EVIDÊNCIA:** `docs/AUDITORIA_ARQUITETURA.md`

---

## Fase 1 — Build e integridade

**TESTE:** Fase 1 — Build e integridade  
**OBJETIVO:** Validar toolchain antes de benchmarks.

**AMBIENTE:**  
- CPU/RAM/GPU: host dev Windows (não documentado modelo)  
- NETWORK: N/A  

**COMANDOS / EVIDÊNCIA:**

| Check | Resultado |
|-------|-----------|
| `cargo check` (rust-processor) | **PASS** |
| `cargo test` | **PASS** — 136 tests, 0 failed |
| `cargo clippy -- -D warnings` | **FAIL** — unused imports / dead_code (104 warnings elevados a erro) |
| `cargo build --release` | **PASS** (~4m27s) |
| `go test ./...` (confvision app) | **PASS** — `visdata`, `modulos/confvision` |
| `go build .` | **PASS** |
| `pytest confvision/tests` | **SKIP** — Python não instalado no host dev |
| Migrations SQL | **PRESENTES** — `sql/migrations/20260326_*`, `20260928_*` (+ scripts `run_stream_migrations`) |

**STATUS:** **PASS COM RESSALVAS** (clippy -D warnings; pytest omitido)

---

## Fases 2–40 — Registro por fase

Legenda: **NE** = não executado · **PE** = parcialmente evidenciado (unit/smoke only)

| Fase | Título | Status | Notas |
|------|--------|--------|-------|
| 2 | Processor isolado (1→1000 câmeras) | **NE** | Template: [BENCHMARK_PROCESSOR.md](../confvision-rust-processor/docs/BENCHMARK_PROCESSOR.md) |
| 3 | Pipeline A–D | **NE** | Decomposição requer RTSP lab |
| 4 | Sharding 2–10 processors | **NE** | Unit tests sharding only (**PE**) |
| 5 | Balanceamento CPU heterogêneo | **NE** | D5 Go ausente no tree canônico |
| 6 | Redistribuição worker_id | **NE** | Manual ops |
| 7 | Ownership / lease | **NE** | **GAP documentado** — sem lease |
| 8 | Redis carga/falha | **NE** | |
| 9 | Banco carga | **NE** | |
| 10 | Eventos perda/duplicata | **NE** | Dedup cooldown unit (**PE**) |
| 11 | Falha processor | **NE** | |
| 12 | Falha cascata | **NE** | |
| 13 | Backpressure | **PE** | `DropOldestQueue` unit tests |
| 14 | Escala horizontal | **NE** | |
| 15 | Capacidade limite processor | **NE** | |
| 16 | Estabilidade 24h | **NE** | |
| 17 | Reconnect storm | **NE** | Reconnect delay unit (**PE**) |
| 18 | YOLO OFF/HTTP/ONNX | **NE** | |
| 19 | Detection scheduler | **PE** | Semaphore inflight in-process |
| 20 | Fairness | **NE** | |
| 21 | Multi-tenant | **NE** | Isolamento lógico PG (**PE** doc) |
| 22 | Regional | **NE** | **NÃO ENCONTRADO** |
| 23 | Control vs data plane | **NE** | Arquitetura descrita (**PE**) |
| 24 | Observabilidade | **PE** | Endpoints Rust; smoke remoto ops |
| 25 | Diagnóstico por câmera | **PE** | stream_health parcial |
| 26 | Escala sintética | **NE** | Sem simulador |
| 27 | Teste 1M sintético | **NE** | |
| 28 | Teste 10M conceitual | **NE** | |
| 29 | Global + regional | **NE** | |
| 30 | Failover regional | **NE** | |
| 31 | Recuperação | **NE** | |
| 32 | Consistência registry | **NE** | Sem ferramenta |
| 33 | Segurança operacional | **NE** | Revisão manual pendente |
| 34 | Config runtime | **PE** | Muitas configs via env (restart) |
| 35 | Rolling upgrade | **NE** | |
| 36 | Rollback | **NE** | |
| 37 | Observabilidade distribuída | **NE** | |
| 38 | Capacidade control plane | **NE** | |
| 39 | Capacidade event plane | **NE** | |
| 40 | Relatório final | **PE** | [RELATORIO_ESCALABILIDADE_CONFVISION.md](./RELATORIO_ESCALABILIDADE_CONFVISION.md) |

---

## Formato de registro (template para execuções futuras)

Copiar por teste executado:

```text
TESTE:
OBJETIVO:

AMBIENTE:
CPU:
RAM:
GPU:
NETWORK:

CARGA:
CAMERAS:
FPS:
EVENTS/S:

RESULTADO:
CPU:
RAM:
FPS:
DROPS:
LATENCY:
ERRORS:

GARGALO:

STATUS:
PASS | PASS COM RESSALVAS | FAIL

EVIDÊNCIAS:
```

---

## Onde executar próximos testes

- Runbook: `confvision-rust-processor/docs/SCALE_TESTING.md`
- Scripts: `confvision-rust-processor/scripts/implantar-fases.ps1`, `d5-auto-assign-analiticas.ps1`, `d3-online-test.mjs`
- Staging referenciado: FoxPro EasyPanel (credenciais necessárias)
