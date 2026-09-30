# Benchmark — Rust Processor isolado (Fase 2)

**Status:** **NÃO EXECUTADO** (2026-09-30). Este arquivo é o template de registro; preencher após testes em staging com RTSP real ou simulador.

**Regra:** distinguir **TESTE REAL** vs **SIMULADO** vs **PROJEÇÃO**. Não usar duração &lt; 30 min como capacidade sustentável (ver `capacity_min_sample_sec` e [SCALE_TESTING.md](./SCALE_TESTING.md)).

---

## Pré-requisitos

- Processor build: `cargo build --release`
- Env documentado (redacted): `WORKER_ID`, YOLO_*, `FRAME_BUFFER_MAX`, `LOAD_*`, `REDIS_URL`
- N câmeras RTSP estáveis (ou lab MediaMTX)
- Coleta: `/capacity-report` a cada 5 min + `/metrics` + logs

---

## Matriz de cenários (planejada)

| Câmeras | Duração mín. | Objetivo |
|---------|--------------|----------|
| 1 | 30 min | Baseline FPS/latência |
| 5 | 30 min | |
| 10 | 30 min | |
| 20 | 30 min | |
| 50 | 30 min | Primeiro sinal de admission/shedding |
| 100 | 30 min+ | Limite operacional candidato |
| 200+ | — | Só se hardware staging permitir |

---

## Registro por run (copiar blocos)

### Run: `<YYYY-MM-DD>-N=<cameras>`

```text
TESTE: Fase 2 — Processor isolado N=<cameras>
OBJETIVO: Medir CPU/RAM/FPS/drops/YOLO sob carga sustentada

AMBIENTE:
CPU: <modelo>
RAM: <GB>
GPU: <modelo / none>
NETWORK: <RTSP LAN / WAN>

CARGA:
CAMERAS: <N>
FPS: <config / medido>
EVENTS/S: <medido>

RESULTADO:
CPU: <% avg / p95>
RAM: <GB>
FPS: <received / processed>
DROPS: <frames / rate>
LATENCY: <decode / yolo / event p95>
ERRORS: <RTSP reconnects / API / redis>

GARGALO: <decode | yolo | cpu | network | redis | unknown>

STATUS: PASS | PASS COM RESSALVAS | FAIL

EVIDÊNCIAS:
- capacity-report JSON series
- comando cargo / commit
- host metrics
```

---

## Capacidade sustentável (preencher após runs)

| Métrica | Valor medido | Condições |
|---------|--------------|-----------|
| Câmeras sustentáveis | **TBD** | CPU &lt; X%, drops &lt; Y%, YOLO p95 &lt; Z ms |
| Limiting resource | **TBD** | from `/capacity-report` |
| Notas | | |

**CAPACIDADE MÁXIMA OBSERVADA (picos):** TBD — não confundir com sustentável.
