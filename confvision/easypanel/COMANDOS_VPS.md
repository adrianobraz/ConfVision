# Comandos no VPS EasyPanel (foxpro) — NÃO rodar no Proxmox core-4

MediaMTX, sync-agent e worker rodam no **VPS** (`vps.confhost.com.br`), não no `core-4`.

O `core-4` (Proxmox) só roda a **API Go** (`confmonit4confvision`).

---

## SSH no VPS EasyPanel

Entre no servidor onde o EasyPanel/Docker está instalado (não no core-4).

```bash
docker ps | grep confvision
```

---

## Testar env + código (MediaMTX)

Substitua `CONTAINER` pelo ID ou nome (ex.: `foxpro_confvision-1`):

```bash
docker exec -it CONTAINER sh -c '
  echo "ENV:"; env | grep VIS_WORKER
  python -c "
import os
print(\"KEY=\", repr(os.getenv(\"VIS_WORKER_API_KEY\")))
from vis_api_auth import vis_api_headers
print(\"HEADERS=\", vis_api_headers())
"
'
```

Esperado:

```text
VIS_WORKER_API_KEY=a7f3c9e2-8b1d-4f6a-9c0e-vis-bridge-2026
HEADERS= {'X-Vis-Worker-Key': 'a7f3c9e2-8b1d-4f6a-9c0e-vis-bridge-2026'}
```

---

## Testar sync-agent

```bash
docker ps | grep sync-agent
docker exec -it CONTAINER_SYNC env | grep VIS_WORKER
docker exec -it CONTAINER_SYNC python -c "from vis_api_auth import vis_api_headers; print(vis_api_headers())"
```

---

## Desbanir IP (Guard RTMP)

Dentro do **VPS**, use o nome Docker do serviço MediaMTX:

```bash
curl -s http://127.0.0.1:8100/bans \
  -H "X-RTMP-Guard-Key: 4XFhEtfY87lIH6Z5pWbJgVyausej9vBMcPikq3dU"

curl -X POST http://127.0.0.1:8100/unban \
  -H "Content-Type: application/json" \
  -H "X-RTMP-Guard-Key: 4XFhEtfY87lIH6Z5pWbJgVyausej9vBMcPikq3dU" \
  -d '{"ip":"189.110.7.137"}'

curl -X POST http://127.0.0.1:8100/unban \
  -H "Content-Type: application/json" \
  -H "X-RTMP-Guard-Key: 4XFhEtfY87lIH6Z5pWbJgVyausej9vBMcPikq3dU" \
  -d '{"ip":"179.162.106.154"}'
```

Se a porta 8100 não estiver exposta no host, entre no container:

```bash
docker exec -it CONTAINER_MEDIAMTX curl -s http://127.0.0.1:8100/bans \
  -H "X-RTMP-Guard-Key: 4XFhEtfY87lIH6Z5pWbJgVyausej9vBMcPikq3dU"
```

Ou edite o arquivo no volume:

```bash
echo '{}' > /opt/confvision/recordings/rtmp_bans.json
```

---

## Ordem de redeploy (EasyPanel UI)

1. GitHub `adrianobraz/ConfVision` — push do código com `vis_api_auth.py`
2. **confvision-sync-agent** — Environment `sync-agent.env` → Save → Deploy
3. **confvision** (MediaMTX) — Environment `mediamtx.env` → Save → Deploy
4. **confvision-worker** — Environment `worker.env` → Save → Deploy
5. Desbanir IPs (comandos acima)
6. Câmera `179.162.106.154`: trocar path RTMP de `live/1` para `cam/{hash12}`

---

## Proxmox core-4 (só API Go)

```bash
chmod +x /home/confmonit/v4.0/confvision/confvision
sudo systemctl restart confmonit4confvision
curl https://vision.confmonit2.com.br/vis_health
curl -H "X-Vis-Worker-Key: a7f3c9e2-8b1d-4f6a-9c0e-vis-bridge-2026" \
  https://vision.confmonit2.com.br/vis_camera/rtmp_auth/2
```
