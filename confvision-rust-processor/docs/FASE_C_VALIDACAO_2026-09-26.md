# Validação Fase C — 2026-09-26

URL piloto: `https://foxpro-rust-pilot.rkr351.easypanel.host`  
API Go: `https://vision.confmonit2.com.br/vis_health` → **ok**

Última leitura remota: **2026-09-26 ~18:22 UTC** (pós-C3, env duplicata corrigida).

## Checklist PASS (`FASE_C.md`)

| Item | Resultado | Evidência |
|------|-----------|-----------|
| `/health` ok | **OK** | `status=ok`, 8/8 online |
| `rtsp_404_count=0` | **OK** | `summary.rtsp_404_count=0` |
| C3 admission ON | **OK** | `load_admission_enabled=true`, `load_policy_mode=admission`, `admission_active=true` |
| C2 monitoramento cron/Prometheus | **PENDENTE** | Ver [Passo 1](#passo-1--c2-monitoramento) — exemplo pronto em `deploy/observability/cron-rust-pilot-verify.example` |
| C1 planilha A/B | **PARCIAL** | Rust: `docs/c1-ab-results-2026-09-26.md` — falta janela **Python** (EasyPanel) |
| C4 2º processor | **ADIADO** | `capacity_state=healthy`, headroom ~2 cams; C4 se voltar **critical** com ~8 cams |
| Doc dedicado lido | **PENDENTE** | `FASE_C_DEDICADO_PENDENTE.md` (processo) |

## Snapshot operacional (pós-C3)

| Métrica | Valor |
|---------|--------|
| `cameras_total` / online | 8 / 8 |
| `capacity_state` | **healthy** (~80% capacity used) |
| `estimated_capacity_cameras` / available | 10 / 2 |
| `max_cameras` (env) | **30** (teto hard; doc recomenda 10) |
| `load_admission_enabled` | **true** |
| CPU container | ~50–51% |
| `fps_total` | ~58–79 |
| IDs ativos | 5, 18, 19, 20, 21, 22, 26, 27 |

**Comportamento C3:** em **critical**, esperar `allow_new_camera=false`. Em **healthy**, `allow_new_camera=true` é normal.

## Passos 1–4 (fechamento Fase C)

### Passo 1 — C2 monitoramento

No **core-4** (ou host com curl + bash + **jq**):

1. Clonar/sincronizar repo `rust-pilot` em path estável (ex. `/home/confmonit/v4.0/confvision-rust-processor/`).
2. Copiar `deploy/observability/cron-rust-pilot-verify.example` → `/etc/cron.d/confvision-rust-pilot`.
3. Ajustar URL se necessário; `chmod 644` no cron.d.
4. Confirmar log: `tail -f /var/log/rust-pilot-verify.log`.

Alternativa: **Uptime Kuma** → GET `/health`, keyword `"status":"ok"`.

### Passo 2 — C1 A/B

1. Escolher 1 câmera (ex. **id 5**).
2. Janela **Python** (`worker_id` legado): anotar CPU/RAM EasyPanel.
3. Mesma câmera no **Rust** — amostras já em `c1-ab-results-2026-09-26.md` ou rodar:
   ```bash
   bash scripts/c1-ab-baseline.sh https://foxpro-rust-pilot.rkr351.easypanel.host 60 10
   ```
4. Marcar decisão no final do arquivo C1.

### Passo 3 — C4 (condicional)

**Não executar agora** — piloto **healthy** com admission ON.

Reavaliar **C4** (`split_second_processor`) se:

- `capacity_state=critical` estável com as **8** cams atuais, ou
- `recommended_actions` voltar a listar `split_second_processor` com `estimated_available_cameras=0`.

Guia: `FASE_C_IMPLANTACAO_EASYPANEL.md` (seção C4), SQL `phase_c_foxpro_split_pilot02.sql`.

### Passo 4 — Verificação automatizada

```bash
bash scripts/phase-c-run-all.sh https://foxpro-rust-pilot.rkr351.easypanel.host
bash scripts/phase-c-verify.sh --strict-c3 https://foxpro-rust-pilot.rkr351.easypanel.host
```

Requer **jq** instalado (core-4 ou WSL). Script corrige aspas/Unicode (commit ops 2026-09-26).

## Histórico

| Momento | C3 | Notas |
|---------|-----|--------|
| Manhã 26/09 | FAIL | `load_admission_enabled=false`, critical |
| Tarde 26/09 | OK | Duplicata `LOAD_ADMISSION_ENABLED` removida no EasyPanel; redeploy |

## Comandos úteis

```bash
bash scripts/phase-c-run-all.sh https://foxpro-rust-pilot.rkr351.easypanel.host
bash scripts/c1-ab-baseline.sh https://foxpro-rust-pilot.rkr351.easypanel.host 60 10
curl -fsS "$URL/capacity-report" | jq '.load, .summary'
```
