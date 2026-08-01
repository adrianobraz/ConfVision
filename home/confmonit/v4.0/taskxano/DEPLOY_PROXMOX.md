# TaskXano — deploy Proxmox / FileZilla

Worker Go que substitui tasks do Xano (WhatsApp, fila ligacao, autofim, arme/desarme).

**Sem frontend** — so binario + `.env`.

## Nome do servico

| Item | Valor |
|------|--------|
| **Servico systemd** | `confmonit4taskxano` |
| **Script start** | `/home/confmonit/start/taskxano.sh` |
| **Pasta app** | `/home/confmonit/v4.0/taskxano` |
| **Binario** | `/home/confmonit/v4.0/taskxano/taskxano` |
| **Notify HTTP** | `:8081` (`.env` → `NOTIFY_LISTEN_ADDR`) |

Arquivos no repositorio:

- `home/confmonit/service/confmonit4taskxano.service`
- `home/confmonit/start/taskxano.sh`

---

## 1. O que subir via FileZilla

Destino no servidor: **`/home/confmonit/v4.0/taskxano/`**

**Producao (sem codigo-fonte Go):**

```
taskxano/
├── taskxano    ← binario Linux (build.ps1 no PC)
└── .env        ← configuracao (URLs Xano, intervalos, chaves)
```

**Nao subir:** `main.go`, `*.go`, `autofim/`, `go.mod`, `*.bak`, `build.ps1`

Este app **nao tem** pasta `assets/`, `public/` ou `recursos/`.

---

## 2. Compilar no PC (Windows)

```powershell
cd C:\sistemaconfmonit\core4\home\confmonit\v4.0\taskxano
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

Teste local:

```powershell
powershell -ExecutionPolicy Bypass -File .\build.ps1 -Local
```

---

## 3. Servico systemd (primeira vez)

```bash
sudo cp /home/confmonit/service/confmonit4taskxano.service /etc/systemd/system/
sudo chmod +x /home/confmonit/start/taskxano.sh
sudo systemctl daemon-reload
sudo systemctl enable confmonit4taskxano
sudo systemctl start confmonit4taskxano
sudo systemctl status confmonit4taskxano
```

---

## 4. Comandos do dia a dia

```bash
chmod +x /home/confmonit/v4.0/taskxano/taskxano
sudo systemctl restart confmonit4taskxano
sudo systemctl status confmonit4taskxano
sudo journalctl -u confmonit4taskxano -f
```

Health check do notify server:

```bash
curl -s http://127.0.0.1:8081/healthz
```

---

## 5. Quando precisa restart?

| Alteracao | Restart? |
|-----------|----------|
| Codigo Go | Sim — build + subir binario + restart |
| `.env` | Sim — restart |

---

## 6. Fluxo rapido

1. **PC:** `powershell -ExecutionPolicy Bypass -File .\build.ps1`
2. **FileZilla:** subir `taskxano` (+ `.env` se mudou config)
3. **SSH:**

```bash
chmod +x /home/confmonit/v4.0/taskxano/taskxano
sudo systemctl restart confmonit4taskxano
```
