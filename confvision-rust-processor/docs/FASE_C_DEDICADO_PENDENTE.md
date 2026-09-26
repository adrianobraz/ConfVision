# Fase C — pendências no servidor dedicado (pós-VPS)

Use esta lista quando o **GEX44** (ou dedicated com **NVIDIA**) estiver provisionado. Nada aqui bloqueia o fechamento da Fase C **na VPS foxpro**.

## Hardware e runtime

- [ ] Driver NVIDIA + `nvidia-container-toolkit` no host
- [ ] EasyPanel/Swarm: runtime GPU no serviço rust-processor e confvision-worker
- [ ] Mount `/dev/shm/confvision` para frame store (worker distributed)

## C5 — Decode / inferência GPU

- [ ] `VIDEO_ACCELERATION=gpu` (ou conforme doc da branch) no rust-processor
- [ ] Validar log: `hardware_decode_active=true`, `video_hw_backend` ≠ `none`
- [ ] Monitorar `hw_fallback_to_cpu_count` (fallback saudável)

## C4 — Escala N processors (modelo alvo)

- [ ] `rust-processor-node-01` … `node-N`, cada um `MAX_CAMERAS=2–4` (CPU) ou 20–50 (GPU)
- [ ] Postgres: `worker_id` = `PROCESSOR_ID` por câmera; sem monolito `MAX_CAMERAS=600` num Rust só
- [ ] Opcional futuro: `SHARD_MODE` + índices quando Go + Rust estiverem alinhados

## C6 — Fila e eventos (prod)

- [ ] Redis dedicado (ou cluster) na rede interna; `REDIS_URL` nos serviços
- [ ] Worker Python `SCHEDULER_BACKEND=redis`, `YOLO_ARCH=distributed` ([`worker.env.gex44-gpu.example`](../../easypanel/worker.env.gex44-gpu.example))
- [ ] Implementar/consumir fila no Rust (`QUEUE_BACKEND=redis`) — **Fase D**, hoje só metadado no ping
- [ ] Contrato de mensagem: motion → YOLO → evento/clips; dead-letter e retry

## C1 — A/B definitivo

- [ ] Mesma câmera, mesma janela: **legacy CPU (VPS)** vs **Rust+GPU (dedicated)**
- [ ] Métrica de negócio: frames analisados / alarmes úteis (quando YOLO ativo)

## C2 — Observabilidade produção

- [ ] Grafana dashboards por processor + por câmera (fps, latency, reconnects)
- [ ] Alertas Pager/WhatsApp: `0/1` replicas, OOM, `rtsp_404_count` > 0, fila Redis crescendo

## C3 — Control plane (Go)

- [ ] Go respeitar capacidade reportada no ping (`max_cameras`, advisory) para auto-assign câmera → processor — **enhancement**, não obrigatório no piloto VPS

## Não migrar para dedicado sem

- Rollback SQL de `worker_id` testado
- Backup env EasyPanel foxpro
- Runbook 404 + stream policy (`RUNBOOK_CAMERAS_404_ATIVO.md`, `STREAM_RETRY_POLICY.md`)
