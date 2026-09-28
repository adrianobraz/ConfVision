# Diagnóstico — confvision-worker não sobe (foxpro)

**Data:** 2026-09-26 (tarde)  
**Contexto:** Rust pilot OK; worker Python parado.

## Testes executados (externos)

| Alvo | Resultado | Evidência |
|------|-----------|-----------|
| `GET …/vis_health` (Go) | **200** | `postgres ok` |
| `GET foxpro-rust-pilot…/health` | **200** | 5/5 cams, `capacity_state=critical`, `load_advisory=reject_admission` |
| `GET foxpro-confvision…/` | **200** | RTMP guard JSON ok |
| `GET foxpro-confvision-worker…/` | **503** | HTML EasyPanel: **"Service is not started"** |

Scripts repetíveis:

```powershell
.\confvision-rust-processor\scripts\foxpro-stack-verify.ps1
```

```bash
bash confvision-rust-processor/scripts/foxpro-stack-verify.sh
```

## Conclusão principal

O worker **não está em crash loop visível pela internet**: o EasyPanel responde **503 "Service is not started"**, ou seja, o serviço **`confvision-worker` está parado no painel** (Stop / nunca deu Start após Deploy / Start falhou antes de manter container).

Isso é **independente** do Rust pilot (outro app, outra imagem, HTTP `/health`).

Deploy verde (*docs: worker.env…*) só **gera/atualiza imagem**; **não substitui Start**.

## O que NÃO é (com base nas sondas)

- Falha da API Go (`vis_health` ok).
- MediaMTX/guard down (`confvision` ok).
- Rust pilot down (health ok).

## Causas prováveis (ordem)

1. **Serviço parado** — CPU/mem 0% no painel; URL worker = "Service is not started".
2. **Start falhou** — ex. `No such image: easypanel/foxpro/confvision-worker:latest` → **Implantar** (build) + **Start**.
3. **Origem errada no EasyPanel** — Dockerfile/contexto `confvision-rust-processor` em vez da **raiz** + `Dockerfile` + `main.py` (GitHub `main` tem ambos).
4. **Após Start, crash** — só visível em **Logs** do app (env S3/RTMP/XANO, OOM YOLO). O `main.py` chama `validate_config()` mas **não encerra** só por ERRO de config; crash seria import/OOM/etc.

## Worker vs domínio HTTP

O `Dockerfile` do worker **não expõe porta HTTP**. Domínio `foxpro-confvision-worker…` pode continuar **502/503** mesmo com processo rodando. Critério de sucesso:

- Logs: `[START] ConfVision worker`
- CPU/mem > 0 estável no EasyPanel
- Ping/sync na API para `WORKER_ID` configurado

## Ações recomendadas (foxpro)

1. EasyPanel → **foxpro / confvision-worker** → aba **Origem**: repo `adrianobraz/ConfVision`, branch **`main`**, root **`/`**, Dockerfile **`Dockerfile`**, comando vazio.
2. **Implantar** (se imagem ausente ou após mudança de origem).
3. **Start** (play) — observar **Logs** nas primeiras 60s.
4. Se imagem faltar na VPS: `bash confvision-rust-processor/scripts/foxpro-fix-worker-image.sh main` (SSH) → Start no painel.
5. Env: `easypanel/worker.env.vps-sem-gpu.example` (`XANO_BASE_URL=https://vision.confmonit2.com.br`, `YOLO_ARCH=legacy` na foxpro).

## SSH / host

Não foi possível auditar Docker na VPS a partir deste ambiente (SSH host key / sem acesso). Na foxpro:

```bash
docker ps -a | grep -i confvision-worker
docker images | grep confvision-worker
docker logs $(docker ps -aq --filter name=confvision-worker | head -1) --tail 100
```

## Fase C

- **C1 A/B (10 Python + 10 Rust)** permanece bloqueado até worker **Start** + logs OK.
- Rust-only (C3/C2) continua testável via `phase-c-verify.sh`.

Referência: [`easypanel/VERIFICACAO_CONFVISION_WORKER_FOXPRO.md`](../../easypanel/VERIFICACAO_CONFVISION_WORKER_FOXPRO.md)
