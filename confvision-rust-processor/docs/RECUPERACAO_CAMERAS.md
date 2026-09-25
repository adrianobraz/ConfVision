# Recuperação de câmeras — piloto Rust + stack ConfVision

Guia operacional **sem alterar Xano**. Tudo abaixo é EasyPanel, Postgres (`vis_camera`) e env dos serviços existentes.

**Checklist Fase 0 / provisionamento tenant (Proxmox):** [`../../deploy/tenant-stack/README.md`](../../deploy/tenant-stack/README.md).

---

## Sintomas comuns

| Sintoma | Causa provável |
|--------|----------------|
| `/health` e `/metrics` do Rust não respondem | Serviço **`rust-pilot` parado** (normal: HTTP só existe com container rodando) |
| `Name or service not known` em `foxpro_confvision` | **`confvision` (MediaMTX) parado** ou Rust **fora da rede Docker** foxpro |
| RTSP **404** no DESCRIBE | Path `cam/{hash}` **sem publisher RTMP** (worker parado ou `RTMP_PUBLISH_SECRET` diferente) |
| Câmera “sumiu” do Python | `worker_id` no Postgres aponta para **`rust-processor-pilot-01`** e Rust está off |
| CPU alta no Rust com câmera offline | Reconnect + métricas em janela; **infra RTSP morta** — corrigir stack antes de otimizar código |

---

## O que cada serviço faz

```text
Câmera ──RTMP──► MediaMTX (confvision / foxpro_confvision:8554)
                      │
                      └── RTSP cam/{hash} ──► confvision-worker (Python) OU rust-pilot (Rust)
```

- **`confvision`**: MediaMTX — **obrigatório** para RTSP interno.
- **`confvision-worker`**: publica streams no MediaMTX — **obrigatório** para paths `cam/...` (piloto atual).
- **`rust-pilot`**: só **consome** RTSP + decode/motion para câmeras cujo `worker_id` = `PROCESSOR_ID` do Rust.
- **Desligar todos os `confvision-*`**: derruba **todas** as câmeras, independente do Rust.

---

## Caminho A — Restaurar operação Python (rápido)

Use quando a prioridade é **voltar produção**, não testar Rust.

1. EasyPanel → projeto **foxpro** → **Start** nesta ordem:
   - `confvision`
   - `confvision-worker`
   - Demais conforme operação: `confvision-motion`, `confvision-sync-agent`, `confvision-dvr`, etc.

2. Postgres — listar câmeras ativas e `worker_id`:

   ```sql
   SELECT id, nome, worker_id, ativo
   FROM vis_camera
   WHERE ativo = true
   ORDER BY id;
   ```

3. Reverter câmeras de teste que foram para o piloto Rust (substituir pelo `WORKER_ID` do worker Python na VPS):

   ```sql
   -- Anotar valor anterior antes de mudar
   SELECT id, worker_id FROM vis_camera WHERE id = <ID_CAMERA>;

   UPDATE vis_camera
   SET worker_id = '<WORKER_ID_PYTHON_DA_VPS>'
   WHERE id = <ID_CAMERA>;
   ```

4. Aguardar 1–2 ciclos de sync (ex.: 60 s) ou reiniciar **`confvision-worker`**.

5. Manter **`rust-pilot` parado** até validar câmeras no painel/logs.

6. Baseline de CPU do host (~5–10%) com Rust off é **esperado**; não indica falha.

---

## Caminho B — Piloto Rust com 1 câmera

1. **Caminho A** estável (`confvision` + worker rodando).

2. Confirmar rede: container **`rust-pilot`** na **mesma rede** que MediaMTX.  
   `MEDIAMTX_RTSP_BASE` = hostname interno EasyPanel (ex.: `rtsp://foxpro_confvision:8554`).

3. Teste opcional (container debug na mesma rede):

   ```bash
   ffprobe -rtsp_transport tcp -i "rtsp://foxpro_confvision:8554/cam/<HASH12>" -show_streams -v error
   ```

4. **Uma** câmera de teste:

   ```sql
   UPDATE vis_camera
   SET worker_id = 'rust-processor-pilot-01'
   WHERE id = <ID_CAMERA_TESTE>;
   ```

   `PROCESSOR_ID` e `WORKER_ID` no env do Rust devem ser **`rust-processor-pilot-01`**.

5. Environment Rust (copiar segredos do worker, **mesmos valores**):
   - `CONFVISION_API_URL`, `VIS_WORKER_API_KEY`, `RTMP_PUBLISH_SECRET`, `MEDIAMTX_RTSP_BASE`
   - `MAX_CAMERAS=1`, `MEDIAMTX_NODE_ID=1` (se as câmeras usam nó 1 na API)
   - Ver [`easypanel.env.example`](../easypanel.env.example)

6. **Start** `rust-pilot` → validar:
   - `GET /health` → `cameras_online: 1`, `fps_total > 0`
   - Logs: `rtsp connected`, `frame received`

7. **Não desligar** `confvision` / worker durante o teste Rust.

---

## Rollback só do piloto Rust

1. Stop `rust-pilot`.
2. SQL na câmera piloto (restaurar `worker_id` Python).
3. Start/sync worker Python.

Detalhes: [DEPLOY_EASYPANEL.md §15](../DEPLOY_EASYPANEL.md).

---

## Checklist env foxpro (referência)

| Variável | Observação |
|----------|------------|
| `WORKER_ID` | Igual a `PROCESSOR_ID` no piloto (`rust-processor-pilot-01`) |
| `MEDIAMTX_NODE_ID` | Alinhar com nó da câmera (foxpro costuma usar `1`) |
| `RTMP_PUBLISH_SECRET` | **Byte a byte** igual ao `confvision-worker` |
| `MEDIAMTX_API_*` | Rust **não usa**; pode omitir no rust-pilot |

---

## Health / metrics

| Endpoint | Requisito |
|----------|-----------|
| `https://<dominio-rust-pilot>/health` | Serviço **rust-pilot running** |
| `/metrics`, `/ready` | Idem |

Com Rust desligado, Traefik/domínio não tem backend — **não é bug de config**.

---

## Próximo deploy Rust (build)

O `Dockerfile` usa **cargo-chef** + cache BuildKit para evitar compilação duplicada.  
Revisar diff local antes de deploy; não usar `Dockerfile.profiling` em produção.
