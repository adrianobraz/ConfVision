# Implantar fases 0–4 (todas as câmeras analíticas)

**Política:** testar e operar **todas** as câmeras analíticas ativas — sem limite artificial de “1 câmera piloto”.  
**Worker Python:** permanece **Stop**; ingest = RTSP Opção A + MediaMTX.

---

## Comando único (Windows / operador)

```powershell
cd C:\sistemaconfmonit\core4-rust-pilot\confvision-rust-processor\scripts
.\implantar-fases.ps1 -DespausarAnalitico   # 1ª vez ou após pausa em massa
.\implantar-fases.ps1 -VerifyOnly             # só auditoria + probes
```

O script executa, em ordem:

| Fase | Ação automática |
|------|-----------------|
| **0** | Opção A `rtsp_url_sec` em **todas** (`opcao-a-apply-all.ps1`) |
| **0** | Auditoria RTSP + `worker_id` (`audit-analiticas.ps1`) |
| **D5** | Assign analíticas **sem** `worker_id` (dry-run + apply) |
| **0/D6** | `foxpro-stack-verify`, health Rust A+B |
| **3** | `d2-redis-verify.sh` (WARN se `QUEUE_BACKEND=none`) |
| **4** | `phase-d6-verify.sh` (bash / CT111) |
| **D3** | `d3-online-test.mjs --quick` |

Fases **1** (novo tenant) e **2** (GPU) exigem passos manuais no host — o script imprime o roteiro.

---

## Fase 0 — Produção foxpro (EasyPanel)

**Ordem Start:** MediaMTX `confvision` → `rust-yolo-sidecar` → `foxpro-rust-pilot` + `foxpro-rust-pilot-b`.

**Env processors (todas câmeras, CPU VPS):**

- Copiar [`easypanel.env.producao-todas-cameras.example`](../../../confvision-rust-processor/easypanel.env.producao-todas-cameras.example) para cada Rust A/B (ajustar `WORKER_ID` / `PROCESSOR_ID`).
- Sidecar: `YOLO_WORKERS=auto`, limites conservadores (RUNBOOK Fase D).

**Go central:**

- `D5_AUTO_ASSIGN_ENABLED=1`, `RUST_PROCESSOR_BASE_URLS` = URLs A+B.
- Deploy com auto `rtsp_url_sec` no cadastro.

**Critérios GO (conjunto inteiro):**

- `audit-analiticas.ps1` → 0 sem RTSP, 0 sem worker (ou assign D5).
- Rust A+B: `cameras_online` soma ≈ câmeras assignadas; `frames_received` sobe.
- `events_published` > 0 e registros em `vis_evento` (D3).
- CPU estável: se `reject_admission` persistente → subir `MAX_CAMERAS` split A/B ou perfil `ANALYSIS_LEGACY_VPS` + Redis fila.

Checklist detalhado: [FASE0_CHECKLIST_OPS.md](./FASE0_CHECKLIST_OPS.md).

---

## Fase 1 — Novo tenant `ct_cli_*`

1. `export TENANT_ID=ct_cli_<slug> MAX_CAMERAS=200`
2. `./scripts/tenant-init-env.sh`
3. `docker compose -f docker-compose.tenant.example.yml up -d`
4. Postgres: câmeras do cliente com `worker_id = TENANT_ID`

Doc: [FASE1_PROVISIONAMENTO_TENANT.md](./FASE1_PROVISIONAMENTO_TENANT.md).

---

## Fase 2 — GPU / NVDEC

- Host Proxmox com GPU; `VIDEO_ACCELERATION=auto`.
- Verificação: `bash confvision-rust-processor/scripts/phase-4.1-verify.sh`
- Benchmark ~200–400 câmeras: [FASE_4_1_FECHAMENTO.md](../../../confvision-rust-processor/docs/FASE_4_1_FECHAMENTO.md).

Foxpro atual **sem GPU** → Fase 2 fica **preparada no código**, ativa quando houver hardware.

---

## Fase 3 — Escala analítica (fila + admission)

1. EasyPanel: `QUEUE_BACKEND=redis` + `REDIS_URL` (mesmo Redis central) nos dois Rust.
2. `d2-redis-verify.sh` → `event_queue_redis_ok=true`.
3. `LOAD_ADMISSION_ENABLED=1`, load shedding conforme `easypanel.env.producao-todas-cameras.example`.

Doc: [FASE_D_D2_KICKOFF.md](../../../confvision-rust-processor/docs/FASE_D_D2_KICKOFF.md).

---

## Fase 4 — Automação / observabilidade

1. CT111 cron: `confvision-foxpro-d6` → `phase-d6-verify.sh`
2. Opcional: regras Prometheus D6.2
3. D5.3 rebalance automático: fora do fechamento; usar `implantar-fases.ps1` + capacity API quando adicionar câmeras

Doc: [FASE_D_FECHAMENTO.md](../../../confvision-rust-processor/docs/FASE_D_FECHAMENTO.md).

---

## Fase 5 — Capacidade medida (todas câmeras)

Rampa progressiva (não pular para 30 se CPU critical):

```bash
export F5_STAGES="5 10 20 30"
export F5_HOLD_SEC=300
bash confvision-rust-processor/scripts/phase-f5-ramp-run.sh
```

Doc: [FASE_5.md](../../../confvision-rust-processor/docs/FASE_5.md).

---

## Referência rápida

| Documento | Conteúdo |
|-----------|----------|
| [OPCAO_A_RTSP_DIRETO.md](../../../confvision-rust-processor/docs/OPCAO_A_RTSP_DIRETO.md) | RTSP sec + hash |
| [FASE_D_FECHAMENTO.md](../../../confvision-rust-processor/docs/FASE_D_FECHAMENTO.md) | Produção foxpro |
| [RECUPERACAO_CAMERAS.md](../../../confvision-rust-processor/docs/RECUPERACAO_CAMERAS.md) | Incidentes |
