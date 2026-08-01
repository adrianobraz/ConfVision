# Terminal Movel — deploy Proxmox / FileZilla

Terminal de atendimento movel (Go + assets).

## Nome do serviço

| Item | Valor |
|------|--------|
| **Serviço systemd** | `confmonit4terminalMovel` |
| **Script start** | `/home/confmonit/start/terminalMovel.sh` |
| **Pasta app** | `/home/confmonit/v4.0/terminalMovel` |
| **Binário** | `/home/confmonit/v4.0/terminalMovel/terminalMovel` |
| **Porta** | `2007` (`.env` → `PORTA=2007`) |

Arquivos no repositório:

- `home/confmonit/service/confmonit4terminalMovel.service`
- `home/confmonit/start/terminalMovel.sh`

---

## 1. O que subir via FileZilla

Destino no servidor: **`/home/confmonit/v4.0/terminalMovel/`**

**Produção (sem código-fonte Go — recomendado):**

```
terminalMovel/
├── terminalMovel   ← binário Linux (build.ps1 no PC)
├── .env
└── assets/         ← HTML, JS, CSS, bootstrap (obrigatório em runtime)
    ├── bootstrap/
    ├── css/
    ├── js/
    ├── pagina/
    └── modulos/
```

**Não subir para produção:** `src/`, `main.go`, `go.mod`, `go.sum`, `*.bak`

Opcional: omitir `assets/**/*.go` no servidor (já estão no binário).

---

## 2. Compilar no PC (Windows)

```powershell
cd C:\sistemaconfmonit\core4\home\confmonit\v4.0\terminalMovel
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

**Não use** `-Obfuscate` — quebra templates HTML.

Teste local Windows:

```powershell
powershell -ExecutionPolicy Bypass -File .\build.ps1 -Local
```

---

## 3. Serviço systemd (primeira vez)

```bash
sudo cp /home/confmonit/service/confmonit4terminalMovel.service /etc/systemd/system/
sudo chmod +x /home/confmonit/start/terminalMovel.sh
sudo systemctl daemon-reload
sudo systemctl enable confmonit4terminalMovel
sudo systemctl start confmonit4terminalMovel
sudo systemctl status confmonit4terminalMovel
```

---

## 4. Comandos do dia a dia

```bash
chmod +x /home/confmonit/v4.0/terminalMovel/terminalMovel
sudo systemctl restart confmonit4terminalMovel
sudo systemctl status confmonit4terminalMovel
sudo journalctl -u confmonit4terminalMovel -f
```

---

## 5. Quando precisa restart?

| Alteração | Restart? |
|-----------|----------|
| HTML / JS / CSS em `assets/` | **Não** — Ctrl+F5 |
| Código Go | **Sim** — `build.ps1` + subir binário + restart |
| `.env` | **Sim** — restart |

---

## 6. Fluxo rápido

1. **PC:** `powershell -ExecutionPolicy Bypass -File .\build.ps1`
2. **FileZilla:** subir `terminalMovel` (+ `assets/` / `.env` se necessário)
3. **SSH:**

```bash
chmod +x /home/confmonit/v4.0/terminalMovel/terminalMovel
sudo systemctl restart confmonit4terminalMovel
```
