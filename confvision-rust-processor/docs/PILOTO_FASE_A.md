# Piloto — Fase A (1 câmera oficial)

Isolar **uma** câmera no Rust, devolver as demais ao worker Python, alinhar env e validar operação.

**Contexto foxpro (2026-09):** câmeras **3, 4, 5** estavam com `worker_id=rust-processor-pilot-01`. Câmera **4** → RTSP **404** (`cam/boezyjmydlk4`). Recomendado piloto: **câmera 5** (ou **3**).

---

## Checklist

| # | Item | Onde |
|---|------|------|
| A1 | Só **1** câmera com `worker_id=rust-processor-pilot-01` | Postgres |
| A2 | `MAX_CAMERAS=1`, `LOG_LEVEL=info`, `SYNC_FILTER_WORKER_ID=true` | EasyPanel → rust-pilot |
| A3 | Câmeras **3, 4** (e outras) de volta ao `WORKER_ID` do Python | Postgres |
| A4 | **confvision** (MediaMTX) running | EasyPanel |
| A5 | **confvision-worker** — **só depois** de A7 ok | EasyPanel (Play por último) |
| A6 | Restart **rust-pilot** (sem rebuild) | EasyPanel |
| A7 | Validar `/health`, `/metrics`, logs | Script ou curl |
| A8 | Baseline CPU 30–60 min (parado vs movimento) | EasyPanel gráficos |

---

**Abrir arquivos / caminhos:** [`PILOTO_FASE_A_ABRIR_AQUI.md`](./PILOTO_FASE_A_ABRIR_AQUI.md)

---

## A1 + A3 — Postgres (`vis_camera` — sem CREATE TABLE)

Arquivo: [`../sql/piloto_fase_a_isolamento.sql`](../sql/piloto_fase_a_isolamento.sql)

1. Conectar ao Postgres da visão (`visionpsql` ou host documentado).
2. Rodar **SELECT** de inventário (início do SQL).
3. Anotar `PYTHON_WORKER_ID` — deve ser **igual** ao env `WORKER_ID` do serviço **confvision-worker** no EasyPanel (ex.: `worker-docker-21` na VPS sem GPU; **confirme no painel**).
4. Definir câmera piloto (padrão sugerido: **id = 5**).
5. Executar bloco **UPDATE** (descomentado após revisar).
6. Aguardar até **60 s** (`SYNC_INTERVAL_SEC`) ou restart rust-pilot.

---

## A2 + A5 + A6 — EasyPanel

1. **rust-pilot → Environment:** copiar de [`../easypanel.env.fase-a.example`](../easypanel.env.fase-a.example) e colar segredos reais (API key, RTMP secret).
2. Alterar se necessário:
   - `LOG_LEVEL=info` (remover `debug`)
   - `MAX_CAMERAS=1`
   - Remover `RUST_LOG` se existir
   - `LOAD_ADMISSION_ENABLED=0` (admission = fase C)
3. **Salvar** → **Restart** rust-pilot.
4. **confvision:** MediaMTX **on** (RTSP da câmera piloto).
5. **confvision-worker:** manter **parado** até A7 passar; depois Play com `WORKER_ID` = mesmo valor usado no SQL (`PYTHON_WORKER_ID`).

---

## A7 — Validação

Na VPS ou máquina com curl:

```bash
bash confvision-rust-processor/scripts/validate-piloto-fase-a.sh \
  https://foxpro-rust-pilot.rkr351.easypanel.host
```

Esperado após sync:

| Sinal | Ok |
|--------|-----|
| `/health` | `"status":"ok"`, `cameras_total`: **1**, `cameras_online`: **1** |
| `/ready` | `"ready": true` |
| Logs | **Um** `worker started camera_id=<piloto>` |
| `/metrics` | `cameras` array length **1**; sem reconnect loop 404 |

Se ainda aparecer câmera 4: Postgres não atualizado ou sync ainda não rodou.

---

## A8 — Baseline CPU/RAM

| Janela | Ação | Anotar |
|--------|------|--------|
| 30 min | Cena **parada** na câmera piloto | CPU % rust-pilot (EasyPanel) |
| 15 min | **Movimento** na câmera | CPU %; `motion_detected` em `/metrics` |

Modelo (copiar para nota interna):

```text
Data:
Câmera piloto id:
CPU parado: ___%
CPU movimento: ___%
frames_received (1h):
capacity_state:
```

---

## Rollback rápido

```sql
-- Restaurar worker_id anterior da câmera piloto (valor anotado no SELECT inicial)
UPDATE vis_camera SET worker_id = '<PYTHON_WORKER_ID>' WHERE id = <PILOTO_CAMERA_ID>;
```

Stop rust-pilot; worker Python retoma em 1–2 syncs.

---

## Próximo

Fase **B** — fechamento formal (handoff, tag git). Ver roadmap em conversa / README piloto.
