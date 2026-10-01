# Deploy imagem.dnsid.com.br + integracao Moni

## 1. Migration Postgres (no servidor Core4 / rede 10.2.2.x)

```bash
cd /home/confmonit/v4.0/confvision   # ou clone core4
export POSTGRES_URL='postgres://confmonit:SENHA@10.2.2.120:5432/confmonit?sslmode=disable'
bash confvision/sql/apply_006_integracao.sh
```

Alternativa manual:

```bash
psql "$POSTGRES_URL" -f confvision/sql/006_integracao_schema.sql
```

## 2. Variaveis .env Go (ja preenchidas em home/confmonit/v4.0/confvision/.env)

Reinicie o servico apos editar:

```bash
sudo systemctl restart confmonit4confvision
curl -s http://127.0.0.1:8086/imagens/health
```

## 3. Cloudflare — imagem.dnsid.com.br

O Moni faz GET publico (sem auth). O subdominio aponta para o **mesmo servidor Apache** do ConfVision (`vision.confmonit2.com.br` → proxy `127.0.0.1:8086`).

### DNS (painel Cloudflare → zona dnsid.com.br)

| Tipo | Nome | Conteudo | Proxy |
|------|------|----------|-------|
| **A** | `imagem` | IP publico do Apache (ex.: `185.130.61.4`) | Proxied (nuvem laranja) **ou** DNS only |

Recomendado para comecar: **DNS only (cinza)** + Let's Encrypt no Apache (igual vision).

Se usar **Proxied (laranja)**:
- SSL/TLS → **Full (strict)** se tiver cert valido na origem
- SSL/TLS → **Full** se cert auto-assinado na origem
- Ative **Always Use HTTPS**

### Apache no servidor

```bash
sudo cp home/confmonit/hosts/imagem-le-ssl.conf /etc/apache2/sites-available/
sudo certbot certonly --apache -d imagem.dnsid.com.br
sudo a2ensite imagem-le-ssl.conf
sudo apache2ctl configtest && sudo systemctl reload apache2
```

Teste:

```bash
curl -I https://imagem.dnsid.com.br/imagens/health
# HTTP/2 200
```

URL de imagem (apos evento enviado ao Moni):

```text
https://imagem.dnsid.com.br/{hash12+}
```

Cadastre no Moni a URL base: `https://imagem.dnsid.com.br/`

## 4. ConfVision UI

1. Menu → **Integracao eventos** → cadastrar Moni (URL GerarEvento, Basic auth, empresa/evento/setor).
2. **Gerenciar Cliente** → campo **Codigo Moni** por cliente.
