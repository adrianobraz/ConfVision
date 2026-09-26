# Postgres — liberar IP do seu PC (porta 5432)

Banco **`confmonit`**, tabela **`vis_camera`** (Fase A): host documentado em [`VARIAVEIS_VPS.md`](../../VARIAVEIS_VPS.md).

| Destino | Host público | Porta |
|---------|--------------|-------|
| Postgres central (Proxmox/NAT) | `191.96.156.116` | `5432` |
| VPS EasyPanel foxpro (se usar SQL via túnel) | `31.97.173.119` | SSH `22` |

O Rust **não** conecta no Postgres — só a **API Go**. Você libera IP para **DBeaver / psql / pgAdmin** no seu computador.

---

## 1) Descobrir seu IP público

No PC:

```powershell
(Invoke-WebRequest -Uri "https://ifconfig.me/ip" -UseBasicParsing).Content.Trim()
```

Anote: `SEU_IP_PUBLICO/32`.

---

## 2) Liberar no servidor Postgres (recomendado)

SSH no **host onde o Postgres escuta** (ex.: Proxmox/Core4 ou máquina do `191.96.156.116`):

```bash
# Copiar script do repo e executar (substitua o IP):
bash postgres-allow-client-ip.sh SEU_IP_PUBLICO
```

Script: [`../scripts/postgres-allow-client-ip.sh`](../scripts/postgres-allow-client-ip.sh)

Manual (UFW):

```bash
ufw allow from SEU_IP_PUBLICO to any port 5432 proto tcp comment 'confmonit psql admin'
ufw status numbered
```

**Também liberar a VPS foxpro** (worker/ops na mesma rede que já acessa DB):

```bash
ufw allow from 31.97.173.119 to any port 5432 proto tcp comment 'foxpro easypanel'
```

### `pg_hba.conf` (se conectar e der “no pg_hba.conf entry”)

No servidor Postgres, incluir (ajuste paths da instalação):

```text
host    confmonit    confmonit    SEU_IP_PUBLICO/32    scram-sha-256
```

Depois:

```bash
sudo systemctl reload postgresql
# ou: docker exec visionpsql psql -U postgres -c "SELECT pg_reload_conf();"
```

---

## 3) Sem abrir IP — túnel SSH pela VPS foxpro

Se **não** quiser expor 5432 na internet, no **seu PC**:

```powershell
ssh -L 15432:191.96.156.116:5432 root@31.97.173.119
```

Deixe a sessão aberta. Conecte o cliente em:

- Host: `127.0.0.1`
- Porta: `15432`
- Database: `confmonit`
- User: `confmonit`

---

## 4) Rodar SQL da Fase A

Arquivo: [`../sql/piloto_fase_a_isolamento.sql`](../sql/piloto_fase_a_isolamento.sql)

```bash
psql "postgres://confmonit:SENHA@127.0.0.1:15432/confmonit?sslmode=disable" -f piloto_fase_a_isolamento.sql
```

---

## 5) Teste rápido

```bash
psql "postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable" -c "SELECT id, worker_id FROM vis_camera WHERE id IN (3,4,5);"
```

Timeout → firewall/IP. `password authentication failed` → senha. `no pg_hba.conf entry` → passo 2 pg_hba.

---

## Segurança

- Prefira **IP /32** (só seu IP), não `0.0.0.0/0`.
- Remova regras antigas: `ufw delete N`.
- Não commitar senha no Git.
