# C2 — Monitor rust-pilot (cron)

## Infraestrutura

| Item | Valor |
|------|--------|
| Proxmox node | `server14525` |
| Container | **CT 111** — `integracao-edge` |
| IP público | `185.130.61.6` |
| Hostname SSH | `ReceptorTeste` |

## Monitoramento

| Path | Descrição |
|------|-----------|
| `/etc/cron.d/confvision-rust-pilot` | Cron */5 |
| `/var/log/rust-pilot-verify.log` | Log |
| `/opt/confvision/ConfVision` | Git `rust-pilot` |
| URL | `https://foxpro-rust-pilot.rkr351.easypanel.host` |

Instalado: **2026-09-26**. Status Fase C: **OK** (verify manual OK).

```bash
tail -50 /var/log/rust-pilot-verify.log
cd /opt/confvision/ConfVision && git pull origin rust-pilot
bash confvision-rust-processor/scripts/phase-c-verify.sh --strict-c3 \
  https://foxpro-rust-pilot.rkr351.easypanel.host
```
