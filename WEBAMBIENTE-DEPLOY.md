# webAmbiente — deploy e restart

Guia rápido para subir alterações no servidor e reiniciar o serviço.

## Onde fica no projeto (local — Windows)

| O quê | Caminho no repositório |
|-------|-------------------------|
| Código fonte Go | `home/confmonit/v4.0/webAmbiente/src/` |
| HTML / JS / CSS | `home/confmonit/v4.0/webAmbiente/public/` |
| Binário compilado | `home/confmonit/v4.0/webAmbiente/webAmbiente` (gerado no Linux) |
| Configuração (.env) | `home/confmonit/v4.0/webAmbiente/.env` |
| Script de start | `home/confmonit/start/webAmbiente.sh` |
| Serviço systemd (cópia no repo) | `home/confmonit/service/confmonit4webAmbiente.service` |

## Onde fica no servidor (Linux)

| O quê | Caminho |
|-------|---------|
| Aplicação | `/home/confmonit/v4.0/webAmbiente/` |
| Binário | `/home/confmonit/v4.0/webAmbiente/webAmbiente` |
| Script de start | `/home/confmonit/start/webAmbiente.sh` |
| Serviço systemd | `/etc/systemd/system/confmonit4webAmbiente.service` |
| URL | `http://185.130.61.4:8085` (porta definida em `.env` → `PORTA=8085`) |

## Serviço systemd

Nome do serviço: **`confmonit4webAmbiente`**

Definição no repositório:

```
home/confmonit/service/confmonit4webAmbiente.service
```

Conteúdo resumido: executa `/home/confmonit/start/webAmbiente.sh`, que faz `cd` em `/home/confmonit/v4.0/webAmbiente` e roda `./webAmbiente`.

### Comandos no servidor (SSH)

```bash
# Reiniciar após alterar binário Go ou .env
sudo systemctl restart confmonit4webAmbiente

# Ver se está rodando
sudo systemctl status confmonit4webAmbiente

# Ver logs em tempo real
sudo journalctl -u confmonit4webAmbiente -f

# Parar / iniciar
sudo systemctl stop confmonit4webAmbiente
sudo systemctl start confmonit4webAmbiente

# Recarregar systemd (só se alterou o arquivo .service)
sudo systemctl daemon-reload
sudo systemctl enable confmonit4webAmbiente
```

## Quando precisa de restart?

| Tipo de alteração | Arquivos | Precisa `systemctl restart`? |
|-------------------|----------|------------------------------|
| HTML, JS, CSS | `public/**` | **Não** — basta subir o arquivo |
| Templates Go embutidos | `public/templates/**` | **Não** (servidos como arquivo estático) |
| Código Go | `src/**`, `app.go` | **Sim** — recompilar binário + restart |
| Configuração | `.env` | **Sim** — restart |
| Xano (APIs/tabelas) | `apis/`, `tables/` | Push no Xano — **não** afeta o restart do webAmbiente |

Após subir só CSS/JS/HTML: use **Ctrl+F5** no navegador para limpar cache.

## Fluxo: alterou aqui → subir no servidor

### 1. Apenas front-end (HTML / JS / CSS)

Subir os arquivos alterados para o mesmo caminho relativo em:

```
/home/confmonit/v4.0/webAmbiente/public/
```

Exemplo (SCP a partir do PC):

```bash
scp home/confmonit/v4.0/webAmbiente/public/css/mapa.css \
  root@185.130.61.4:/home/confmonit/v4.0/webAmbiente/public/css/
```

### 2. Alterou código Go

No **servidor Linux** (Go não compila no Windows para Linux sem cross-compile):

```bash
cd /home/confmonit/v4.0/webAmbiente
go build -o webAmbiente .
sudo systemctl restart confmonit4webAmbiente
sudo systemctl status confmonit4webAmbiente
```

Ou: compilar localmente no servidor após enviar os `.go` via SCP/rsync.

### 3. Baixar do servidor → colocar no projeto local

Para trazer o que está no servidor para esta pasta do repositório:

```bash
# Um arquivo
scp root@185.130.61.4:/home/confmonit/v4.0/webAmbiente/public/css/mapa.css \
  home/confmonit/v4.0/webAmbiente/public/css/

# Pasta public inteira (cuidado: sobrescreve local)
scp -r root@185.130.61.4:/home/confmonit/v4.0/webAmbiente/public/ \
  home/confmonit/v4.0/webAmbiente/

# Binário + .env (referência — .env não versionar com senhas)
scp root@185.130.61.4:/home/confmonit/v4.0/webAmbiente/webAmbiente \
  home/confmonit/v4.0/webAmbiente/
```

No Windows (PowerShell), use WinSCP, FileZilla ou `scp` do OpenSSH se instalado.

## Conferir que salvou / está no ar

1. **Arquivo no servidor:** `ls -la /home/confmonit/v4.0/webAmbiente/public/caminho/do/arquivo`
2. **Serviço ativo:** `systemctl status confmonit4webAmbiente` → `active (running)`
3. **Site responde:** abrir `http://185.130.61.4:8085` ou `curl -I http://localhost:8085`
4. **Browser:** Ctrl+F5 após deploy de CSS/JS

## Xano (backend mapa)

Alterações em `apis/mapa_ambiente/` e `tables/` vão para o Xano (não para o servidor webAmbiente):

- Push via extensão Xano / ferramenta do projeto
- Endpoint principal: `POST /mapa_ambiente_query_all`

## Resumo de um comando

```bash
# Depois de subir binário ou .env no servidor:
sudo systemctl restart confmonit4webAmbiente
```
