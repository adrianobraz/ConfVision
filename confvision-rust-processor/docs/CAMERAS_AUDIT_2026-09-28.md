# Auditoria câmeras ativas — export 2026-09-27

Fonte: export `vis_camera` (18 linhas `ativo=true`). Decisão: **worker Python off**; assign Rust já no CSV.

---

## 1) D5 assign — ✅ OK no Postgres (CSV)

Todas as 18 ativas têm `worker_id`:

| Processor | Qtd (ativo) |
|-----------|-------------|
| `rust-processor-pilot-a-01` | 9 |
| `rust-processor-pilot-b-02` | 9 |

**Não precisa** reassign em massa só por D5 — salvo rebalancear carga.

---

## 2) Sync Go → Rust — 🟡 bloqueio principal: `analitico_pausado`

O Go só envia no sync câmeras com:

- `ativo = true`
- `deteccao_humano = true`
- **`analitico_pausado IS NOT TRUE`**

No CSV, **a maioria** das ativas está com **`analitico_pausado = true`** (analítico **pausado** no cadastro). Só **6** aparecem elegíveis no export:

| id | nome (resumo) | worker_id |
|----|---------------|-------------|
| 5 | corredor mesanino | b-02 |
| 18 | FAZENDA MARANATA 1 | a-01 |
| 19 | FAZENDA MARANATA 2 | b-02 |
| 22 | FAZENDA CAICARA 3 | a-01 |
| 26 | FONTANA MOEGA 2 | a-01 |
| 27 | FONTANA MOEGA 3 | b-02 |

**Teste live (2026-09-28):** `GET /vis_camera_sync_ativas` retornou **`cameras: []`** (0). Conferir no **mesmo Postgres** do Go:

```sql
SELECT id, nome, ativo, deteccao_humano, analitico_pausado, worker_id
FROM vis_camera
WHERE ativo = TRUE AND deteccao_humano = TRUE AND (analitico_pausado IS NOT TRUE)
ORDER BY id;
```

Se o SQL retornar linhas e a API 0 → investigar Go/Postgres URL. Se SQL retornar 0 → **despausar** câmeras que devem analisar.

**Despausar (exemplo — ajuste IDs):**

```sql
UPDATE vis_camera SET analitico_pausado = FALSE
WHERE id IN (2,3,4,6,8,9,15,20,21,25,28,29);  -- revisar lista com negócio
```

Ou API: `POST /vis_camera/analitico/pausar/{id}` com body que **despausa** (conforme contrato do painel).

---

## 3) RTSP — 🟡 sem worker Python

- **`rtsp_url_sec` vazio** em todas as ativas → Rust usa `rtsp://…/cam/{hash}` no MediaMTX (`10.11.1.174:8554` nos erros).
- **Sem publisher RTMP** no path → **RTSP 404** (ids **2, 20, 21, 22** com `stream_erro_classe=rtsp_404` no CSV).
- **Câmera 4** (`Entrada escada pedestre`): assign **a-01**; no CSV **sem** rtsp_404 recente, mas **`rtsp_url_sec` vazio** — depende do mesmo path MediaMTX.

**Sem `confvision-worker`:** é obrigatório **ingest alternativo** (RTSP direto no cadastro, outro publicador, ou evolução Rust→RTMP). Ver [FASE_D_FECHAMENTO.md](./FASE_D_FECHAMENTO.md).

---

## 4) RAM critical — o que era

Alerta no `/health` do Rust: **`capacity_state=critical`**, **`limiting_resource=memory`** quando **muitas câmeras decodificam** na mesma VPS. Com **0 câmeras online** no processor, pode aparecer **healthy** — o aviso volta quando subir carga. **D5** não assigna câmera nova se `assign_eligible=false`.

---

## 5) D3 — “1º evento E2E” — o que falta

Pipeline desejado:

```text
vídeo → motion/YOLO (Rust) → events_published++ → fila Redis → POST /vis_evento → linha vis_evento
```

Enquanto **`events_published=0`** e **nenhuma linha nova** em `vis_evento` por detecção Rust, o **produto analítico** não fechou aceite D3 — mesmo com D5/D6 ok.

---

## 6) Próximos passos (ordem)

1. **Despausar** câmeras que clientes vão testar (`analitico_pausado=false`).
2. **Garantir stream RTSP** (MediaMTX com publisher **ou** `rtsp_url_sec` direto do DVR).
3. Confirmar sync: `curl …/vis_camera_sync_ativas?worker_id=rust-processor-pilot-a-01` → `"cameras":[…]`.
4. **Movimento** na cena → `/health` `events_published` > 0 → `SELECT … FROM vis_evento ORDER BY id DESC LIMIT 5`.

Script fila/health: `bash confvision-rust-processor/scripts/d3-event-verify.sh https://foxpro-rust-pilot…/health …`
