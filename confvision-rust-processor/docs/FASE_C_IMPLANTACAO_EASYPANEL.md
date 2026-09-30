# Fase C — implantar C1–C4 na VPS foxpro

Ordem obrigatória. **C5/C6** ficam fora (GPU e fila prod = dedicado / Fase D).

## Pré-requisito Postgres

```bash
# Revisar SELECTs, depois COMMIT
psql "$POSTGRES_URL" -f sql/cameras_404_worker_python.sql
psql "$POSTGRES_URL" -f sql/phase_c_foxpro_split_pilot02.sql   # só ao subir 2º processor
```

---

## C3 — Admission (rust-pilot-01)

1. EasyPanel → serviço **foxpro-rust-pilot** → **Environment**
2. Copiar de `easypanel.env.fase-c.vps.example` (ou ajustar manualmente):

```env
MAX_CAMERAS=10
LOAD_ADMISSION_ENABLED=1
LOAD_POLICY_MODE=admission
```

3. **Redeploy**
4. Validar:

```bash
curl -fsS "$URL/capacity-report" | jq '.load'
# Esperado: load_admission_enabled=true, admission_active=true
# Em critical: allow_new_camera=false
bash scripts/phase-c-verify.sh --strict-c3 "$URL"
```

---

## C2 — Observabilidade

**Opção mínima (core-4 ou host com curl):**

```bash
# /etc/cron.d/confvision-rust-pilot
*/5 * * * * root /path/confvision-rust-processor/scripts/phase-c-verify.sh https://foxpro-rust-pilot.rkr351.easypanel.host >> /var/log/rust-pilot-verify.log 2>&1
```

**Prometheus textfile:** ver `deploy/observability/README.md`

**Uptime Kuma:** GET `/health`, keyword `"status":"ok"`

---

## C1 — A/B baseline

1. Escolher **1 câmera** (ex. id 5).
2. Janela A: `worker_id` Python — anotar CPU EasyPanel worker.
3. Janela B: mesma câmera no Rust — rodar:

```bash
bash scripts/c1-ab-baseline.sh "$URL" 60 10 > /tmp/c1-rust.csv
```

4. Preencher `docs/c1-ab-results.template.md`

---

## C4 — 2º processor (pilot-02)

Só após C3 ON e `/capacity-report` recomendar `split_second_processor`.

1. **Novo app** EasyPanel (mesma imagem Git `rust-pilot`), domínio ex.: `foxpro-rust-pilot-02...`
2. Environment: `easypanel.env.fase-c.processor-02.example`
3. Executar SQL `phase_c_foxpro_split_pilot02.sql` (ajuste ids)
4. Redeploy **01** e **02**
5. Validar cada base URL:

```bash
bash scripts/phase-c-verify.sh https://foxpro-rust-pilot.rkr351.easypanel.host
bash scripts/phase-c-verify.sh https://foxpro-rust-pilot-02.SEU_DOMINIO
```

Detalhes: `deploy/phase-c/EASYPANEL_SECOND_PROCESSOR.md`

---

## Verificação única pós-implantação

```bash
bash scripts/phase-c-run-all.sh https://foxpro-rust-pilot.rkr351.easypanel.host
```
