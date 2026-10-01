# Plano de testes — ConfVision MediaMTX Node V1

## Carga progressiva (por node)

1 → 10 → 50 → 100 → 500 → 1000 publishers RTMP **simulados ou reais**.

Medir a cada degrau:

- CPU / RAM host
- Rede (bitrate agregado)
- Conexões RTMP / RTSP / HLS readers
- Latência HLS (se houver viewers)
- Disco I/O se `record=true` em subset
- Erros log MediaMTX + Guard
- Disconnect / reconnect NVR

## Funcional

- [ ] Publish `cam/{hash}` auth 200 / 403 casos (inativa, bloqueada, sistema_stream)
- [ ] RTSP read Rust (sem alterar binário — smoke URL)
- [ ] DVR `sync_record_paths` PATCH path API v3
- [ ] Metrics scrape `:9998`
- [ ] HLS só com viewer (sem remux global)

## Falha

| Cenário | Esperado |
|---------|----------|
| Restart container | NVR reconecta; paths dinâmicos |
| Guard down no boot | Log warn; auth refused até Guard up |
| CP API down | Guard cache TTL; streams **já publicados** seguem (MTX local) |
| Auth down publish **novo** | 403/503 conforme Guard |
| Disco cheio | record falha local; node degradado |
| Node drain | Sem novas câmeras no CP; streams existentes OK |

## Ferramentas

- `rtmp_online.listar_online` (Guard)
- MediaMTX API `/v3/paths/list`
- Prometheus metrics endpoint

Números alimentam **Capacity Engine Rust** e **Media capacity CP** separadamente — não misturar.
