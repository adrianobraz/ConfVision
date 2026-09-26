# Verificação — foxpro / confvision-worker (EasyPanel)

Data referência: **2026-09-26**

## Sonda externa (HTTP)

| URL | Resultado | Interpretação |
|-----|-----------|----------------|
| `https://foxpro-rust-pilot.rkr351.easypanel.host/health` | **200** ok | Rust pilot operacional |
| `https://foxpro-confvision.rkr351.easypanel.host/` | **200** | MediaMTX + Guard (`confvision`) up |
| `https://foxpro-confvision-worker.rkr351.easypanel.host/` | **502** | Proxy existe; **container worker não está rodando** |

Erro típico no painel: **`No such image: easypanel/foxpro/confvision-worker:latest`** → imagem **nunca buildada** ou **removida**; serviço **amarelo**.

---

## Configuração correta (EasyPanel)

Projeto: **foxpro** · App: **confvision-worker**

| Campo | Valor esperado |
|-------|----------------|
| **Source** | GitHub **`adrianobraz/ConfVision`** |
| **Branch** | **`main`** (ou `rust-pilot` **só** se a raiz tiver `main.py` + `Dockerfile`) |
| **Build context / root** | **`/`** (raiz do repo — **não** `confvision-rust-processor/`) |
| **Dockerfile** | **`Dockerfile`** (YOLO worker — **não** `Dockerfile-mediamtx`) |
| **Comando / Arguments** | vazio (CMD: `python -u main.py`) |
| **Rede** | mesma rede Docker dos apps `foxpro_confvision` / rust-pilot |

Env VPS sem GPU: copiar [`worker.env.vps-sem-gpu.example`](./worker.env.vps-sem-gpu.example) (API **`https://vision.confmonit2.com.br`**, não Xano).

| Crítico | Exemplo |
|---------|---------|
| `WORKER_ID` | `worker-docker-21` (= Postgres `vis_camera.worker_id` das cams Python) |
| `MEDIAMTX_RTSP_BASE` | `rtsp://foxpro_confvision:8554` |
| `RTMP_PUBLISH_SECRET` | igual Go / MediaMTX |
| `VIS_WORKER_API_KEY` | igual Go / Rust |

Detalhe completo: [`../DEPLOY_EASYPANEL.md`](../DEPLOY_EASYPANEL.md) § confvision-worker.

---

## Corrigir *No such image*

### Opção 1 — EasyPanel (recomendado)

1. Abrir **confvision-worker** → **Implantar / Deploy** (build), **não** só Start.
2. Log de build deve terminar com tag tipo `easypanel/foxpro/confvision-worker:latest`.
3. **Start** → logs `[START] ConfVision worker`.

### Opção 2 — SSH na VPS foxpro (31.97.173.119 ou host EasyPanel)

```bash
docker images | grep -E 'confvision-worker|foxpro/confvision-worker'

# Se vazio — build na pasta do código (clone ConfVision main, raiz com main.py):
git clone -b main https://github.com/adrianobraz/ConfVision.git /opt/confvision-build
cd /opt/confvision-build
docker build -t confvision-worker:1 .
docker tag confvision-worker:1 easypanel/foxpro/confvision-worker:latest

# EasyPanel → Start confvision-worker
```

Script auxiliar no repo: [`../confvision-rust-processor/scripts/foxpro-fix-worker-image.sh`](../confvision-rust-processor/scripts/foxpro-fix-worker-image.sh)

---

## Checklist pós-fix

- [ ] `docker ps` mostra container **confvision-worker** running
- [ ] Logs: `[START]`, sync câmeras, sem loop de crash
- [ ] URL `foxpro-confvision-worker…` deixa de ser 502 (se expuser HTTP) ou ignore se app sem porta pública
- [ ] Postgres: câmeras com `worker_id` = `WORKER_ID` do env entram no sync Python
- [ ] C1 A/B: 10 cams Python + 10 Rust possível

---

## Erros comuns

| Sintoma | Causa |
|---------|--------|
| No such image `:latest` | Start sem Deploy; ou prune de imagens |
| Build Success só em **docs** | Deploy acionado no app errado ou branch sem `Dockerfile` na raiz |
| Dockerfile-mediamtx no worker | App errado — worker usa **`Dockerfile`** |
| `live/{id}` no log RTSP | Falta `RTMP_PUBLISH_SECRET` ou secret ≠ Go |

---

## Relacionado

- Monitor C2: CT 111 — `docs/C2_MONITOR_INTEGRACAO_EDGE.md` (a criar no repo)
- Piloto Rust: `confvision-rust-processor/DEPLOY_EASYPANEL.md`
