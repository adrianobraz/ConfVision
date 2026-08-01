# Proxy Apache no SERVIDOR DA API (185.130.61.4)
# ================================================
#
# Objetivo: expor /v4 na porta 80 (Apache) apontando para a API Go na porta 2010.
# Motivo: a hospedagem KingHost (admConfmonit) bloqueia saida para :2010;
#         porta 80/443 normalmente e liberada.
#
# 1) Habilite os modulos (Debian/Ubuntu):
#      sudo a2enmod proxy proxy_http headers
#      sudo systemctl reload apache2
#
# 2) No VirtualHost da porta 80 (ou em conf-available), inclua:
#
# ---- cole a partir daqui ----

ProxyPreserveHost On
ProxyRequests Off

<Location /v4/>
    ProxyPass        http://127.0.0.1:2010/v4/
    ProxyPassReverse http://127.0.0.1:2010/v4/
    RequestHeader set X-Forwarded-Proto "http"
</Location>

# ---- ate aqui ----
#
# 3) Reinicie Apache:
#      sudo systemctl reload apache2
#
# 4) Teste NO servidor da API:
#      curl -X POST http://127.0.0.1/v4/notificacao/listar -H "Content-Type: application/json" -d "{}"
#      curl -X POST http://185.130.61.4/v4/notificacao/listar -H "Content-Type: application/json" -d "{}"
#
# Esperado: {"status":"Vazio"} ou {"status":"OK",...}
#
# 5) Na KingHost, teste o proxy PHP:
#      https://confhost.com.br/admconfmonit/api-v4-proxy.php?ping=1
#
# Windows (httpd.conf / httpd-vhosts.conf): mesmos blocos ProxyPass,
# com os modulos proxy_module e proxy_http_module carregados.
