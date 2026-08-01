# ConfService

Marketplace de monitoramento terceirizado: parceiro recebe eventos por webhook, franqueado escolhe o parceiro por cliente, ConfMonit fatura e fica com comissão.

**Local:** `home/confmonit/v4.0/confservice`  
**Stack:** Go API + React/TS + MySQL (prefixo `cs_`)

## Estrutura

```
confservice/
  api/     → Go (porta 2020)
  web/     → Portal do parceiro (Vite React)
  README.md
```

## 1. Banco

```bash
mysql -u root -p < api/sql/schema.sql
```

Ajuste `MYSQL_DSN` em `api/.env`.

## 2. API (dev)

```bash
cd api
copy .env.example .env   # Windows
# edite MYSQL_DSN, API_KEY, JWT_SECRET
go run .
```

Health: `GET http://localhost:2020/health`

## 2b. Build / deploy (servidor Linux)

```powershell
cd api
.\build.ps1 -Package
```

Gera `deploy/confservice/` com binário, `.env`, SQL, systemd e start script.

Subir para `/home/confmonit/v4.0/confservice/` e seguir `LEIA-ME.txt` do pacote.

```bash
systemctl enable confmonit4confservice
systemctl restart confmonit4confservice
curl http://127.0.0.1:2020/health
```

### Rotas principais

| Método | Path | Auth | Uso |
|--------|------|------|-----|
| POST | `/parceiro/registrar` | — | Cadastro portal |
| POST | `/parceiro/login` | — | Login portal |
| GET | `/parceiro/me` | Bearer | Dados do parceiro |
| PUT | `/parceiro/me/atualizar` | Bearer | Webhook / preço |
| GET | `/parceiro/me/eventos` | Bearer | Relatório de eventos/webhooks (prova de envio) |
| GET | `/internal/eventos` | X-Api-Key | Relatório admin (filtro `idParceiro`, `de`, `ate`, `status`) |
| GET | `/parceiros` | — | Legado (retorna vazio — use catálogo filtrado) |
| GET | `/internal/parceiros/global` | X-Api-Key | Pool global (admConfmonit) |
| GET | `/internal/parceiros/rep` | X-Api-Key | Parceiros liberados CEN→REP |
| POST | `/internal/parceiros/rep/salvar` | X-Api-Key | Salvar liberação CEN→REP |
| GET | `/internal/parceiros/franqueado` | X-Api-Key | Parceiros liberados REP→FQ |
| POST | `/internal/parceiros/franqueado/salvar` | X-Api-Key | Salvar liberação REP→FQ |
| GET | `/internal/parceiros/catalogo` | X-Api-Key | Catálogo filtrado Franqueado Pro |
| POST | `/internal/vinculo` | X-Api-Key | Vincular cliente → parceiro |
| POST | `/internal/vinculo/desativar` | X-Api-Key | Remover vínculo |
| GET | `/internal/vinculo/cliente` | X-Api-Key | Consultar vínculo |
| POST | `/internal/evento` | X-Api-Key | Enfileirar webhook (Xano/#22) |
| POST | `/internal/webhook/processar` | X-Api-Key | Processar fila manual |

### Exemplo enfileirar evento (Xano)

```http
POST /internal/evento
X-Api-Key: <API_KEY>
Content-Type: application/json

{
  "idFranqueado": "...",
  "idCliente": "...",
  "alarmEventsId": 123,
  "payload": {
    "grupo": "ALARME",
    "codigo": "E130",
    "descricao": "...",
    "dispositivo": "...",
    "dataHora": "..."
  }
}
```

O parceiro recebe `POST` na `webhook_url` com esse JSON + `Authorization: Bearer <webhook_token>`.

## 3. Portal / marketplace

```bash
cd web
copy .env.example .env
npm install
npm run dev
```

Abre em `http://localhost:5173`.

- `/` — marketplace (vitrine de serviços)
- `/cadastro` — profissional / empresa (+ comissão plataforma)
- `/cadastro-fabricante` — marcas (Intelbras, JFL…) + comissão disponível
- `/parceiro` — portal monitoramento (webhook / faturas)
- `/admin` — admin (default `admin` / `admin`) lista parceiros e reseta senha

### Marketplace — banco

```bash
mysql -u root -p confservice < api/sql/schema-marketplace.sql
```

Rotas públicas: `/marketplace/categorias`, `/marketplace/prestadores`, `/marketplace/fabricantes`,  
`POST /marketplace/prestador/registrar`, `POST /marketplace/fabricante/registrar`, `POST /marketplace/orcamento`.

## Integrações

### Xano `#22`
Consulta vínculo e, se existir, `POST /internal/evento` e interrompe WhatsApp/IA.  
Ajuste URL/key em `functions/22_...xs` (`$confServiceUrl` / `$confServiceKey`).

### Franqueado Pro
Menu **Comercial → Parceiro Monitoramento**.  
No `.env` do Franqueado Pro:

```
CONFSERVICE_URL=http://127.0.0.1:2020
CONFSERVICE_API_KEY=dev-confservice-api-key
```

Parceiro cadastrado no portal **não aparece** até a **Central** liberar para o **Representante** e o **Representante** liberar para o **Franqueado**.

SQL ACL (rodar uma vez no banco `confservice`):

```bash
mysql -u root -p confservice < api/sql/20260729_parceiro_acl.sql
```

### admConfmonit (via API V4 — sem Xano)

No `.env` da API V4, mesmas variáveis `CONFSERVICE_URL` e `CONFSERVICE_API_KEY`.

| Método | Path | Perfil | Uso |
|--------|------|--------|-----|
| POST | `/v4/confservice/parceiros/global` | CEN/REP | Lista pool global |
| POST | `/v4/confservice/parceiros-rep/listar` | CEN | `{ idRepresentante }` |
| POST | `/v4/confservice/parceiros-rep/salvar` | CEN | `{ idRepresentante, idsParceiro[] }` |
| POST | `/v4/confservice/parceiros-rep/disponiveis` | REP | Parceiros liberados pela central |
| POST | `/v4/confservice/parceiros-franqueado/listar` | REP | `{ idFranqueado }` |
| POST | `/v4/confservice/parceiros-franqueado/salvar` | REP | `{ idFranqueado, idsParceiro[] }` |

Autenticação: `Authorization: Bearer <token>` do login `/v4/admfinanceiro/logar`.

### Fechamento quinzenal + comissão

| Quem | Como |
|------|------|
| Admin break-glass (admConfmonit) | Menu **ConfService (fechamento)** → fecha **todos** |
| Parceiro (portal) | Botão **Fechar quinzena / gerar cobrança** → só **dele** |
| Cron | dias 1 e 16 → todos |

```http
POST /internal/fatura/fechar
X-Api-Key: <API_KEY>

{ "periodoInicio": "2026-07-01", "periodoFim": "2026-07-15" }
```

```http
POST /parceiro/me/fatura/fechar
Authorization: Bearer <token>
```

Sem body → fecha a quinzena atual (1–15 ou 16–fim do mês).  
Comissão: `COMISSAO_PADRAO_PCT` (default 20%) ou `cs_parceiro.comissao_pct`.  
Franqueado paga o bruto; parceiro recebe líquido (bruto − comissão).

Agendar no servidor (FileZilla **não** grava em `/etc/cron.d`):

1. Suba para `/home/confmonit/v4.0/confservice/`:
   - `cron/confservice-fatura`
   - `start/confservice-fatura-fechar.sh`
   - `start/install-confservice-fatura-cron.sh`

2. SSH:
```bash
chmod +x /home/confmonit/v4.0/confservice/start/install-confservice-fatura-cron.sh
sudo /home/confmonit/v4.0/confservice/start/install-confservice-fatura-cron.sh
/home/confmonit/start/confservice-fatura-fechar.sh   # teste manual
tail -20 /var/log/confservice-fatura.log
```

O install copia o cron para `/etc/cron.d/` e o script para `/home/confmonit/start/`.  
Roda às 00:05 nos dias **1** e **16**. A key vem do `.env` (`API_KEY`).

## Prefixo MySQL

Todas as tabelas usam `cs_`: `cs_parceiro`, `cs_cliente_vinculo`, `cs_webhook_fila`, `cs_fatura_franqueado`, `cs_fatura_parceiro`, `cs_fatura_item`.
