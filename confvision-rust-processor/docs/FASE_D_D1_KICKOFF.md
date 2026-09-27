# D1 — Kickoff (2º processor na foxpro)

## Pré-requisitos

- [ ] **confvision-worker** → **Stop** (Rust-only).
- [ ] **confvision** (MediaMTX) → **Running**.
- [ ] **foxpro-rust-pilot** (pilot-01) → `/health` ok.
- [ ] Câmeras 404 conhecidas pausadas ou fora do split (ex. id **2** Entrada Principal).
- [ ] Env pilot-01: `MAX_CAMERAS=3` ou `4` se CPU critical (EasyPanel → Ambiente → Redeploy).

---

## 1. Criar app EasyPanel `foxpro-rust-pilot-02`

| Campo | Valor |
|-------|--------|
| Git | `adrianobraz/ConfVision` branch **`rust-pilot`** |
| Dockerfile | `confvision-rust-processor/Dockerfile` |
| Rede | Mesma de `foxpro_confvision` |
| Domínio | Novo (ex. `foxpro-rust-pilot-02.rkr351.easypanel.host`) |

**Environment:** copiar [`easypanel.env.fase-c.processor-02.example`](../easypanel.env.fase-c.processor-02.example):

- `CONFVISION_API_URL=https://vision.confmonit2.com.br`
- `PROCESSOR_ID` / `WORKER_ID` = **`rust-processor-pilot-02`**
- `MAX_CAMERAS=3` (VPS conservador)
- `SYNC_FILTER_WORKER_ID=true`
- Mesmos `VIS_WORKER_API_KEY`, `RTMP_PUBLISH_SECRET`, `MEDIAMTX_RTSP_BASE` que pilot-01.

**Implantar** → **Start** → logs sem panic; `/health` 200.

---

## 2. Postgres (Go — sem Xano)

Rodar **SELECT** em [`sql/phase_d_foxpro_split_vps_conservative.sql`](../sql/phase_d_foxpro_split_vps_conservative.sql).

Ajustar lista de `id` conforme inventário. **BEGIN/COMMIT** só após revisar.

Regra: cada câmera analítica → **um** `worker_id` (`pilot-01` **ou** `pilot-02`).

Aguardar `SYNC_INTERVAL_SEC` (~60s) ou restart dos dois rust apps.

---

## 3. Verificar

```bash
bash confvision-rust-processor/scripts/d1-processor-verify.sh \
  "https://foxpro-rust-pilot.rkr351.easypanel.host" \
  "https://SEU-DOMINIO-PILOT-02/health"
```

Esperado:

- `processor_id` distintos (`rust-processor-pilot-01` vs `-02`).
- Soma `cameras_total` ≈ fleet analítica; nenhum processor com 0 se SQL correto.
- `rtsp_404_count=0` em cada `/capacity-report` (ideal).

---

## 4. Rollback

```sql
UPDATE vis_camera SET worker_id = 'rust-processor-pilot-01'
WHERE worker_id = 'rust-processor-pilot-02';
```

EasyPanel → **Stop** `foxpro-rust-pilot-02`.

---

## Limitação VPS

Dois apps no **mesmo host** não dobram CPU disponível — só **particionam câmeras** entre processos. Escala N **hosts** = repetir este padrão (D1 em cada nó). GPU = **D4** (servidor novo).

Detalhe: [`deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md`](../deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md).
