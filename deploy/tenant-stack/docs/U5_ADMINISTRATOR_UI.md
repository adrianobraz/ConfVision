# U5 — Painéis Administrator

## Painel 2 — Cluster / capacidade (ADM)

**Rota UI:** `/relatorio-operacional` (ConfVision)

**APIs (proxy interno):**

| Tab / uso | Backend |
|-----------|---------|
| Health | `GET /vis_relatorio_operacional/health` |
| Métricas | `GET /vis_relatorio_operacional/metric` |
| Stream eventos | `GET /vis_relatorio_operacional/stream` |

**D5 (capacidade Rust):** `GET /vis_rust_processor_capacity`  
Filtro por host: `?servidor_id=srv-confvision-042`

Dados persistidos pela coleta (`COLETA_RELATORIO_ENABLED=1`):

- `vis_sistema_health` — componente `rust_processor` + `base_url`
- `vis_sistema_metric` — chave `rust_processors` (JSON com `servidor_id` quando registry usa `servidor|url`)

## Painel 1 — Câmera (drill-down)

Fluxo existente ConfVision: câmera → stream / analítico / `worker_id`.

Após U6, assign manual ou auto com escopo:

```http
POST /vis_camera_assign_processor
{"camera_id": 123, "servidor_id": "srv-confvision-042"}
```

## Próximos passos UI (opcional)

- Coluna `servidor_id` na tabela capacity no `relatorio-operacional.js`
- Filtro dropdown de servidores parseado de `processors[].servidor_id`
