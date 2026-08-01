# WebAmbiente — deploy Proxmox / FileZilla

Painel web Go (mapas, ambientes, monitor, editor de planta).

## Nome do serviço

| Item | Valor |
|------|--------|
| **Serviço systemd** | `confmonit4webAmbiente` |
| **Script start** | `/home/confmonit/start/webAmbiente.sh` |
| **Pasta app** | `/home/confmonit/v4.0/webAmbiente` |
| **Binário** | `/home/confmonit/v4.0/webAmbiente/webAmbiente` |
| **Porta** | `8085` (`.env` → `PORTA=8085`) |

Arquivos no repositório:

- `home/confmonit/service/confmonit4webAmbiente.service`
- `home/confmonit/start/webAmbiente.sh`

---

## 1. O que subir via FileZilla

Destino no servidor: **`/home/confmonit/v4.0/webAmbiente/`**

**Produção (sem código-fonte Go — recomendado):**

```
webAmbiente/
├── webAmbiente   ← binário Linux (compilado no PC com build.ps1)
├── .env
└── public/       ← HTML, JS, CSS, imagens (obrigatório em runtime)
    ├── css/
    ├── js/
    ├── img/      ← ex.: capaCentral2.jpg (login)
    └── templates/
```

**Não subir para produção:** `src/`, `app.go`, `go.mod`, `go.sum`, `scripts/`, `webAmbiente.exe`

`scripts/` (ex.: policy JSON Contabo) é só referência local — não é usado em runtime.

---

## 2. Compilar no PC (Windows)

```powershell
cd C:\sistemaconfmonit\core4\home\confmonit\v4.0\webAmbiente

# Build endurecido (padrão)
.\build.ps1

# Build ofuscado com Garble
.\build.ps1 -Obfuscate

# Teste local no Windows
.\build.ps1 -Local
```

Gera o binário **`webAmbiente`** (Linux). Suba só esse arquivo quando alterar backend Go.

---

## 3. Serviço systemd (primeira vez)

```bash
sudo cp /home/confmonit/service/confmonit4webAmbiente.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable confmonit4webAmbiente
sudo systemctl start confmonit4webAmbiente
sudo systemctl status confmonit4webAmbiente
```

---

## 4. Comandos do dia a dia

```bash
sudo systemctl restart confmonit4webAmbiente
sudo systemctl status confmonit4webAmbiente
sudo journalctl -u confmonit4webAmbiente -f
```

---

## 5. Quando precisa restart?

| Alteração | Restart? |
|-----------|----------|
| HTML / JS / CSS em `public/` | **Não** — Ctrl+F5 no browser |
| Código Go (`src/`, `app.go`) | **Sim** — `build.ps1` + subir binário + restart |
| `.env` | **Sim** — `systemctl restart` |

---

## 6. Fluxo rápido

1. **PC:** `.\build.ps1` ou `.\build.ps1 -Obfuscate`
2. **FileZilla:** subir `webAmbiente` (e `public/` se alterou frontend)
3. **SSH:**

```bash
chmod +x /home/confmonit/v4.0/webAmbiente/webAmbiente
sudo systemctl restart confmonit4webAmbiente
```

---

## 7. Por que `public/` é obrigatório?

O binário lê templates do disco em cada requisição e serve estáticos de `public/`:

- `template.ParseFiles("public/templates/...")` nos módulos Go
- `http.FileServer(http.Dir("public"))` em `/public/`

Sem a pasta `public/`, páginas quebram e assets (CSS/JS/img) não carregam.
