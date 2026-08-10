# ConfVision — deploy Proxmox / FileZilla

Painel web Go (login franqueado + câmeras + eventos + ao vivo HLS).

## Nome do serviço (registrar)

| Item | Valor |
|------|--------|
| **Serviço systemd** | `confmonit4confvision` |
| **Script start** | `/home/confmonit/start/confvision.sh` |
| **Pasta app** | `/home/confmonit/v4.0/confvision` |
| **Binário** | `/home/confmonit/v4.0/confvision/confvision` |
| **Porta** | `8086` (`.env` → `PORTA=8086`) |
| **URL** | `http://185.130.61.4:8086/login` |

Arquivos no repositório:

- `home/confmonit/service/confmonit4confvision.service`
- `home/confmonit/start/confvision.sh`

---

## 1. O que subir via FileZilla

Destino no servidor: **`/home/confmonit/v4.0/confvision/`**

**Produção (sem código-fonte Go — recomendado):**

```
confvision/
├── confvision    ← binário Linux (compilado no PC com build.ps1)
├── .env
└── recursos/     ← HTML, JS, CSS (obrigatório em runtime)
```

**Não subir para produção:** `src/`, `app.go`, `go.mod`, `go.sum`, `confvision.exe`

Também subir (primeira vez ou se alterou):

```
/home/confmonit/start/confvision.sh
/etc/systemd/system/confmonit4confvision.service   ← ver seção 3
```

---

## 2. Compilar no PC (Windows)

Na pasta do projeto:

```powershell
cd C:\sistemaconfmonit\core4\home\confmonit\v4.0\confvision

# Build endurecido (padrão)
.\build.ps1

# Build ofuscado com Garble (instalar antes: go install mvdan.cc/garble@latest)
.\build.ps1 -Obfuscate

# Teste local no Windows
.\build.ps1 -Local
```

Gera o binário **`confvision`** (Linux). Suba só esse arquivo via FileZilla.

Teste manual no servidor (opcional):

```bash
cd /home/confmonit/v4.0/confvision
chmod +x confvision
./confvision
# Ctrl+C — deve aparecer: ConfVision rodando em HTTP na porta 8086
```

---

## 3. Criar serviço systemd (primeira vez)

```bash
# Copiar unit do repo (se ainda não estiver em /etc)
sudo cp /home/confmonit/service/confmonit4confvision.service /etc/systemd/system/

# Ou criar direto:
sudo nano /etc/systemd/system/confmonit4confvision.service
# (cole o conteúdo de home/confmonit/service/confmonit4confvision.service)

sudo systemctl daemon-reload
sudo systemctl enable confmonit4confvision
sudo systemctl start confmonit4confvision
sudo systemctl status confmonit4confvision
```

---

## 4. Comandos do dia a dia

```bash
# Reiniciar (após recompilar ou alterar .env)
sudo systemctl restart confmonit4confvision

# Status
sudo systemctl status confmonit4confvision

# Logs ao vivo
sudo journalctl -u confmonit4confvision -f

# Parar / iniciar
sudo systemctl stop confmonit4confvision
sudo systemctl start confmonit4confvision
```

---

## 5. Firewall / Proxmox

Liberar porta **8086** TCP na VM (ufw ou painel Proxmox):

```bash
sudo ufw allow 8086/tcp
sudo ufw reload
```

---

## 6. Quando precisa restart?

| Alteração | Restart? |
|-----------|----------|
| HTML / JS / CSS em `recursos/` | **Não** — Ctrl+F5 no browser |
| Código Go (`src/`, `app.go`) | **Sim** — `build.ps1` + subir binário + `systemctl restart` |
| `.env` | **Sim** — `systemctl restart` |
| APIs Xano | Push no Xano — não afeta o painel Go |

---

## 7. Fluxo rápido (build no PC + FileZilla + SSH)

1. **PC:** `.\build.ps1` ou `.\build.ps1 -Obfuscate`
2. **FileZilla:** subir `confvision` (e `recursos/` se alterou frontend)
3. **SSH:**

```bash
chmod +x /home/confmonit/v4.0/confvision/confvision
sudo systemctl restart confmonit4confvision
```

4. Abrir: `http://185.130.61.4:8086/login`

---

## 8. Variáveis `.env` (produção)

Ver template completo: [`.env.producao.example`](.env.producao.example)

**Novidade — dispatch terminal CV01 (analítico):**

```env
RECEPTOR_WEB_URL=http://185.130.61.3:5000
RECEPTOR_WEB_SENHA=<SenhaWeb receptorWeb>
TERMINAL_NOTIFY_ENABLED=true
POSTGRES_URL=postgres://confmonit:SENHA@191.96.156.116:5432/confmonit?sslmode=disable
BD_HOST_MV4=...
```

Deploy terminal: [`../../confvision/DEPLOY_TERMINAL_DISPATCH.md`](../../confvision/DEPLOY_TERMINAL_DISPATCH.md)

URLs de câmera no painel: **`live/{id}`** (não grava no banco).
