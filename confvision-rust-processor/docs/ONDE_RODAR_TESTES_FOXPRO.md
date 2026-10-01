# Onde rodar cada teste (foxpro / CT111 / central)

Os scripts **não rodam dentro do EasyPanel**. Eles rodam **na sua máquina** ou no **CT111**, chamando as URLs **públicas** do foxpro (HTTPS).

---

## Mapa rápido

| Ambiente | O que é | Terminal | Repo no disco |
|----------|---------|----------|----------------|
| **A — Windows (dev / Cursor)** | Seu PC | **PowerShell** ou **Git Bash** | `C:\sistemaconfmonit\core4-rust-pilot` (clone ConfVision `rust-pilot`) |
| **B — CT111 ReceptorTeste** | LXC monitoramento | **SSH** → bash | `/opt/confvision/ConfVision` |
| **C — Proxmox (central Go)** | API visão | SSH ou curl de A/B | Não precisa do repo (só curl + API key) |
| **D — VPS foxpro EasyPanel** | Containers A/B/YOLO/MediaMTX | **Painel EasyPanel** (Logs, Start) — **não** `cd` no repo | Código vem do Git no build; sem shell interativo típico |

**Regra:** smoke D3/Fase 5 = ambiente **A** ou **B** (equivalente). Só muda o caminho do `cd`.

---

## A — Windows (PowerShell) — recomendado agora

Abra **PowerShell** (não precisa estar no foxpro).

```powershell
cd C:\sistemaconfmonit\core4-rust-pilot

# Sonda rápida Go + rust A + confvision + worker
powershell -File confvision-rust-processor\scripts\foxpro-stack-verify.ps1

# D3: sidecar + pilots (precisa Node instalado)
node confvision-rust-processor\scripts\d3-online-test.mjs
```

Pilot **B** (opcional):

```powershell
powershell -File confvision-rust-processor\scripts\foxpro-stack-verify.ps1 `
  -RustBase "https://foxpro-rust-pilot-b.rkr351.easypanel.host"
```

Central **D5** (substitua a chave; não commitar):

```powershell
$env:VIS_WORKER_API_KEY = "sua-chave"
curl.exe -fsS -H "Authorization: Bearer $env:VIS_WORKER_API_KEY" `
  "https://vision.confmonit2.com.br/vis_rust_processor_capacity"
```

---

## A — Windows (Git Bash) — scripts `.sh` Fase 5 / D6

Abra **Git Bash** (vem com Git for Windows):

```bash
cd /c/sistemaconfmonit/core4-rust-pilot

# D6 completo (sidecar + pilots + Go se VIS_WORKER_API_KEY exportada)
bash confvision-rust-processor/scripts/phase-d6-verify.sh

# Fase 5 endpoints HTTP
bash confvision-rust-processor/scripts/phase-f5-verify.sh

# Fase 5 resumo câmeras (precisa jq: apt no CT111 ou choco install jq no Windows)
bash confvision-rust-processor/scripts/phase-f5-health-metrics.sh
```

Se `phase-f5-verify` parar em `jq obrigatorio`, instale **jq** ou rode só a parte HTTP (primeiras linhas do script) / use CT111.

---

## B — CT111 (ReceptorTeste)

SSH no CT111, depois:

```bash
cd /opt/confvision/ConfVision
git pull origin rust-pilot

bash confvision-rust-processor/scripts/phase-d6-verify.sh
bash confvision-rust-processor/scripts/phase-f5-verify.sh
bash confvision-rust-processor/scripts/phase-f5-health-metrics.sh
```

CT111 **não** precisa de Rust/cargo — só curl, bash, jq, node (opcional).

---

## C — Proxmox / central (sem repo)

```bash
curl -fsS https://vision.confmonit2.com.br/vis_health
curl -fsS -H "Authorization: Bearer $VIS_WORKER_API_KEY" \
  https://vision.confmonit2.com.br/vis_rust_processor_capacity
```

Esperado pós-deploy: processors foxpro com `"reachable": true`.

---

## D — EasyPanel (foxpro) — quando o teste remoto falha

| Sintoma remoto | Onde olhar no EasyPanel |
|----------------|-------------------------|
| Sidecar **502** / HTML "Not Found" | App **`rust-yolo-sidecar`**: container **Running**, domínio apontando porta **8091**, rede interna `foxpro_rust-yolo-sidecar:8091` nos pilots |
| **confvision** **503** | App **MediaMTX/guard** — Start ou redeploy |
| **worker** **503** "not started" | **Esperado** — worker Python desligado de propósito |
| Pilots **200** /health | OK na borda; YOLO HTTP falha se sidecar 502 |

Pilots falam com sidecar pela **rede Docker** (`YOLO_HTTP_URL=http://foxpro_rust-yolo-sidecar:8091`), não pelo domínio público. Mesmo assim o domínio público do sidecar deve responder **400/200** em `/v1/detect` para validar TLS/proxy.

---

## URLs fixas (testes remotos)

| Serviço | URL |
|---------|-----|
| Rust A | `https://foxpro-rust-pilot.rkr351.easypanel.host` |
| Rust B | `https://foxpro-rust-pilot-b.rkr351.easypanel.host` |
| YOLO sidecar | `https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host` |
| confvision | `https://foxpro-confvision.rkr351.easypanel.host` |
| Go API | `https://vision.confmonit2.com.br` |

---

## Ordem sugerida (copiar e colar)

1. Windows PowerShell: `foxpro-stack-verify.ps1` + `d3-online-test.mjs`
2. Git Bash ou CT111: `phase-d6-verify.sh` → `phase-f5-verify.sh` → `phase-f5-health-metrics.sh`
3. Proxmox: `vis_rust_processor_capacity`
4. Rampa `phase-f5-ramp-run.sh` **só** com câmeras no sync Go

Ver também: [RUNBOOK_YOLO_SIDECAR.md](./RUNBOOK_YOLO_SIDECAR.md), [FASE_5.md](./FASE_5.md).

---

## Interpretar logs foxpro (2026-09-28)

### Sidecar: `ModuleNotFoundError: No module named 'tqdm'`

Container sobe a imagem com `ultralytics --no-deps` mas faltou pacote na lista → **rebuild** após atualizar `requirements-yolo-sidecar.txt` (commit com `tqdm`+). Até subir de novo, proxy público do sidecar tende **502**.

### Rust A/B: RTSP → `Name or service not known`

| Causa | Ação EasyPanel |
|--------|------------------|
| App **`rust-mediamtx`** (MediaMTX+Guard) **parado** | **Start** / redeploy |
| Hostname interno errado | `MEDIAMTX_RTSP_BASE=rtsp://foxpro_rust-mediamtx:8554` e `rtsp_url_sec` com o mesmo host — ver [FOXPRO_EASYPANEL_HOSTNAMES.md](./FOXPRO_EASYPANEL_HOSTNAMES.md) |
| Path `cam/{hash}` sem publisher | Worker Python off → usar **Opção A** `rtsp_url_sec` na câmera ([OPCAO_A_RTSP_DIRETO.md](./OPCAO_A_RTSP_DIRETO.md)) |

Pilots **OK** (redis, YOLO URL `http://foxpro_rust-yolo-sidecar:8091`) mesmo com RTSP quebrado — `/health` pode mostrar `cameras_total=1` e `cameras_online=0`.
