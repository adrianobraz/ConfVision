# DRAINING — MediaMTX Node (operacional)

**DRAINING não é implementado dentro do MediaMTX.** É estado do **Control Plane** + procedimento de frota.

## Semântica

1. Control Plane marca `vis_mediamtx_node.status = draining` (futuro) ou operador seta `CONFVISION_NODE_STATUS=DRAINING` (metadado).
2. **Novas câmeras** não recebem `vis_mediamtx_node_id` deste nó (PickMediamtxNode / assignment).
3. **Publish existente** continua: MediaMTX mantém paths ativos; RTMP-GUARD continua auth por câmera (sem alteração).
4. Operador migra câmeras gradualmente (UPDATE node + worker no CP).
5. Quando API MTX `/v3/paths/list` não tiver publishers ativos, node pode desligar.

## O que o node V1 fornece

- Métricas `:9998` e API `:9997` para contar streams ativos antes do shutdown.
- Restart graceful: SIGTERM → `start_mediamtx_guard.py` encerra Guard + MediaMTX juntos.

## O que NÃO fazer

- Não derrubar container com publishers ativos sem janela de migração.
- Não implementar assignment/lease no YAML MediaMTX.
