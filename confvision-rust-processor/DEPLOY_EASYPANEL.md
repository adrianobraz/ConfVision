# confvision-rust-processor — Piloto EasyPanel (1 câmera RTSP)

Deploy **mínimo** do processing plane Rust na VPS atual, como **serviço Docker separado** no EasyPanel.  
Sem alterar Go, Python, MediaMTX, Postgres ou Xano.

**Pré-requisitos:** Fase 1 commitada; imagem buildável pelo `Dockerfile` desta pasta; worker Python e MediaMTX já em produção.

Modelo de variáveis: [`easypanel.env.example`](./easypanel.env.example).

---

## 1. Criar o serviço Docker no EasyPanel

1. No projeto EasyPanel (ex.: **foxpro**), adicionar um **novo App** → tipo **Docker**.
2. Nome sugerido: `confvision-rust-processor` (ou `confvision-rust-pilot`).
3. **Source:** repositório oficial [https://github.com/adrianobraz/ConfVision](https://github.com/adrianobraz/ConfVision) (branch `main`); o serviço Rust fica na pasta **`confvision-rust-processor/`** na raiz do repo (mesmo fluxo dos outros serviços ConfVision).
4. **Root / build context:** diretório `confvision-rust-processor` (caminho relativo ao repo conforme layout no GitHub).
5. **Dockerfile path:** `Dockerfile` (na raiz dessa pasta).
6. **Comando / Arguments:** deixar vazio — a imagem usa `ENTRYPOINT` do binário.
7. **Restart policy:** always ou on-failure.

Não instalar Rust na VPS: o EasyPanel faz **build da imagem** a partir do Dockerfile (multi-stage).

---

## 2. Usar o Dockerfile existente

O `Dockerfile` atual:

- Stage **builder:** `rust:1.88-bookworm` → `cargo build --release`
- Stage **runtime:** `debian:bookworm-slim` + `ca-certificates` + binário em `/usr/local/bin/confvision-rust-processor`

Tag sugerida após build: `confvision-rust-processor:0.1.0` (ou `:latest` no registry interno do EasyPanel).

---

## 3. Porta interna 8090

| Campo | Valor |
|--------|--------|
| Porta do **container** | **8090** (`HTTP_PORT` default) |
| Publicação | Mapear para host/domínio interno conforme necessidade (ex.: `8091` no host se 8090 estiver ocupada) |

Variáveis na imagem: `HTTP_HOST=0.0.0.0`, `HTTP_PORT=8090`, `EXPOSE 8090`.

---

## 4. Variáveis de ambiente

1. Abra **Environment** do serviço no EasyPanel.
2. Cole/adapte o conteúdo de [`easypanel.env.example`](./easypanel.env.example).
3. Substitua **todos** os placeholders (`SUBSTITUA_*`) por valores reais **somente no painel** — não commitar segredos.

| Variável | Piloto |
|----------|--------|
| `CONFVISION_API_URL` | URL da **app Go** (HTTPS ou HTTP interno) |
| `VIS_WORKER_API_KEY` | Igual Go / worker Python |
| `RTMP_PUBLISH_SECRET` | Igual Go / MediaMTX |
| `MEDIAMTX_RTSP_BASE` | RTSP alcançável **de dentro** do container |
| `PROCESSOR_ID` | ID exclusivo (ex.: `rust-processor-pilot-01`) |
| `SYNC_FILTER_WORKER_ID` | `true` |
| `MAX_CAMERAS` | `1` |
| `WORKER_TIPO` | `rust_processor` |

**Não definir:** `RTSP_SIMULATE`, `XANO_BASE_URL`, `XANO_API_FRANQUEADO_PRO`, `REDIS_URL`, `S3_*`.

`CONFVISION_API_URL` aponta para a **API Go existente** (`/vis_camera_sync_ativas`, `/vis_worker_ping`), **não** para Xano. O binário recusa `*.xano.io` e variáveis `XANO_*`.

---

## 5. Health check `/ready`

No EasyPanel (Health check HTTP):

| Campo | Valor |
|--------|--------|
| Path | `/ready` |
| Porta | **8090** (porta do container) |
| Intervalo | 30 s |
| Start period | 60–120 s (primeiro sync + RTSP) |
| Sucesso | HTTP **200** e JSON `"ready": true` |

Alternativa informativa: `GET /health` (sempre 200 quando o processo está de pé; agrega contadores).

---

## 6. Rede com MediaMTX

O Rust abre RTSP **para o MediaMTX** (TCP, crate `retina`), não para a câmera diretamente (salvo `rtsp_url_sec` na câmera).

1. Coloque o serviço Rust na **mesma rede Docker** que o serviço **confvision** (MediaMTX), ou rota equivalente na VPS.
2. `MEDIAMTX_RTSP_BASE` deve usar o **hostname interno** EasyPanel do MediaMTX, por exemplo:
   - `rtsp://confvision:8554` ou o nome real do serviço na sua stack
3. Se a API Go enviar `mediamtx_rtsp_base` por câmera (join `vis_mediamtx_node`), esse valor **prevalece** sobre o env quando monta a URL.

Teste de conectividade (container debug na mesma rede):

```bash
# substituir host/porta/path
ffprobe -rtsp_transport tcp -i "rtsp://HOST:8554/cam/HASH_EXEMPLO" -show_streams -v error
```

Fluxo de mídia do piloto:

```text
Câmera ──RTMP──► MediaMTX (path cam/{hash}) ──RTSP TCP──► Rust (contagem de frames)
```

---

## 7. Validar comunicação com a API Go

Do host VPS ou de um container `curl` na mesma rede com saída HTTPS:

```bash
export CONFVISION_API_URL="https://SUA_APP_GO"
export VIS_WORKER_API_KEY="SUA_CHAVE"
export PROCESSOR_ID="rust-processor-pilot-01"

curl -sS -o /dev/null -w "HTTP %{http_code}\n" \
  -H "X-Vis-Worker-Key: ${VIS_WORKER_API_KEY}" \
  "${CONFVISION_API_URL}/vis_camera_sync_ativas?worker_id=${PROCESSOR_ID}"
```

Esperado: **HTTP 200** e JSON com `cameras` (array vazio **antes** de alterar `worker_id`, ou com **1** câmera após o UPDATE).

Ping (opcional):

```bash
curl -sS -X POST \
  -H "Content-Type: application/json" \
  -H "X-Vis-Worker-Key: ${VIS_WORKER_API_KEY}" \
  "${CONFVISION_API_URL}/vis_worker_ping" \
  -d '{"worker_id":"'"${PROCESSOR_ID}"'","worker_tipo":"rust_processor","hostname":"piloto","versao":"0.1.0","cameras_ativas":0,"ultimo_ping_em":"2020-01-01T00:00:00Z","ativo":true,"yolo_device":"none","queue_backend":"none"}'
```

Logs do Rust: ausência de `worker ping failed` / `camera sync failed` contínuos após subida.

---

## 8. Selecionar uma única câmera

Critérios:

1. Câmera **ativa**, `deteccao_humano = true`, analítico **não** pausado (regras já usadas pelo Go em `vis_camera_sync_ativas`).
2. Stream **já publicando** no MediaMTX em `cam/{hash12}` (RTMP online).
3. Câmera de **teste**, não crítica para operação até validar o piloto.

Anote: `id` da câmera, `worker_id` **atual** (Python), path/hash se conhecido.

---

## 9. Alterar somente o `worker_id` dessa câmera

**Sem migration e sem schema change** — apenas UPDATE na linha existente:

```sql
-- 1) Anotar o valor anterior
SELECT id, nome, worker_id FROM vis_camera WHERE id = <ID_CAMERA_PILOTO>;

-- 2) Atribuir ao processor Rust (igual PROCESSOR_ID no EasyPanel)
UPDATE vis_camera
SET worker_id = 'rust-processor-pilot-01'
WHERE id = <ID_CAMERA_PILOTO>;

-- 3) Confirmar
SELECT id, worker_id FROM vis_camera WHERE id = <ID_CAMERA_PILOTO>;
```

**Não** altere `worker_id` das demais câmeras.

Aguarde até um ciclo de `SYNC_INTERVAL_SEC` (60 s default) ou reinicie o container Rust uma vez.

---

## 10. Garantias do piloto

| Regra | Como garantir |
|--------|----------------|
| `SYNC_FILTER_WORKER_ID=true` | Env no EasyPanel + default do binário é `true` |
| `MAX_CAMERAS=1` | Env explícito |
| `PROCESSOR_ID` exclusivo | Diferente de `WORKER_ID` do Python; igual ao `worker_id` só da câmera piloto |
| `RTSP_SIMULATE` ausente | Não adicionar variável no Environment |

Defesa em profundidade: Go filtra por `worker_id` na query; o manager trunca se vierem mais de `MAX_CAMERAS` câmeras.

---

## 11. Logs — somente uma câmera iniciada

No EasyPanel → **Logs** do serviço, após sync:

Procure **uma** linha:

```text
worker started camera_id=<ID_PILOTO> ... url=rtsp://.../cam/...
```

Com `LOG_LEVEL=debug`, também:

```text
rtsp connected
frame received camera_id=... seq=...
```

**FAIL se:** vários `worker started` com `camera_id` diferentes, ou câmeras que não são a piloto.

Não deve aparecer `RTSP_SIMULATE` / loop simulado.

---

## 12. Verificar `/health` e `/metrics`

Substitua `HOST:PORTA` pelo mapeamento publicado (ou `exec` + curl na rede interna):

```bash
curl -sS "http://HOST:PORTA/ready" | jq .
curl -sS "http://HOST:PORTA/health" | jq .
curl -sS "http://HOST:PORTA/metrics" | jq .
```

Interpretação piloto:

| Endpoint | Esperado |
|----------|----------|
| `/ready` | `"ready": true` após API client OK |
| `/health` | `cameras_total` ≤ 1; `cameras_online` 1 quando RTSP estável; `frames_received` cresce |
| `/metrics` | `frames_received`, `reconnects`, `rtsp_errors`, `errors` |

---

## 13. CPU e RAM (EasyPanel / host)

O serviço **não** exporta CPU/RAM no JSON (Fase 1). Meça externamente:

- **EasyPanel:** gráficos de CPU/RAM do container `confvision-rust-processor`.
- **Host Linux:** `docker stats confvision-rust-processor` (nome do container pode variar).

Registre baseline durante 30–60 min com 1 câmera (para capacidade futura).

---

## 14. Critérios PASS / FAIL

### PASS (piloto mínimo, ≥ 30 min estáveis)

1. Container **running**; `/ready` → 200, `ready: true`.
2. Sync Go retorna **apenas** a câmera piloto (com `worker_id` correto).
3. Logs: **um** `worker started` para o `camera_id` piloto; RTSP conecta (debug ou frames crescendo).
4. `/metrics`: `frames_received` aumenta de forma contínua; `fps_total` > 0 na maior parte do tempo.
5. `reconnects` / `rtsp_errors` = 0 ou evento único explicado (reboot câmera).
6. Worker Python: **outras** câmeras seguem normais (smoke operacional).

### FAIL

- Container reinicia em loop ou exit por `config error`.
- Nenhum frame (`frames_received` parado > `RTSP_FRAME_TIMEOUT_SEC`).
- Mais de uma câmera processada.
- Python perdeu câmeras de produção por `worker_id` alterado em massa.
- RTSP inacessível do container (rede/`MEDIAMTX_RTSP_BASE` errado).

---

## 15. Rollback completo

1. **EasyPanel:** parar o serviço `confvision-rust-processor` (Scale 0 ou Stop).  
   O processo trata SIGTERM e envia ping final `ativo=false` quando possível.

2. **Postgres:** restaurar `worker_id` anterior **só** na câmera piloto:

```sql
UPDATE vis_camera
SET worker_id = '<WORKER_ID_PYTHON_ANTIGO>'
WHERE id = <ID_CAMERA_PILOTO>;
```

3. Aguardar 1–2 ciclos do sync do **worker Python** (`SYNC_INTERVAL_SEC`).

4. Confirmar que a câmera voltou a aparecer no sync do Python (mesmo `WORKER_ID` de antes).

5. Opcional: remover o app no EasyPanel ou manter parado; **não** é necessário alterar Go/MediaMTX.

6. Remover variáveis de ambiente sensíveis do serviço se o app for excluído.

---

## Referência rápida

| Item | Valor |
|------|--------|
| Binário | `/usr/local/bin/confvision-rust-processor` |
| HTTP | `:8090` `/health`, `/ready`, `/metrics` |
| Sync | `GET /vis_camera_sync_ativas?worker_id=...` |
| Ping | `POST /vis_worker_ping` |
| Isolamento | `worker_id` Postgres + `SYNC_FILTER_WORKER_ID` + `MAX_CAMERAS=1` |
