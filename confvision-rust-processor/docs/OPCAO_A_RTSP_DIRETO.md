# Opção A — RTSP direto (`rtsp_url_sec`) sem worker Python

**Contexto:** com `confvision-worker` **desligado**, o path MediaMTX `rtsp://…/cam/{hash12}` só funciona se alguém publicar RTMP nesse path. Sem publisher → Rust vê **RTSP 404**.

**Opção A:** gravar na câmera a **URL RTSP real** do DVR/NVR/câmera em `vis_camera.rtsp_url_sec`. O processor Rust (e o legado Python, se existisse) usa essa URL quando:

- começa com `rtsp://` ou `rtsps://`, e
- **não** contém `/live/` (formato legado ignorado → cai no `cam/{hash}`).

Referência de código:

- Rust: `src/camera/stream_url.rs` (`resolve_rtsp_url`)
- Python: `confvision/urls.py` (`rtsp_url_for_camera`)

**Não há** descoberta automática de URL RTSP no repositório: a URL vem do cadastro (painel, SQL ou API).

---

## Pré-requisitos

1. EasyPanel foxpro: **MediaMTX** + **rust A/B** + **yolo-sidecar** — **sem** `confvision-worker`.
2. Go central OK: `GET https://vision.confmonit2.com.br/vis_health` → `"status":"ok"`.
3. Câmera **ativa**, `deteccao_humano = true`, **`analitico_pausado = false`** (senão não entra em `GET /vis_camera_sync_ativas`).
4. `worker_id` apontando para um processor Rust (`rust-processor-pilot-a-01` / `…-b-02` ou D5 assign).

---

## Passo 1 — Obter URL RTSP por câmera

Exemplos típicos (ajustar IP, canal, usuário/senha do cliente):

```text
rtsp://usuario:senha@192.168.1.10:554/Streaming/Channels/101
rtsp://usuario:senha@dvr.cliente.com:554/cam/realmonitor?channel=1&subtype=0
```

Regras:

- Use **substream** (subtype=1 / canal “secundário”) se a VPS foxpro tiver RAM limitada.
- Evite URLs com `/live/` no path — serão ignoradas.
- Credenciais ficam **só** no Postgres; logs Rust usam `redact_rtsp_url` (sem user/pass).

---

## Passo 2 — Aplicar cadastro

### A) Painel ConfVision

Editar câmera → campo RTSP secundário / URL RTSP → salvar. Despausar analítico se necessário.

### B) API Go (lote) — `VIS_WORKER_API_KEY`

CSV modelo: `sql/phase_d_rtsp_url_sec_template.csv`

```powershell
$env:CONFVISION_API_URL = "https://vision.confmonit2.com.br"
$env:VIS_WORKER_API_KEY = "<mesma chave dos processors Rust>"
.\confvision-rust-processor\scripts\apply-rtsp-option-a.ps1 -CsvPath .\meu_rtsp.csv -DryRun
.\confvision-rust-processor\scripts\apply-rtsp-option-a.ps1 -CsvPath .\meu_rtsp.csv
```

Linux (CT111 / operador):

```bash
export CONFVISION_API_URL=https://vision.confmonit2.com.br
export VIS_WORKER_API_KEY='...'
bash confvision-rust-processor/scripts/apply-rtsp-option-a.sh sql/phase_d_rtsp_url_sec_template.csv --dry-run
bash confvision-rust-processor/scripts/apply-rtsp-option-a.sh meu_rtsp.csv
```

Colunas CSV: `id`, `rtsp_url_sec`, opcional `despausar_analitico` (`true`/`false`).

### C) SQL direto (Postgres produção)

Ver `sql/phase_d_despausar_analitico_exemplo.sql` — **sempre** `BEGIN` + conferir + `COMMIT`.

Após UPDATE em massa, incrementar política de stream se o Go não fizer sozinho via API (PUT via API chama `bumpStreamPolicyGeneration`).

---

## Passo 3 — Verificar sync

```bash
export CONFVISION_API_URL=https://vision.confmonit2.com.br
export VIS_WORKER_API_KEY='...'
export WORKER_ID=rust-processor-pilot-a-01   # opcional
bash confvision-rust-processor/scripts/opcao-a-verify.sh
```

Esperado:

- `sync_ativas`: lista com câmeras de teste (não vazia se despausadas + RTSP preenchido).
- Health Rust (`:8090/health`): `cameras_online` > 0 para IDs assignados àquele processor.
- Métricas: `events_published` > 0 após movimento/humano (D3).

Consulta SQL útil:

```sql
SELECT id, nome, ativo, deteccao_humano, analitico_pausado, worker_id,
       LEFT(COALESCE(rtsp_url_sec, ''), 40) AS rtsp_preview,
       stream_erro_classe, ultimo_stream_ok_em
FROM vis_camera
WHERE ativo = TRUE
ORDER BY id;
```

---

## Troubleshooting

| Sintoma | Causa provável | Ação |
|--------|----------------|------|
| `sync_ativas` → `cameras: []` | Todas pausadas ou `deteccao_humano=false` | Despausar; conferir filtros em `ListCamerasAnaliticas` |
| RTSP 404 no Rust | `rtsp_url_sec` vazio → `cam/{hash}` sem publisher | Preencher Opção A |
| RTSP timeout | Firewall/NAT DVR | RTSP acessível **da VPS foxpro**, não só LAN cliente |
| Câmera no processor errado | `worker_id` | D5 assign ou `POST /vis_camera_assign_processor` |

---

## Referências

- [FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md)
- [REVISAO_PRIORIDADES_2026-09-28.md](./REVISAO_PRIORIDADES_2026-09-28.md)
