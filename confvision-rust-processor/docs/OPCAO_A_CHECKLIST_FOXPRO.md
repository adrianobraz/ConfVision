# Opção A — checklist foxpro (RTSP direto, sem worker Python)

**Atualizado:** 2026-09-28  
**Branch:** `rust-pilot`  
**Contexto:** worker Python **descontinuado**; ingest = `rtsp_url_sec` no Postgres via API Go.

---

## Já feito (repo + operação)

| Item | Status | Referência |
|------|--------|------------|
| Sidecar YOLO build (pin ultralytics 8.3 + tqdm) | OK | commits `9a0e851`, `aa1ee43` |
| `foxpro-stack-verify` **não FAIL** por worker parado | OK | commit `b59fa32` — SKIP worker |
| Scripts Opção A | OK | `opcao-a-preflight.ps1`, `opcao-a-run.ps1`, `apply-rtsp-option-a.ps1` |
| CSV modelo 6 câmeras | OK | `sql/opcao_a_foxpro_6_cameras.csv` (placeholders) |
| Guia onde rodar testes | OK | `ONDE_RODAR_TESTES_FOXPRO.md` |
| Infra smoke (D3, health, confvision guard) | OK | sidecar + pilots + MediaMTX HTTP 200 |
| **Despausar analítico** IDs 5,8,9,18,19,22,26,27 | Feito via API (sessão 2026-09-28) | ver seção “Estado cadastro” |
| EasyPanel: deletar app `confvision-worker` | Pendente operador | política: não subir worker |

---

## Estado cadastro (após despausar — conferir de novo)

Sync Go (`GET /vis_camera_sync_ativas?worker_id=…` + `X-Vis-Worker-Key`):

| Worker | Câmeras típicas no sync | Observação |
|--------|-------------------------|------------|
| `rust-processor-pilot-a-01` | 9, 18, 22, 26 (+ testes) | |
| `rust-processor-pilot-b-02` | 5, 8, 19, 27 (+ testes) | |

**Rust `/health` (referência pós-despausar):** até **4 total / 2 online** por pilot — RTSP ainda via `cam/{hash}` → **404** → política pode **repausar** (ex.: ids **8**, **19**).

**Bloqueio real Opção A:** `rtsp_url_sec` **vazio** em todas; `protocolo=dvr/wifi` no cadastro **não gera** URL automaticamente.

---

## Correções pendentes (ordem)

### 1. Dados para subir — `rtsp_url_sec` (obrigatório)

Para **cada** câmera que deve analisar:

1. Obter URL RTSP **real** (DVR/NVR/câmera), preferir **substream** se RAM apertada.
2. URL deve ser alcançável **da VPS foxpro** (teste: `ffprobe` de dentro do container rust-pilot ou rede foxpro).
3. Regras: `rtsp://` ou `rtsps://`; **sem** `/live/` no path.

**Arquivo:** editar `sql/opcao_a_foxpro_6_cameras.csv` (ou cópia local **fora do git** com senhas):

```text
id,rtsp_url_sec,despausar_analitico
5,rtsp://...,true
...
```

Incluir **8 e 9** se forem câmeras de teste permanentes. Para **18 ativas** no D5, repetir lógica para todos os IDs assignados A/B (ver `CAMERAS_AUDIT_2026-09-28.md`).

**Aplicar:**

```powershell
cd C:\sistemaconfmonit\core4-rust-pilot
$env:VIS_WORKER_API_KEY = "<chave>"   # rotacionar se vazou no chat
.\confvision-rust-processor\scripts\opcao-a-preflight.ps1 -VisWorkerKey $env:VIS_WORKER_API_KEY
.\confvision-rust-processor\scripts\opcao-a-run.ps1 -VisWorkerKey $env:VIS_WORKER_API_KEY -Apply
```

Alternativa: painel ConfVision → campo RTSP secundário por câmera.

### 2. Despausar analítico (de novo se necessário)

Se `analitico_pausado=true` após 404:

- API: `POST /vis_camera/analitico/pausar/{id}` body `{"pausado":false}`
- Ou coluna `despausar_analitico=true` no CSV do `apply-rtsp-option-a`

SQL exemplo: `sql/phase_d_despausar_analitico_exemplo.sql` (revisar IDs, `COMMIT` manual).

### 3. Verificar sync + Rust

```bash
bash confvision-rust-processor/scripts/opcao-a-verify.sh
bash confvision-rust-processor/scripts/phase-f5-health-metrics.sh   # CT111, jq
```

Esperado: `sync` com `rtsp_sec=sim`; logs Rust `worker started … url=rtsp://…` **sem** `foxpro_confvision:8554/cam/`; `/health` **`cameras_online`** subindo (~6+ conforme cadastro).

### 4. EasyPanel / rede (não Opção A, mas impacta)

| Item | Ação |
|------|------|
| App **confvision** (MediaMTX) | Manter **Running** (`MEDIAMTX_RTSP_BASE` interno `rtsp://foxpro_confvision:8554`) |
| **confvision-worker** | **Deletar** app (não Start) |
| Redis `No route to host` nos logs | Conferir serviço Redis foxpro na mesma rede Docker que pilots |
| Sidecar YOLO | Running; `d3-online-test.mjs` OK |

### 5. Segurança

- Rotacionar **`VIS_WORKER_API_KEY`** se exposta.
- Não commitar CSV com senhas RTSP — usar `.local.csv` gitignored ou painel.

### 6. Aceite produto (após vídeo)

- `/health`: `events_published` > 0 com movimento/pessoa
- `d3-event-verify.sh` / `vis_evento` com linhas novas
- Fase 5: rampa só com **online** estável

---

## Referências

- [OPCAO_A_RTSP_DIRETO.md](./OPCAO_A_RTSP_DIRETO.md)
- [CAMERAS_AUDIT_2026-09-28.md](./CAMERAS_AUDIT_2026-09-28.md)
- [RUNBOOK_CAMERAS_404_ATIVO.md](./RUNBOOK_CAMERAS_404_ATIVO.md)
- [ONDE_RODAR_TESTES_FOXPRO.md](./ONDE_RODAR_TESTES_FOXPRO.md)
