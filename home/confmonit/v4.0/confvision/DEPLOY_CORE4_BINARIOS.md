# Deploy binarios no core-4 (sem fontes Go no servidor)

O core-4 roda apenas os executaveis em `/home/confmonit/v4.0/*/`.  
Nao ha `go.mod` nem `src/` no servidor — **compile na maquina de dev** e copie.

## Build (dev — Git Bash / WSL / Linux)

```bash
cd home/confmonit/v4.0
bash scripts/build-core4-linux.sh
```

Ou manual:

```bash
cd confvision
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o deploy-core4-confvision .

cd ../webTerminal
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o deploy-core4-terminal-linux .
```

## Copiar para core-4

```bash
scp confvision/deploy-core4-confvision root@CORE4:/home/confmonit/v4.0/confvision/confvision
scp webTerminal/deploy-core4-terminal-linux root@CORE4:/home/confmonit/v4.0/webTerminal/terminal
scp webTerminal/assets/js/confvision.js root@CORE4:/home/confmonit/v4.0/webTerminal/assets/js/confvision.js
```

## Reiniciar

```bash
ssh root@CORE4 'systemctl restart confmonit4confvision confmonit4webTerminal'
```

## Validar

```bash
# ConfVision: particao 01 e 1 devem funcionar (apos binario novo)
curl -s -H "Authorization: Bearer SEU_TOKEN" \
  "https://vision.confmonit2.com.br/vis_evento_ultimos25_setor?id_dispositivo=ID&particao=01&zonauser=001" | head -c 120

# Resposta deve conter "camera" e opcionalmente "hls_url" / "stream_path"
```

## WebTerminal — fallback particao

Mesmo com ConfVision antigo, o **webTerminal novo** tenta variantes `01` → `1` ao chamar a API.  
Deploy do binario `terminal` + Ctrl+F5 no browser.

## Integracoes (Moni / ConfMonit / Nenhum)

O envio de eventos (terminal CV01, Moni, etc.) depende de **Integracao → Sistema + Ativa** em `vis_integracao_config`.

- **Nenhum** (ou sem registro): nenhum envio externo.
- **ConfMonit** + ativa: terminal via CV01.
- **Moni** + ativa: API Moni (campos proprios em `integracao_moni.go`).

`TERMINAL_NOTIFY_ENABLED` so habilita o servidor receptor; nao dispara terminal sem integracao ConfMonit ativa.
