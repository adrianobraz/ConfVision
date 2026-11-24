# Monitor de infraestrutura

Serviço independente, sem banco e sem fila. A coleta fica em memória.
O ConfCam, no servidor dedicado, consulta `GET /api/v1/overview` e grava no PostgreSQL `confmonit`.

## Execução local

```text
MONITOR_API_SECRET=... 
MONITOR_LISTEN=127.0.0.1:8090
go run ./cmd/monitor
```

Fora do loopback o processo só sobe com `MONITOR_TLS_CERT` e `MONITOR_TLS_KEY`.
Este repositório não abre porta, firewall nem publica o binário.

## Coleta

`/proc` e cgroup quando existem. No Windows a leitura de host volta `indisponivel`.
Docker, se o socket local existir, só aceita GET de lista, inspect, stats, logs e top.
MediaMTX: `GET /v3/paths/list` em `MEDIAMTX_API_BASE`.
Redis: `INFO`, `TYPE` e contagem. Filas conhecidas no código do worker: `confvision:eventos` (lista) e `confvision:yolo:queue` (sorted set).
`[METRICS]` vem de `METRICS_LOG_PATH` ou do log do container do worker. Campo `-` não vira zero.

A linha real não tem FPS configurado nem contador de reconnect. Esses campos ficam ausentes. `cameras_em_backoff` existe só na linha de processo, não por câmera.

## ConfCam

Página `/administrator/infraestrutura`. O login continua em `/administrator/rtmp-falhas`.
O ciclo de gravação só liga com `MONITOR_BASE_URL` e `MONITOR_API_SECRET`.
A migração `sql/029_vis_infra.sql` não foi aplicada.
