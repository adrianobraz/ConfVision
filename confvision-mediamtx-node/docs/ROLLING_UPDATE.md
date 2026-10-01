# Rolling update — frota MediaMTX

## Princípio

Atualizar **fatias** da frota, nunca 100% simultâneo sem validação.

## Sequência sugerida

| Fase | % nodes | Validação |
|------|---------|-----------|
| 1 | 5% | RTMP publish, RTSP Rust, metrics, DVR patch record |
| 2 | 25% | Erros auth, reconnect NVR, HLS on-demand |
| 3 | 50% | Carga, bandwidth, disco `/recordings` |
| 4 | 100% | Restante |

## Por node

1. Preferir **novos** nodes na versão alvo antes de drenar antigos.
2. Marcar node antigo **DRAINING** (ver DRAIN.md).
3. Deploy imagem `confvision-mediamtx-node:1.0.0` (`MEDIAMTX_VERSION=1.21.0`, `CONFIG_VERSION=confvision-mtx-v1.0.0`).
4. Hot reload: MediaMTX API `PATCH /v3/config/global` para mudanças **não destrutivas**; troca de imagem = rollout container.

## Compatibilidade

- Mesmo path `cam/{hash}`, mesma URL auth `127.0.0.1:8100/auth`.
- Rust inalterado: RTSP `rtsp://<node>:8554/cam/{hash}`.
