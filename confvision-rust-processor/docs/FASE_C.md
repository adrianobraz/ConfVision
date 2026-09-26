# Fase C — preparar escala (VPS foxpro + roadmap dedicado)

Fase C vem **depois** da Fase 1 (Rust ingest + sync + HTTP + deploy). Objetivo: **medir**, **observar**, **proteger CPU**, **repartir câmeras** e **preparar fila** — sem reescrever o worker Python inteiro nem exigir GPU na VPS atual.

## Resumo em uma frase

Instalar “painel e limites” no piloto **hoje na VPS**; deixar **GPU, fila produtiva e centenas de câmeras** para o **servidor dedicado (GEX44 ou similar)**.

---

## O que dá para fazer **na VPS foxpro (EasyPanel)** — Fase C completa *operacional*

| Bloco | Na VPS? | O que fazer aqui |
|-------|---------|------------------|
| **C1** Prova A/B | **Sim** | Script `scripts/c1-ab-baseline.sh` + mover 1 câmera entre Python e Rust; anotar CPU/RAM (EasyPanel + `/metrics`). |
| **C2** Observabilidade | **Sim** | Scrape `/health`, `/metrics`, `/capacity-report` (Prometheus + alertas em `deploy/observability/`). Uptime Kuma/curl também serve. |
| **C3** Admission | **Sim** | `LOAD_ADMISSION_ENABLED=1`, `LOAD_POLICY_MODE=admission`, `MAX_CAMERAS` = teto **realista** (ex.: 10, não 30). Ver `easypanel.env.fase-c.vps.example`. |
| **C4** 2º processor | **Parcial** | **Possível** 2º app EasyPanel (`rust-processor-pilot-02`, porta/host diferente) + SQL `worker_id` — **só se CPU/RAM couber**; hoje 10 cam já saturam CPU. |
| **C5** GPU / NVDEC | **Não** | VPS sem placa → manter CPU; logs `gpu_detected=false` são esperados. |
| **C6** Fila Redis | **Parcial** | Subir **Redis** na rede interna (compose exemplo); Rust hoje só **registra** `REDIS_URL` / `QUEUE_BACKEND` no ping — **enqueue prod = Fase D**. |

**Conclusão:** dá para fechar **C1+C2+C3** na VPS de forma sólida; **C4** como piloto leve (2 processors × poucas câmeras) se aliviar carga; **C5** pular; **C6** = infra pronta + contrato documentado.

---

## O que fica para **servidor dedicado (GEX44 / GPU)**

| Item | Por quê não na VPS atual |
|------|---------------------------|
| **C5** `VIDEO_ACCELERATION=gpu`, NVDEC, `hardware_decode_active=true` | Requer NVIDIA + `nvidia-container-toolkit`. |
| **C4** escala **N processors** (dezenas/centenas de câmeras) | CPU da VPS já `capacity_state=critical` com ~10 streams decode CPU. |
| **C6** fila **produtiva** (YOLO batch, clips, alarmes) | Redis + workers GPU + `/dev/shm` frame store — ver `easypanel/worker.env.gex44-gpu.example`. |
| **C1** A/B “justo” com YOLO distribuído | Comparar Rust+GPU node vs Python legacy na mesma classe de hardware. |
| **D1–D3** (Fase 2 produto) | Multi-tenant, YOLO 1→N streams, SLO 24/7. |

Referência env GPU: [`easypanel/worker.env.gex44-gpu.example`](../../easypanel/worker.env.gex44-gpu.example).

---

## Ordem de execução (VPS — uma passada)

1. **Postgres:** câmera **4** (e 2,8,9 se ainda no Rust) → `worker_id` Python (`sql/phase_c_worker_split.example.sql`).
2. **EasyPanel rust-pilot:** copiar env de [`easypanel.env.fase-c.vps.example`](../easypanel.env.fase-c.vps.example) → Redeploy.
3. **C2:** aplicar scrape Prometheus (`deploy/observability/`) ou cron + `scripts/phase-c-verify.sh`.
4. **C3:** confirmar `load.allow_new_camera=false` em critical quando admission ON (`scripts/capacity-report.sh`).
5. **C1:** rodar `scripts/c1-ab-baseline.sh` (janela parada + movimento).
6. **C4 (opcional VPS):** segundo serviço + env [`easypanel.env.fase-c.processor-02.example`](../easypanel.env.fase-c.processor-02.example) + SQL split.
7. **C6 (opcional VPS):** `deploy/phase-c/redis-compose.example.yml` na rede foxpro; manter `QUEUE_BACKEND=none` até Fase D.

Implantação passo a passo: **`docs/FASE_C_IMPLANTACAO_EASYPANEL.md`**  
Análise 2026-09-26: **`docs/FASE_C_ANALISE_2026-09-26.md`**

Validação única:

```bash
bash scripts/phase-c-run-all.sh https://foxpro-rust-pilot.rkr351.easypanel.host
bash scripts/phase-c-verify.sh --strict-c3 https://foxpro-rust-pilot.rkr351.easypanel.host
```

---

## Critérios PASS Fase C (VPS)

- [ ] `rtsp_404_count=0` estável em `/capacity-report`
- [ ] `LOAD_ADMISSION_ENABLED=1` e log de rejeição se tentar 11ª câmera em critical (teste controlado)
- [ ] Prometheus ou verify script cron ≥ 1 scrape/15 min
- [ ] Nota C1 preenchida (`docs/c1-ab-results.template.md` ou wiki)
- [ ] Documento “dedicado pendente” lido pelo time (`FASE_C_DEDICADO_PENDENTE.md`)

---

## Arquivos desta fase

| Arquivo | Uso |
|---------|-----|
| `easypanel.env.fase-c.vps.example` | Env recomendado foxpro |
| `easypanel.env.fase-c.processor-02.example` | 2º processor (VPS ou dedicado) |
| `scripts/phase-c-verify.sh` | Checklist automatizado |
| `scripts/c1-ab-baseline.sh` | Amostras C1 |
| `deploy/observability/*` | Prometheus scrape + alertas |
| `deploy/phase-c/redis-compose.example.yml` | Redis preparatório C6 |
| `sql/phase_c_worker_split.example.sql` | Repartir `worker_id` |
| `docs/FASE_C_DEDICADO_PENDENTE.md` | Lista pós-VPS |

Fase 1 encerrada: ver handoff em conversa / `PHASE1_CLOSURE` (opcional no repo).
