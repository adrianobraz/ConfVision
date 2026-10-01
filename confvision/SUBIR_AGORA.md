# Subir agora — terminal CV01 via Go central

Tudo pronto. Só executar os passos abaixo.

---

## A) Proxmox — Go ConfVision (`185.130.61.4`)

### FileZilla → `/home/confmonit/v4.0/confvision/`

| Arquivo local | Destino |
|---------------|---------|
| `home\confmonit\v4.0\confvision\confvision` | `/home/confmonit/v4.0/confvision/confvision` |
| `home\confmonit\v4.0\confvision\.env` | `/home/confmonit/v4.0/confvision/.env` |

### SSH

```bash
chmod +x /home/confmonit/v4.0/confvision/confvision
sudo systemctl restart confmonit4confvision
sudo journalctl -u confmonit4confvision -f
```

Após detecção analítica, deve aparecer:

```text
[TERMINAL] ok evento=... idEvento=... processo=...
```

---

## B) EasyPanel — worker (`confvision-worker`)

1. Abra o serviço **confvision-worker** no projeto **foxpro**
2. Aba **Environment** → apague tudo → cole o conteúdo de **`confvision/.env`**
3. **Redeploy** (pull GitHub se o código já estiver lá, ou redeploy da imagem atual)

Log esperado no worker:

```text
[CONFIG] OK | xano=https://vision.confmonit2.com.br | event_store=postgres
```

**Não** deve aparecer `[TERMINAL]` no worker.

---

## C) GitHub (código — sem .env)

Os `.env` **não vão pro Git** (`.gitignore`). Só o código.

Repo workers EasyPanel: `adrianobraz/ConfVision`  
Repo core4: configure `git remote` e `git push origin main`

---

## Checklist rápido

- [ ] Binário `confvision` + `.env` no Proxmox
- [ ] `systemctl restart confmonit4confvision`
- [ ] Env do worker colado no EasyPanel
- [ ] Redeploy `confvision-worker`
- [ ] Teste detecção → CV01 no terminal
