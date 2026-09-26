# Load admission (Fase 6.2)

Controla **entrada de novas câmeras** no sync quando a capacidade está **critical**. Sessões já abertas **não** são derrubadas.

## Variáveis

| Variável | Default típico | Efeito |
|----------|----------------|--------|
| `LOAD_ADMISSION_ENABLED` | `0` (piloto) / `1` (produção saturada) | `1` + capacity **critical** → bloqueia **novas** câmeras no `CameraManager` |
| `LOAD_POLICY_MODE` | `advisory` | Em `admission`, métricas usam advisory `reject_admission` em critical; o bloqueio efetivo de sync usa `LOAD_ADMISSION_ENABLED=1` |

## Comportamento

- **OFF** (`LOAD_ADMISSION_ENABLED=0`): sync continua adicionando câmeras até `MAX_CAMERAS` (hard limit); capacity pode ficar `saturated` nas métricas.
- **ON** (`LOAD_ADMISSION_ENABLED=1`): em `capacity_state=critical`, `allow_new_camera()` retorna false → log `load admission rejected new camera`.
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

Campos úteis: `load.allow_new_camera`, `load.advisory`, `summary.rtsp_404_count`.

## Ordem recomendada antes de 2º processor

1. Mover câmeras **404** para Python (`sql/cameras_404_worker_python.sql`).
2. Ativar **admission** se ainda entrar câmera nova em critical.
3. Só então **`split_second_processor`** (novo serviço, novo `PROCESSOR_ID` / `WORKER_ID`, shard no Postgres).
