# Infra central no Proxmox (core-4)

Postgres e Redis centralizados no Proxmox. Workers (EasyPanel / srvN) usam **API Go** + **Redis remoto**.

## Postgres — já no Proxmox

O Go ConfVision **já usa** rede privada:

```env
POSTGRES_URL=postgres://confmonit:SENHA@10.2.2.120:5432/confmonit?sslmode=disable
```

Arquivo: `/home/confmonit/v4.0/confvision/.env`

Verificar:

```bash
curl -s https://vision.confmonit2.com.br/vis_health
# {"enabled":true,"postgres":"ok","status":"ok"}

psql "$POSTGRES_URL" -c "SELECT COUNT(*) FROM vis_camera;"
```

**Pode desligar** `foxpro/visionpsql` na VPS após:

1. Confirmar contagem de câmeras/licenças no Proxmox
2. Atualizar EasyPanel (sensor sem POSTGRES_URL da VPS)
3. Redeploy workers

> O Postgres da VPS (`191.96.158.116`) era cópia/migração — **não** é usado pelo Go em produção.

---

## Redis — Memurai no Proxmox (`185.130.61.5`)

Redis central já roda como **Memurai** em `185.130.61.5:6379` (não usar `foxpro/visionredis` na VPS).

Teste:

```bash
memurai-cli -h 185.130.61.5 -p 6379 -a "2008.03.28.Dri.zin" ping
# PONG
```

### 1. Firewall

Libere **6379/tcp** apenas para IPs confiáveis (VPS EasyPanel, srv1, srv2…):

```bash
ufw allow from IP_DA_VPS to any port 6379 proto tcp comment 'confvision-memurai'
ufw reload
```

### 2. REDIS_URL nos workers (EasyPanel)

Use o IP do Memurai (`185.130.61.5`):

```env
REDIS_URL=redis://default:2008.03.28.Dri.zin@185.130.61.5:6379/0
CONFIG_CACHE_BACKEND=redis
```

Arquivos locais: `confvision/easypanel/*.env` (já atualizados).

Redeploy: sync-agent, mediamtx, worker, dvr, motion, timelapse.

Log esperado sync-agent:

```text
[SYNC-AGENT] rtmp_auth cache=N ttl=600s node=1
```

### 3. Desligar VPS

Após Redis OK e sync funcionando:

- Parar `foxpro/visionredis`
- Parar `foxpro/visionpsql`

---

## Topologia

```
core-4 (Proxmox)
  ├── Go API     → POSTGRES 10.2.2.120
  └── Redis      → :6379 (185.130.61.5 Memurai)

VPS / srv1
  ├── MediaMTX + Guard
  ├── sync-agent → HTTPS vision.confmonit2.com.br + Redis Memurai
  └── workers    → idem
```
