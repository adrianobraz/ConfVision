# Load admission (Fase 6.2)

Controla **entrada de novas câmeras** no sync quando a capacidade está **critical**. Sessões já abertas **não** são derrubadas.

## Variáveis

| Variável | Default típico | Efeito |
|----------|----------------|--------|
| `LOAD_ADMISSION_ENABLED` | `1` (default) | `1` + capacity **critical** ou `estimated_available_cameras<=0` → bloqueia **novas** câmeras; entra na **fila** (`cameras_pending_admission`) |
| `LOAD_POLICY_MODE` | `admission` | Em `admission`, métricas usam advisory `reject_admission` em critical; o bloqueio efetivo de sync usa `LOAD_ADMISSION_ENABLED=1` |
| `LOAD_SHEDDING_ENABLED` | `0` (default) | `1` derruba workers ativos sob CPU/RAM — use só emergência; preferir admission + fila |

## Comportamento

- **OFF** (`LOAD_ADMISSION_ENABLED=0`): sync continua adicionando câmeras até `MAX_CAMERAS` (hard limit); capacity pode ficar `saturated` nas métricas.
- **ON** (`LOAD_ADMISSION_ENABLED=1`): em `capacity_state=critical` (ou sem headroom estimado), `allow_new_camera()` retorna false → log `câmera aguardando recurso (fila admission)`; re-tentativa a cada sync (~60s).
- **`LOAD_POLICY_MODE=disabled`**: desliga política de advisory; admission ainda depende de `LOAD_ADMISSION_ENABLED` para bloqueio.

Fase C na VPS: env completo em `easypanel.env.fase-c.vps.example` e checklist `scripts/phase-c-verify.sh`.

## Piloto EasyPanel (recomendado)

1. Subir carga / medir com `/metrics` e **`GET /capacity-report`**.
2. Se `recommended_actions` incluir `enable_load_admission`, definir no Environment:
   ```env
   LOAD_ADMISSION_ENABLED=1
   LOAD_POLICY_MODE=admission
   ```
3. Redeploy rust-pilot; validar com `scripts/capacity-report.sh`.

## Relatório operacional

```bash
curl -fsS "https://foxpro-rust-pilot.rkr351.easypanel.host/capacity-report" | jq '.load, .recommended_actions'
```

Campos úteis: `load.allow_new_camera`, `load.advisory`, `summary.rtsp_404_count`, `summary.cameras_pending_admission`, `/health` → `cameras_pending_admission`.

## Ordem recomendada antes de 2º processor

1. Mover câmeras **404** para Python (`sql/cameras_404_worker_python.sql`).
2. Ativar **admission** se ainda entrar câmera nova em critical.
3. Só então **`split_second_processor`** (novo serviço, novo `PROCESSOR_ID` / `WORKER_ID`, shard no Postgres).
