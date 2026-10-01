# RTMP — auto-ban de IP (NVR) e operação

## Sintoma

Logs MediaMTX/Guard:

```text
failed to authenticate ... "motivo": "ip_banido"
[RTMP-GUARD] IP BANIDO
```

Enquanto o IP estiver banido, **todos** os canais daquele NVR/DVR param (403 antes de validar `cam/{hash}`).

## Causa típica

1. Vários canais no **mesmo IP público** com path errado (`live/1`, hash typo, URL antiga).
2. Limiar antigo: **3 falhas / 60 s no IP inteiro** → ban **3600 s**.
3. Câmeras **inativas** no cadastro (`camera_inativa`) — não banem, mas não publicam.

## Correção imediata (produção atual)

1. **Desbanir IPs** no painel ConfVision (RTMP → bans) ou `POST /api/rtmp-bans/unban` com `{"ip":"x.x.x.x"}`.
2. Opcional: esvaziar bans no volume (`/opt/confvision/recordings/rtmp_bans.json` → `{"bans":[]}`) e reiniciar serviço **confvision** (só se souber o impacto).
3. Corrigir URLs nos aparelhos: `rtmp://…/cam/{hash12}` exatamente como no cadastro.
4. Ativar câmeras que devem transmitir (`camera_inativa`).

## Correção no código (após deploy)

A partir desta versão:

- `path_invalido`, `chave_invalida`, `camera_nao_encontrada`: contagem **por path**, não soma todos os canais do NVR.
- Default `RTMP_BAN_MAX_FAILS=12` (ajustável no EasyPanel).

**Deploy:** rebuild imagem `Dockerfile.mediamtx`, redeploy serviço **confvision**, depois **unban** dos IPs ainda listados em `rtmp_bans.json` (TTL 1 h ou manual).

## Env úteis

| Variável | Default | Uso |
|----------|---------|-----|
| `RTMP_BAN_MAX_FAILS` | `12` | Falhas hard na janela (por path ou `*`) |
| `RTMP_BAN_WINDOW_SEC` | `60` | Janela |
| `RTMP_BAN_TTL_SEC` | `3600` | Duração do ban |
| `RTMP_BAN_AUTO_UNBAN` | `1` | Publish OK remove auto-ban (não manual) |
