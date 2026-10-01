# Deploy — terminal CV01

**Guia rápido:** [`SUBIR_AGORA.md`](SUBIR_AGORA.md) — arquivos `.env` reais, só subir.

---

## Checklist rápido

| # | Onde | Ação |
|---|------|------|
| 1 | **Proxmox** (Core4 Go) | Atualizar `.env` + subir binário + `systemctl restart` |
| 2 | **EasyPanel** (VPS worker) | Atualizar env + redeploy `confvision-worker` |
| 3 | **GitHub** | Push `core4` + push repo workers `ConfVision` |
| 4 | **Teste** | Detecção analítica → processo CV01 no terminal |

---

## 1. Core4 Go (Proxmox — `185.130.61.4`)

### Arquivos para subir (FileZilla)

Destino: `/home/confmonit/v4.0/confvision/`

```
confvision          ← binário Linux (build.ps1 no PC)
.env                ← editar no servidor (nunca commitar senhas)
recursos/           ← só se alterou HTML/JS/CSS
```

**Não subir:** `src/`, `go.mod`, `app.go`

### Build no PC (Windows)

```powershell
cd C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision
.\build.ps1
```

### SSH após subir binário

```bash
chmod +x /home/confmonit/v4.0/confvision/confvision
sudo systemctl restart confmonit4confvision
sudo journalctl -u confmonit4confvision -f
```

### `.env` — adicionar/ajustar (Core4 Go)

Copie de `.env.example` e preencha. **Obrigatório para terminal:**

```env
# Postgres central (vis_evento)
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable

# MySQL legado (checa UsuarioTerminal do franqueado)
BD_HOST_MV4=...
BD_USER_MV4=...
BD_PASS_MV4=...
BD_BASE_MV4=confmonitV4

# Terminal — dispatch CV01
# Core-4: receptorWeb escuta :5000 neste host (receptorWeb.sh) — use localhost
RECEPTOR_WEB_URL=http://127.0.0.1:5000
RECEPTOR_WEB_SENHA=<mesma senha do receptorWeb (SENHA_WEB)>
TERMINAL_NOTIFY_ENABLED=true
TERMINAL_NOTIFY_RETRIES=3
TERMINAL_CONTACT_ID=CV01
```

`RECEPTOR_WEB_SENHA` = variável `SENHA_WEB` em `/home/confmonit/v4.0/receptorWeb/.env`.

**Não use `http://185.130.61.3:5000`** — o IP `185.130.61.3` é o DNS do receptor de alarme (TCP 2030/2031). O HTTP do receptorWeb (`/recebe-evento-confvision`) roda no **core-4** na porta 5000.

Teste rápido no core-4:

```bash
ss -tlnp | grep 5000          # deve listar receptorWeb
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:5000/   # 404 = porta OK
curl -s -o /dev/null -w "%{http_code}\n" http://185.130.61.3:5000/  # connection refused = URL errada no .env
```

### Log esperado após detecção

```text
[TERMINAL] ok evento=12345 idEvento=... processo=... conta=0001 part=01 zona=001
```

---

## 2. VPS EasyPanel (workers Python)

### Serviço que precisa redeploy

| Serviço | Redeploy? |
|---------|-----------|
| `confvision-worker` | **Sim** |
| `confvision-sensor` | Não (sensor não usa CV01 analítico) |
| `confvision` (MediaMTX) | Não |
| `confvision-dvr` / `motion` / `timelapse` | Não |

### Como redeployar

EasyPanel → projeto **foxpro** → `confvision-worker` → **Redeploy** (pull GitHub branch `main`).

Repo workers: [github.com/adrianobraz/ConfVision](https://github.com/adrianobraz/ConfVision)

### `.env` — worker (`confvision-worker`)

```env
# API operacional = Go Core4 (NÃO o Xano legado)
XANO_BASE_URL=https://vision.confmonit2.com.br

# Eventos via API Go → Postgres central + dispatch terminal no Go
EVENT_STORE=postgres

# Worker NÃO avisa terminal (Go central faz)
TERMINAL_NOTIFY_ENABLED=false

# Demais (já existentes)
MEDIAMTX_RTSP_BASE=rtsp://foxpro_confvision:8554
RTMP_PUBLISH_SECRET=<igual Go e MediaMTX>
WORKER_ID=worker-docker-01
CONTABO_S3_ACCESS_KEY=...
CONTABO_S3_SECRET_KEY=...
CONTABO_S3_TENANT_ID=...
```

**Remover ou deixar vazio no worker:** `RECEPTOR_WEB_URL`, `RECEPTOR_WEB_SENHA`  
(não são mais usados com `TERMINAL_NOTIFY_ENABLED=false`)

`POSTGRES_URL` **não é obrigatório** no worker com `EVENT_STORE=postgres` (eventos vão pela API Go).

### Log worker esperado

```text
[CONFIG] OK | xano=https://vision.confmonit2.com.br | event_store=postgres
[CAPTURA] snapshot contabo ok evento=12345 url=...
```

**Não** deve aparecer `[TERMINAL]` no worker.

---

## 3. GitHub

### Repositório Core4 (monorepo local)

Contém Go + workers em `confvision/`:

```bash
cd C:\sistemaconfmonit\core4
git add home/confmonit/v4.0/confvision/src/modulos/visdata/
git add home/confmonit/v4.0/confvision/src/config/config.go
git add home/confmonit/v4.0/confvision/.env.example
git add confvision/event_sink.py confvision/config.py confvision/.env.example
git add confvision/VARIAVEIS_VPS.md confvision/DEPLOY_TERMINAL_DISPATCH.md
git commit -m "ConfVision: dispatch terminal no Go central; worker sem terminal_notify"
git push origin main
```

> Se `git remote` estiver vazio, configure:  
> `git remote add origin https://github.com/SEU_USUARIO/core4.git`

### Repositório workers (EasyPanel)

O EasyPanel puxa de `adrianobraz/ConfVision`. Sincronize a pasta `core4/confvision/` para lá (push manual ou mirror).

Arquivos-chave desta mudança no workers:

- `event_sink.py`
- `config.py`
- `.env.example`
- `DEPLOY_TERMINAL_DISPATCH.md`
- `VARIAVEIS_VPS.md`

---

## 4. Teste end-to-end

1. Franqueado com `UsuarioTerminal = S` no MySQL
2. Câmera analítica ativa, dispositivo armado
3. Provocar detecção (pessoa na área)
4. Verificar:
   - `vis_evento` no Postgres com `snapshot_url` preenchido
   - `id_evento` / `id_processo` preenchidos após dispatch
   - Processo **CV01** aberto no WebTerminal
5. Evento **sensor** não deve abrir CV01 (terminal já abre pelo alarme)

### Comandos úteis

```bash
# API Go
curl https://vision.confmonit2.com.br/vis_health

# Log Go
sudo journalctl -u confmonit4confvision -f | grep TERMINAL
```

---

## Fluxo final

```
Worker VPS
  → POST https://vision.confmonit2.com.br/vis_evento
  → PUT snapshot / POST vis_evento_finalizar
  → (para — sem terminal_notify)

Go central
  → goroutine após snapshot
  → checa UsuarioTerminal (MySQL)
  → POST receptorWeb /recebe-evento-confvision (CV01)
  → UPDATE vis_evento id_evento, id_processo
```

---

## Referências

- Deploy Go: [`../home/confmonit/v4.0/confvision/DEPLOY_PROXMOX.md`](../home/confmonit/v4.0/confvision/DEPLOY_PROXMOX.md)
- Deploy VPS: [`DEPLOY_EASYPANEL.md`](DEPLOY_EASYPANEL.md)
- Variáveis VPS: [`VARIAVEIS_VPS.md`](VARIAVEIS_VPS.md)
