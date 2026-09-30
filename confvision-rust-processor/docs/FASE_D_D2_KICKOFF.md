# D2 — Fila Redis (eventos)

## Objetivo

- `QUEUE_BACKEND=redis` + `REDIS_URL` no rust-processor (mesmo Redis do worker: `185.130.61.5:6379`).
- Lista principal: `EVENT_QUEUE_KEY=confvision:eventos` (paridade `confvision/event_queue.py`).
- DLQ: `confvision:eventos:dlq` (ou `EVENT_QUEUE_DLQ_KEY`).
- Retry publish: `QUEUE_PUBLISH_RETRIES` (default 3).

Enfileiramento após detecção YOLO = **D3**; D2 deixa o backend pronto e visível no `/health`.

## EasyPanel (pilot A e B)

Copiar do worker (senha só no painel):

```env
REDIS_URL=redis://default:SENHA@185.130.61.5:6379/0
QUEUE_BACKEND=redis
EVENT_QUEUE_KEY=confvision:eventos
EVENT_QUEUE_MAX_SIZE=1000
# EVENT_QUEUE_DLQ_KEY=confvision:eventos:dlq
# QUEUE_PUBLISH_RETRIES=3
```

Redeploy **A** depois **B** (build Rust com dependência `redis`).

## Verificar

```bash
bash confvision-rust-processor/scripts/d2-redis-verify.sh \
  "https://foxpro-rust-pilot.rkr351.easypanel.host" \
  "https://foxpro-rust-pilot-b.rkr351.easypanel.host"
```

Esperado no `/health`:

- `queue_backend`: `redis`
- `event_queue_redis_ok`: `true`
- `event_queue_key`: `confvision:eventos`

## Rollback

```env
QUEUE_BACKEND=none
```

Redeploy (Redis URL pode permanecer).
