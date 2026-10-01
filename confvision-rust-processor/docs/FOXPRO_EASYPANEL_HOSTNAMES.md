# Foxpro EasyPanel — hostnames internos (2026-10)

Serviço MediaMTX + Guard no painel: app **`rust-mediamtx`** (rede Docker).

| Público | Interno (container→container) | Porta |
|---------|----------------------------------|-------|
| `https://foxpro-confvision.rkr351.easypanel.host/` | `http://foxpro_rust-mediamtx:8100/` | Guard (admin, `/online`, `/auth` local MTX) |
| `https://hls1.dnsid.com.br/` | `http://foxpro_rust-mediamtx:8888/` | HLS |

## Rust processor (mesma rede foxpro)

```env
MEDIAMTX_RTSP_BASE=rtsp://foxpro_rust-mediamtx:8554
MEDIAMTX_NODE_ID=1
```

Cadastro câmera (`rtsp_url_sec`):

```text
rtsp://foxpro_rust-mediamtx:8554/cam/{hash12}
```

## Go core-4 (proxy `/api/rtmp-online`, fora da rede Docker)

```env
RTMP_GUARD_URL=https://foxpro-confvision.rkr351.easypanel.host
RTMP_WATCH_URL=https://foxpro-confvision.rkr351.easypanel.host
```

Teste: `curl -s -H "X-RTMP-Guard-Key: …" https://foxpro-confvision.rkr351.easypanel.host/online`

## Workers Python (mesma rede)

```env
MEDIAMTX_RTSP_BASE=rtsp://foxpro_rust-mediamtx:8554
MEDIAMTX_API_BASE=http://foxpro_rust-mediamtx:9997
```

## Postgres `vis_mediamtx_node` (RTSP interno para sync)

```sql
UPDATE vis_mediamtx_node
SET rtsp_internal = 'rtsp://foxpro_rust-mediamtx:8554'
WHERE id = 1;
```

(Ajuste `id` e `rtmp_public` / `hls_public` conforme DNS RTMP/HLS de produção.)

Legado **`foxpro_confvision`** — substituído por **`foxpro_rust-mediamtx`** neste deploy.

## Erro `Name or service not known` com `foxpro_confvision` no log

O Rust **usa `rtsp_url_sec` da API tal qual** quando o campo está preenchido; `MEDIAMTX_RTSP_BASE` no EasyPanel **não substitui** essa URL.

Corrigir **cadastro** (e opcionalmente o nó MTX):

```sql
-- Todas as câmeras com host legado no RTSP interno
UPDATE vis_camera
SET rtsp_url_sec = REPLACE(rtsp_url_sec, 'foxpro_confvision', 'foxpro_rust-mediamtx')
WHERE rtsp_url_sec LIKE '%foxpro_confvision%';

UPDATE vis_mediamtx_node
SET rtsp_internal = 'rtsp://foxpro_rust-mediamtx:8554'
WHERE rtsp_internal LIKE '%foxpro_confvision%' OR id = 1;
```

Câmera **32** (API): `rtsp_url_sec` = `rtsp://foxpro_rust-mediamtx:8554/cam/qjg3p8j296z7`.

Após salvar, reinicie o processor ou aguarde o sync — o log deve mostrar `url=rtsp://foxpro_rust-mediamtx:8554/...`.
