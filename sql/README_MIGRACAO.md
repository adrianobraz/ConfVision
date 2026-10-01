# Migração Xano → Postgres central (Completa — operacional)

## Arquitetura

| Onde | Função |
|------|--------|
| **Postgres central** | Câmeras, licenças, áreas, eventos, gravação, workers, nós |
| **Core4 Go** | Site + **API confVision** (substitui Xano operacional) |
| **VPS** | Workers vídeo — apontam `XANO_BASE_URL` para Core4 Go |
| **Xano fp_franqueadoPro** | Faturas/cobrança automática (opcional manter) |
| **Xano WhatsApp/ligação** | Filas operacionais (mantidas) |

## Comportamento com `POSTGRES_URL` definido

- **Cadastro de câmera** (`POST /api/cameras`) → Postgres (não Xano)
- **Licenças, eventos, gravação, sync worker** → Postgres
- **`proxyXano`** tenta Postgres primeiro; fallback Xano se rota não implementada
- **Grade horária** → Postgres via `proxyXanoCvg` → `visdata` (sem `XANO_CVG_BASE_URL`)

## 1. Schema

```powershell
cd confvision\sql\cmd\apply
$env:POSTGRES_URL = "postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable"
go run .
go run . ..\..\003_extend_schema.sql
go run . -check
```

## 2. Importar dados do Xano (vis_*)

```powershell
cd confvision\sql
.\export_xano.ps1
cd cmd\apply
go run . ..\..\003_import_xano.sql
go run . -check
```

## 3. Core4 Go

`.env`:

```env
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable
```

Deploy: `.\build.ps1` → FileZilla → `systemctl restart confmonit4confvision`

## 4. Ops: CVG + arme/desarme + autofim (Postgres)

Endpoints dedicados (fazer **push Xano** antes do export):

- `GET /cvg_export_all` (grupo `confVisionGrade`) — grade config/slots/escopo
- `GET /ops_export_janelas` (grupo `RoboAtendimento`) — todas as janelas arme/desarme

```powershell
cd confvision\sql\cmd\apply
go run . ..\..\004_ops_schema.sql
cd ..\..
.\export_ops_xano.ps1
cd cmd\apply
go run . ..\..\005_import_ops.sql
```

### taskxano `.env`

```env
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable
CONFVISION_GRADE_TICK_URL=https://vision.confmonit2.com.br/cvg_worker_tick
CONFVISION_GRADE_WORKER_KEY=<mesma chave>
ARME_DESARME_AUTO_QUERY_MODE=postgres
```

Com `POSTGRES_URL`: autofim e arme/desarme usam Postgres; WhatsApp/ligação continuam no Xano.

### eventgateway `.env`

```env
POSTGRES_URL=postgres://...
OPS_SIDECAR_ENABLED=true
```

Sidecar grava `ops_alarm_events` + enfileira autofim **após** encaminhar ao Xano (WhatsApp intacto).

### Xano — desligar writer autofim

Push `functions/22_funcao_sistema_send_whats_event_central_alarm.xs` (bloco `bot_finalizaeventoauto` comentado).

## 5. VPS workers (EasyPanel)

> **Documentação completa:** [`../VARIAVEIS_VPS.md`](../VARIAVEIS_VPS.md) — o que são `XANO_BASE_URL` e `EVENT_STORE`, onde colocar e o que **não** afeta.

Variáveis no **EasyPanel** (serviços Python: `confvision-worker`, `dvr`, `motion`, etc.) — **não** no `.env` do Go Core4:

```env
# Endereço da API operacional (nome legado; hoje aponta para o Go, não para o Xano)
XANO_BASE_URL=https://vision.confmonit2.com.br

# Eventos gravados no Postgres central (não no Xano)
EVENT_STORE=postgres
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable
```

| Variável | Significado em uma linha |
|----------|--------------------------|
| `XANO_BASE_URL` | URL base que o **worker Python** usa para buscar câmeras, áreas, ping (`/vis_*`) |
| `EVENT_STORE=postgres` | Detecções/eventos vão **direto** para o Postgres (`vis_evento`), sem passar pelo Xano |

**Erro 502 em `/cameras` no navegador** = problema no **Core4 Go** (seção 3), não nessas variáveis da VPS.

## Endpoints implementados (visdata)

Câmeras, áreas, licenças, eventos, clips, gravação storage/segmentos, workers, nós MediaMTX, sync, rtmp_auth.

- **Grade horária CVG** → Postgres (`vis_cliente_grade_*`, `/cvg_worker_tick`)
- **Arme/desarme auto** → Postgres (`ops_arme_janela`)
- **Autofim** → Postgres (`ops_bot_finalizaeventoauto` + sidecar eventgateway)
- Webhook fatura Xano → criar licença Postgres
- Terminal/receptor (`vis_evento_disparo_sensor`, queries processo) — **implementados no Go** (`eventos_terminal.go`)
