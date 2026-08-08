# Migração Xano → Postgres central (Completa — operacional)

## Arquitetura

| Onde | Função |
|------|--------|
| **Postgres central** | Câmeras, licenças, áreas, eventos, gravação, workers, nós |
| **Core4 Go** | Site + **API confVision** (substitui Xano operacional) |
| **VPS** | Workers vídeo — apontam `XANO_BASE_URL` para Core4 Go |
| **Xano CVG** | Grade horária (`XANO_CVG_BASE_URL`) — pendente migrar |
| **Xano fp_franqueadoPro** | Faturas/cobrança automática (opcional manter) |

## Comportamento com `POSTGRES_URL` definido

- **Cadastro de câmera** (`POST /api/cameras`) → Postgres (não Xano)
- **Licenças, eventos, gravação, sync worker** → Postgres
- **`proxyXano`** tenta Postgres primeiro; fallback Xano se rota não implementada
- **Grade horária** continua via `proxyXanoCvg` → Xano CVG

## 1. Schema

```powershell
cd confvision\sql\cmd\apply
$env:POSTGRES_URL = "postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable"
go run .
go run . ..\..\003_extend_schema.sql
go run . -check
```

## 2. Importar dados do Xano

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

## 4. VPS workers (EasyPanel)

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

## Pendente (fase 2)

- Grade horária CVG → Postgres
- Webhook fatura Xano → criar licença Postgres
- Terminal/receptor (`vis_evento_disparo_sensor`, queries processo)
