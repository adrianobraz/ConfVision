# Variáveis da VPS EasyPanel — o que são e onde colocar

Este arquivo explica **`XANO_BASE_URL`** e **`EVENT_STORE`** após a migração Xano → Postgres.

> **Importante:** essas variáveis vão **só nos serviços Python da VPS EasyPanel**  
> (`confvision-worker`, `confvision-dvr`, `confvision-motion`, `confvision-timelapse`, MediaMTX/Guard).  
> **Não** vão no `.env` da app Go do Core4 (`home/confmonit/v4.0/confvision/.env`).

---

## Três sistemas diferentes (não misturar)

| Sistema | Onde roda | Para quê |
|---------|-----------|----------|
| **Site Go (Core4)** | `185.130.61.4` → Apache → Go `:8086` | Login, cadastro de câmeras, tela `/cameras` |
| **Workers Python (EasyPanel)** | VPS foxpro (EasyPanel) | Detecção YOLO, gravação DVR/motion, sync de câmeras |
| **Postgres central** | `191.96.156.116:5432` | Banco `confmonit` — câmeras, licenças, eventos |

Configurar o EasyPanel **não corrige** erro no navegador.  
Erro **502** em `/cameras` = problema no **Core4 Go** (deploy/restart do binário).

---

## `XANO_BASE_URL`

### O que é

Endereço base da **API operacional** que os workers Python chamam.

O nome continua `XANO_BASE_URL` por compatibilidade com o código antigo, mas **não precisa ser o Xano**.

### Antes (Xano)

```env
XANO_BASE_URL=https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW
```

Workers chamavam o Xano, por exemplo:

- `GET …/vis_camera_sync_ativas` — lista câmeras para analítico
- `POST …/vis_evento` — registrar evento
- `GET …/vis_camera/rtmp_auth/{id}` — validar publicação RTMP

### Depois (Core4 Go — migração)

```env
XANO_BASE_URL=https://vision.confmonit2.com.br
```

Workers passam a chamar a **mesma API** no servidor Core4 (Apache faz proxy para o Go).

Exemplos reais de URL que o worker monta:

```
https://vision.confmonit2.com.br/vis_camera_sync_ativas?worker_id=worker-01
https://vision.confmonit2.com.br/vis_evento
https://vision.confmonit2.com.br/vis_worker_ping
```

### Onde colocar

No **EasyPanel**, aba **Environment** de cada serviço que usa API:

| Serviço EasyPanel | Precisa de `XANO_BASE_URL`? |
|-------------------|----------------------------|
| `confvision-worker` | Sim |
| `confvision-dvr` | Sim |
| `confvision-motion` | Sim |
| `confvision-timelapse` | Sim |
| `confvision` (MediaMTX + Guard) | Sim (auth RTMP) |
| App Go Core4 | **Não** — lá usa `POSTGRES_URL`, `XANO_BASE_URL` antigo só para fallback/grade |

### Alternativa (sem passar pelo Apache)

Se o Core4 expuser a porta Go diretamente (`BIND_HOST=0.0.0.0`):

```env
XANO_BASE_URL=http://185.130.61.4:8086
```

Na prática usamos `https://vision.confmonit2.com.br` (SSL + domínio fixo).

---

## `EVENT_STORE`

### O que é

Define **onde os eventos de detecção são gravados** quando o worker detecta movimento/pessoa.

### Valores possíveis

| Valor | Comportamento |
|-------|----------------|
| `xano` | Grava eventos só no Xano (legado) |
| `postgres` | Grava eventos **direto no Postgres central** |
| `dual` | Postgres + sync/cópia para Xano |

### Migração recomendada

```env
EVENT_STORE=postgres
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable
```

Com `EVENT_STORE=postgres`:

- Worker **não** envia `POST /vis_evento` para gravar o evento principal
- Escreve na tabela `vis_evento` (e clips em `vis_evento_clip`) no Postgres
- Ainda usa `XANO_BASE_URL` para **ler** config (câmeras ativas, áreas, ping)

### Onde colocar

EasyPanel → serviços que **geram eventos**:

| Serviço | `EVENT_STORE=postgres`? |
|---------|-------------------------|
| `confvision-worker` | Sim |
| `confvision-dvr` | Opcional (grava segmentos, não eventos analíticos) |
| `confvision-motion` | Sim |
| MediaMTX / Guard | Não |

---

## Resumo — o que colocar em cada lugar

### VPS EasyPanel (workers Python)

```env
# API operacional (substitui o Xano)
XANO_BASE_URL=https://vision.confmonit2.com.br

# Eventos vão para o Postgres, não para o Xano
EVENT_STORE=postgres
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable

# Demais (já existentes)
MEDIAMTX_RTSP_BASE=rtsp://foxpro_confvision:8554
WORKER_ID=worker-01
# ...
```

### Core4 Go (`/home/confmonit/v4.0/confvision/.env`)

```env
# Banco central — site + API /vis_* + /api/cameras
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable

# Fallback Xano (rotas ainda não migradas) e cobrança
XANO_BASE_URL=https://xpcy-oyme-lno7.b2.xano.io/api:AC7rgWwW
XANO_CVG_BASE_URL=...
XANO_API_FRANQUEADO_PRO=...
```

**Não** coloque `EVENT_STORE` no Core4 Go — essa variável é só dos workers Python.

---

## Como testar se está certo

### 1. API acessível pelos workers

```bash
curl https://vision.confmonit2.com.br/vis_health
# esperado: {"status":"ok","enabled":true}
```

### 2. Log do worker ao subir

```text
[CONFIG] OK | xano=https://vision.confmonit2.com.br | ... | event_store=postgres
```

### 3. Site (outro caminho)

Abrir `/cameras` logado no navegador. Se der 502, ver logs:

```bash
sudo journalctl -u confmonit4confvision -f
```

Isso é deploy/restart do **binário Go**, não config do EasyPanel.

---

## Referências

- Migração completa: [`sql/README_MIGRACAO.md`](sql/README_MIGRACAO.md)
- Deploy EasyPanel: [`DEPLOY_EASYPANEL.md`](DEPLOY_EASYPANEL.md)
- Deploy Core4 Go: [`../home/confmonit/v4.0/confvision/DEPLOY_PROXMOX.md`](../home/confmonit/v4.0/confvision/DEPLOY_PROXMOX.md)
