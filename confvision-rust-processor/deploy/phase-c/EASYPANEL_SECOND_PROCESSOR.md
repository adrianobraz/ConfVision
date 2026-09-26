# C4 — segundo rust-processor no EasyPanel

## Quando

- `capacity-report` → `recommended_actions` contém `split_second_processor`
- C3 (`LOAD_ADMISSION_ENABLED=1`) já ativo no pilot-01
- `rtsp_404_count = 0`

## Serviço novo

| Campo | pilot-01 | pilot-02 |
|--------|----------|----------|
| App name | foxpro-rust-pilot | foxpro-rust-pilot-02 |
| Git branch | rust-pilot | rust-pilot |
| Dockerfile | confvision-rust-processor/Dockerfile | idem |
| `PROCESSOR_ID` | rust-processor-pilot-01 | rust-processor-pilot-02 |
| `WORKER_ID` | idem | idem |
| `MAX_CAMERAS` | 6–10 | 4–6 |
| Domínio | foxpro-rust-pilot... | subdomínio novo |
| Porta interna | 8090 | 8090 |

Mesma rede Docker que `foxpro_confvision` (RTSP `MEDIAMTX_RTSP_BASE=rtsp://foxpro_confvision:8554`).

## Postgres

Rodar `sql/phase_c_foxpro_split_pilot02.sql` — revisar lista de ids antes do `COMMIT`.

## Rollback

```sql
UPDATE vis_camera SET worker_id = 'rust-processor-pilot-01'
WHERE worker_id = 'rust-processor-pilot-02';
```

Stop serviço pilot-02.

## Limitação VPS

Com CPU já ~77%+ em 10 câmeras num nó, **split em 2 apps no mesmo host** só ajuda se cada app tiver **limite de CPU** (cgroups) ou se mover **metade** das câmeras reduzir carga por processo. Ideal: **2º processor no dedicado** quando existir.
