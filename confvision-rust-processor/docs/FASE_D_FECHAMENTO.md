# Fase D — fechamento D5 + D6 (produção oficial foxpro)

**Data:** 2026-09-28  
**Decisão de produto:** os apps EasyPanel ainda usam o sufixo *rust-pilot* / *piloto* no nome, mas **para operação e clientes isto é produção oficial** — stack **Rust analítico + Go/Postgres**.

**`confvision-worker` (Python):** **desligado permanentemente** — não sobe mais para ingest/analítico. Controle e processamento = **Rust A/B + Go D5**.

**Control plane:** `https://vision.confmonit2.com.br` (ConfVision Go + D5 deployado).

---

## Blocos encerrados (operacional)

| Bloco | Status | Evidência |
|-------|--------|-----------|
| **D5.1 / D5.2** | **Fechado** | `GET /vis_rust_processor_capacity` → **200**; env `RUST_PROCESSOR_BASE_URLS` + `D5_AUTO_ASSIGN_ENABLED=1` no Proxmox |
| **D6.1** | **Fechado** | CT111 `phase-d6-verify.sh` → **OK D6** (A+B+sidecar+Go); cron `/etc/cron.d/confvision-foxpro-d6` |
| **D5.3** rebalance automático | Fora do fechamento | Assign manual / API |
| **D6.2** alertas Prometheus | Opcional | Regras no repo; ligar na infra quando quiser |

---

## Nomenclatura (não confundir)

| Nome legado | Significado hoje |
|-------------|------------------|
| `foxpro-rust-pilot` / `-b` | **Processors analíticos produção** (A/B) |
| `rust-processor-pilot-a-01` etc. | **`worker_id` real** no Postgres / sync |
| “Piloto” nos docs | Histórico Fase C/D — **não** implica ambiente de teste |

---

## O que continua aberto (não bloqueia D5/D6)

- **D3 aceite:** ≥1 evento E2E `vis_evento` ( `events_published` > 0 ) — melhoria contínua.
- **Capacidade RAM** na VPS foxpro — monitorar `capacity_state` / D5 `assign_eligible`.

---

## Colocar ConfVision no ar (clientes voltarem a testar)

Quando “**tudo parado**”, normalmente faltam **MediaMTX**, **processors Rust** e/ou **sidecar YOLO**. O **Go central** pode estar OK (`/vis_health`) enquanto a **foxpro** está down. **Não** usar `confvision-worker`.

### Camada 1 — Central (clientes: painel, API, cadastro)

| Onde | Ação |
|------|------|
| **Proxmox** | `systemctl status confmonit4confvision` → active |
| Teste | `curl -sS https://vision.confmonit2.com.br/vis_health` → `"status":"ok"` |
| Teste D5 | `curl -sS -H "Authorization: Bearer $VIS_WORKER_API_KEY" …/vis_rust_processor_capacity` → **200** |

Sem isto: clientes não sincronizam câmeras/licenças via Go.

### Camada 2 — Foxpro EasyPanel (vídeo + analítico)

**Ordem sugerida de Start** (sem worker Python):

1. **`confvision`** (MediaMTX) — RTSP/HLS internos quando houver publisher no path `cam/{hash}`.
2. **`foxpro/rust-yolo-sidecar`** — Start (porta **8091**).
3. **`foxpro-rust-pilot`** e **`foxpro-rust-pilot-b`** — Start (HTTP **8090** cada).

**`confvision-worker`:** manter **Stop** no EasyPanel (decisão definitiva).

Serviços auxiliares só se ainda fizerem sentido na operação (ex.: `confvision-sync-agent`); **não** dependem do worker Python.

**Câmeras / RTSP (sem worker):**

- **Produção recomendada:** [Opção A — RTSP direto](./OPCAO_A_RTSP_DIRETO.md) (`rtsp_url_sec` + despausar analítico). Scripts: `scripts/apply-rtsp-option-a.ps1`, `opcao-a-verify.sh`.
- Rust consome **RTSP** via sync Go (`rtsp_url_sec` direto na câmera **ou** `rtsp://…/cam/{hash}` no MediaMTX).
- Path `cam/{hash}` **sem publisher RTMP** → RTSP 404 no Rust até existir ingest no Rust/MediaMTX (roadmap) ou **`rtsp_url_sec`** apontando para a fonte real da câmera.
- **`worker_id`** analítico = `rust-processor-pilot-a-01` / `rust-processor-pilot-b-02` (ou D5 assign). **Nunca** voltar `worker_id` para o Python.

### Camada 3 — Postgres (câmeras no processor certo)

```sql
SELECT id, nome, ativo, worker_id, deteccao_humano, analitico_pausado
FROM vis_camera
WHERE ativo = true
ORDER BY id;
```

- Câmeras analíticas devem apontar para um **`worker_id`** Rust ativo.
- Assign: `POST /vis_camera_assign_processor` ou `POST /ops/d5/auto_assign_analiticas` (`dry_run: true` primeiro).

### Verificação rápida (qualquer host)

```bash
bash confvision-rust-processor/scripts/phase-d6-verify.sh
node confvision-rust-processor/scripts/d3-online-test.mjs
```

CT111 (cron): `tail -50 /var/log/foxpro-d6-verify.log`

---

## Rollback / legado

**Não** reativar `confvision-worker` como política. Em falha Rust: corrigir env/rede/RAM, rebalance D5, ou pausar analítico na câmera — ver [RECUPERACAO_CAMERAS.md](./RECUPERACAO_CAMERAS.md) (Caminho B / runbooks). Caminho A (Python) fica **só referência histórica**.

---

## Implantar fases 0–4 (todas câmeras)

Roteiro único + script: [IMPLANTAR_FASES_0_4.md](../../deploy/tenant-stack/docs/IMPLANTAR_FASES_0_4.md) · `scripts/implantar-fases.ps1`

---

## Referências

- [FASE_D.md](./FASE_D.md) — visão geral Fase D  
- [FASE_D_D5_KICKOFF.md](./FASE_D_D5_KICKOFF.md) · [FASE_D_D6_KICKOFF.md](./FASE_D_D6_KICKOFF.md)  
- [C2_MONITOR_INTEGRACAO_EDGE.md](./C2_MONITOR_INTEGRACAO_EDGE.md)  
- [RECUPERACAO_CAMERAS.md](./RECUPERACAO_CAMERAS.md)
