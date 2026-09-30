# Observabilidade Fase C2

O rust-processor expõe **JSON** em `/metrics`, não formato Prometheus nativo.

## Opção A — Simples (VPS foxpro)

Cron a cada 5 min no **core-4** ou host com curl:

```bash
*/5 * * * * /path/scripts/phase-c-verify.sh https://foxpro-rust-pilot.rkr351.easypanel.host >> /var/log/rust-pilot-verify.log 2>&1
```

**Uptime Kuma:** monitor HTTP(s) para `/health` (keyword `ok`).

## Opção B — Métricas texto para Prometheus

```bash
scripts/phase-c-verify.sh --prometheus-text https://foxpro-rust-pilot.rkr351.easypanel.host > /var/lib/node_exporter/textfile/rust_pilot.prom
```

Node exporter `textfile collector` + regras em `alerts-rust-pilot.rules.yml`.

## Opção C — Stack completa (dedicado)

Prometheus + Grafana na rede interna; json_exporter ou custom scraper parseando `/metrics` e `/capacity-report`.

Campos prioritários por câmera: `camera_id`, `status`, `fps`, `rtsp_errors`, `stream_failures_consecutive`, `reconnect_count`.

Globais: `capacity_state`, `load_advisory`, `frames_dropped`, `rtsp_404_count` (capacity-report).
