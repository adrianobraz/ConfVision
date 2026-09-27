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
| URL A | `https://foxpro-rust-pilot.rkr351.easypanel.host` |
| URL B | `https://foxpro-rust-pilot-b.rkr351.easypanel.host` |
| Sidecar | `https://foxpro-rust-yolo-sidecar.rkr351.easypanel.host` |

Instalado: **2026-09-26**. **D6 (2026-09-27):** usar `phase-d6-verify.sh` (A + B + sidecar + Go D5 opcional).

```bash
tail -50 /var/log/rust-pilot-verify.log
tail -50 /var/log/foxpro-d6-verify.log
cd /opt/confvision/ConfVision && git pull origin rust-pilot
bash confvision-rust-processor/scripts/phase-d6-verify.sh
```

Cron exemplo: [`deploy/observability/cron-d6-foxpro.example`](../deploy/observability/cron-d6-foxpro.example).
