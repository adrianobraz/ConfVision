package apache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const sitesEnabledDir = "/etc/apache2/sites-enabled"

type Result struct {
	OK      bool   `json:"ok"`
	SSLOK   bool   `json:"ssl_ok"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Slug    string `json:"slug"`
}

type Manager struct {
	SitesDir     string
	ProxyTarget  string
	CertbotEmail string
}

func SlugFromFQDN(fqdn string) string {
	s := strings.ToLower(strings.TrimSpace(fqdn))
	s = strings.ReplaceAll(s, ".", "-")
	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "")
	if s == "" {
		return "fp-invalid"
	}
	return "fp-" + s
}

func (m *Manager) Provisionar(fqdn string) Result {
	fqdn = strings.ToLower(strings.TrimSpace(fqdn))
	if fqdn == "" || !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`).MatchString(fqdn) {
		return Result{OK: false, Status: "erro", Message: "FQDN invalido"}
	}

	slug := SlugFromFQDN(fqdn)
	if err := os.MkdirAll("/var/www/html/.well-known/acme-challenge", 0755); err != nil {
		return Result{OK: false, Status: "erro", Message: "mkdir acme-challenge: " + err.Error(), Slug: slug}
	}

	// Evita vhosts duplicados/antigos do mesmo FQDN (causa "vhost ambiguity" no Certbot)
	m.limparVhostsRelacionados(fqdn, slug)

	if err := m.escreverVhostHTTP(fqdn, slug, false); err != nil {
		return Result{OK: false, Status: "erro", Message: err.Error(), Slug: slug}
	}
	if out, err := run("a2ensite", slug+".conf"); err != nil {
		return Result{OK: false, Status: "erro", Message: "a2ensite http: " + out, Slug: slug}
	}
	if out, err := run("apache2ctl", "configtest"); err != nil {
		return Result{OK: false, Status: "erro", Message: "configtest: " + out, Slug: slug}
	}
	if out, err := run("systemctl", "reload", "apache2"); err != nil {
		return Result{OK: false, Status: "erro", Message: "reload apache: " + out, Slug: slug}
	}

	certOut := m.emitirCertificadoWebroot(fqdn)
	if chain, key, ok := certificadoLetsEncrypt(fqdn); ok {
		if err := m.ativarHTTPSComCert(fqdn, slug, chain, key); err != nil {
			return Result{
				OK:      true,
				SSLOK:   false,
				Status:  "pendente_dns",
				Message: "Certificado emitido, falha ao ativar HTTPS: " + err.Error(),
				Slug:    slug,
			}
		}
		return Result{
			OK:      true,
			SSLOK:   true,
			Status:  "ativo",
			Message: "Dominio provisionado com SSL",
			Slug:    slug,
		}
	}

	return Result{
		OK:      true,
		SSLOK:   false,
		Status:  "pendente_dns",
		Message: "Vhost HTTP criado. SSL pendente (configure DNS tipo A e aguarde propagacao): " + certOut,
		Slug:    slug,
	}
}

func (m *Manager) escreverVhostHTTP(fqdn, slug string, redirectHTTPS bool) error {
	httpFile := filepath.Join(m.SitesDir, slug+".conf")
	redirect := ""
	if redirectHTTPS {
		redirect = `
    RewriteEngine on
    RewriteCond %{REQUEST_URI} !^/\.well-known/acme-challenge/
    RewriteCond %{HTTPS} !=on
    RewriteRule ^ https://%{SERVER_NAME}%{REQUEST_URI} [END,NE,R=permanent]
`
	} else {
		// Enquanto SSL nao estiver ativo, HTTP ja faz proxy para o app
		redirect = fmt.Sprintf(`
    ProxyPreserveHost On
    ProxyPass /.well-known/acme-challenge/ !
    ProxyPass / %s
    ProxyPassReverse / %s
`, m.ProxyTarget, m.ProxyTarget)
	}

	httpConf := fmt.Sprintf(`<VirtualHost *:80>
    ServerName %s

    Alias /.well-known/acme-challenge/ /var/www/html/.well-known/acme-challenge/
    <Directory /var/www/html/.well-known/acme-challenge/>
        Options None
        AllowOverride None
        Require all granted
    </Directory>
%s
    ErrorLog ${APACHE_LOG_DIR}/%s_error.log
    CustomLog ${APACHE_LOG_DIR}/%s_access.log combined
</VirtualHost>
`, fqdn, redirect, slug, slug)

	if err := os.WriteFile(httpFile, []byte(httpConf), 0644); err != nil {
		return fmt.Errorf("gravar http vhost: %w", err)
	}
	return nil
}

func (m *Manager) escreverVhostSSL(fqdn, slug, certChain, certKey string) error {
	sslFile := filepath.Join(m.SitesDir, slug+"-le-ssl.conf")
	sslInclude := ""
	if arquivoExiste("/etc/letsencrypt/options-ssl-apache.conf") {
		sslInclude = "    Include /etc/letsencrypt/options-ssl-apache.conf\n"
	}

	sslConf := fmt.Sprintf(`<IfModule mod_ssl.c>
<VirtualHost *:443>
    ServerName %s

    SSLEngine on
    SSLCertificateFile %s
    SSLCertificateKeyFile %s
%s
    ProxyPreserveHost On
    ProxyPass / %s
    ProxyPassReverse / %s

    ErrorLog ${APACHE_LOG_DIR}/%s_ssl_error.log
    CustomLog ${APACHE_LOG_DIR}/%s_ssl_access.log combined
</VirtualHost>
</IfModule>
`, fqdn, certChain, certKey, sslInclude, m.ProxyTarget, m.ProxyTarget, slug, slug)

	if err := os.WriteFile(sslFile, []byte(sslConf), 0644); err != nil {
		return fmt.Errorf("gravar ssl vhost: %w", err)
	}
	return nil
}

func (m *Manager) ativarHTTPSComCert(fqdn, slug, certChain, certKey string) error {
	if err := m.escreverVhostHTTP(fqdn, slug, true); err != nil {
		return err
	}
	if err := m.escreverVhostSSL(fqdn, slug, certChain, certKey); err != nil {
		return err
	}
	if out, err := run("a2ensite", slug+".conf"); err != nil {
		return fmt.Errorf("a2ensite http: %s", out)
	}
	if out, err := run("a2ensite", slug+"-le-ssl.conf"); err != nil {
		return fmt.Errorf("a2ensite ssl: %s", out)
	}
	if out, err := run("apache2ctl", "configtest"); err != nil {
		return fmt.Errorf("configtest: %s", out)
	}
	if out, err := run("systemctl", "reload", "apache2"); err != nil {
		if out2, err2 := run("systemctl", "restart", "apache2"); err2 != nil {
			return fmt.Errorf("reload/restart apache: %s | %s", out, out2)
		}
	}
	return nil
}

func (m *Manager) emitirCertificadoWebroot(fqdn string) string {
	if _, _, ok := certificadoLetsEncrypt(fqdn); ok {
		return "certificado ja existia"
	}
	out, _ := run(
		"certbot", "certonly", "--webroot",
		"-w", "/var/www/html",
		"-d", fqdn,
		"--non-interactive",
		"--agree-tos",
		"--email", m.CertbotEmail,
		"--keep-until-expiring",
	)
	return out
}

func (m *Manager) Remover(fqdn string) Result {
	fqdn = strings.ToLower(strings.TrimSpace(fqdn))
	if fqdn == "" || !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`).MatchString(fqdn) {
		return Result{OK: false, Status: "erro", Message: "FQDN invalido"}
	}

	slug := SlugFromFQDN(fqdn)
	httpFile := filepath.Join(m.SitesDir, slug+".conf")
	sslFile := filepath.Join(m.SitesDir, slug+"-le-ssl.conf")
	disabledDir := filepath.Join("/var/www/fp-dominio-disabled", slug)

	m.limparVhostsRelacionados(fqdn, slug)

	if err := os.MkdirAll(disabledDir, 0755); err != nil {
		return Result{OK: false, Status: "erro", Message: "mkdir desativado: " + err.Error(), Slug: slug}
	}
	if err := os.WriteFile(filepath.Join(disabledDir, "index.html"), []byte(htmlDominioDesativado(fqdn)), 0644); err != nil {
		return Result{OK: false, Status: "erro", Message: "gravar pagina desativada: " + err.Error(), Slug: slug}
	}

	if err := os.MkdirAll("/var/www/html/.well-known/acme-challenge", 0755); err != nil {
		return Result{OK: false, Status: "erro", Message: "mkdir acme-challenge: " + err.Error(), Slug: slug}
	}

	httpConf := fmt.Sprintf(`<VirtualHost *:80>
    ServerName %s

    DocumentRoot %s

    Alias /.well-known/acme-challenge/ /var/www/html/.well-known/acme-challenge/
    <Directory /var/www/html/.well-known/acme-challenge/>
        Options None
        AllowOverride None
        Require all granted
    </Directory>

    <Directory %s>
        Options -Indexes
        AllowOverride None
        Require all granted
    </Directory>

    ErrorDocument 403 /index.html
    ErrorDocument 404 /index.html
</VirtualHost>
`, fqdn, disabledDir, disabledDir)

	if err := os.WriteFile(httpFile, []byte(httpConf), 0644); err != nil {
		return Result{OK: false, Status: "erro", Message: "gravar vhost http desativado: " + err.Error(), Slug: slug}
	}
	if out, err := run("a2ensite", slug+".conf"); err != nil {
		return Result{OK: false, Status: "erro", Message: "a2ensite http desativado: " + out, Slug: slug}
	}
	if out, err := run("apache2ctl", "configtest"); err != nil {
		return Result{OK: false, Status: "erro", Message: "configtest http: " + out, Slug: slug}
	}
	if out, err := run("systemctl", "reload", "apache2"); err != nil {
		return Result{OK: false, Status: "erro", Message: "reload apache (http): " + out, Slug: slug}
	}

	m.tentarEmitirCertificadoLE(fqdn, disabledDir)

	certChain, certKey, sslOK, certErr := m.garantirCertificadoDesativado(fqdn, disabledDir)
	if certErr != nil {
		return Result{OK: false, Status: "erro", Message: "certificado ssl: " + certErr.Error(), Slug: slug}
	}

	sslInclude := ""
	if arquivoExiste("/etc/letsencrypt/options-ssl-apache.conf") {
		sslInclude = "    Include /etc/letsencrypt/options-ssl-apache.conf\n"
	}

	sslConf := fmt.Sprintf(`<IfModule mod_ssl.c>
<VirtualHost *:443>
    ServerName %s

    DocumentRoot %s

    SSLEngine on
    SSLCertificateFile %s
    SSLCertificateKeyFile %s
%s
    <Directory %s>
        Options -Indexes
        AllowOverride None
        Require all granted
    </Directory>

    ErrorDocument 403 /index.html
    ErrorDocument 404 /index.html
</VirtualHost>
</IfModule>
`, fqdn, disabledDir, certChain, certKey, sslInclude, disabledDir)

	if err := os.WriteFile(sslFile, []byte(sslConf), 0644); err != nil {
		return Result{OK: false, Status: "erro", Message: "gravar vhost ssl desativado: " + err.Error(), Slug: slug}
	}
	if out, err := run("a2ensite", slug+"-le-ssl.conf"); err != nil {
		return Result{OK: false, Status: "erro", Message: "a2ensite ssl desativado: " + out, Slug: slug}
	}

	if out, err := run("apache2ctl", "configtest"); err != nil {
		return Result{OK: false, Status: "erro", Message: "configtest: " + out, Slug: slug}
	}
	if out, err := run("systemctl", "reload", "apache2"); err != nil {
		if out2, err2 := run("systemctl", "restart", "apache2"); err2 != nil {
			return Result{OK: false, Status: "erro", Message: "reload/restart apache: " + out + " | " + out2, Slug: slug}
		}
	}

	vhosts, _ := run("apache2ctl", "-S")
	msg := "Dominio desativado no servidor (nao exibe mais o sistema)"
	if strings.Contains(vhosts, fqdn) {
		msg += " — vhost ativo confirmado"
	}
	if sslOK {
		msg += " — SSL Let's Encrypt ativo"
	} else {
		msg += " — HTTPS com certificado temporario (acesse via HTTP ou aguarde emissao SSL)"
	}

	return Result{
		OK:      true,
		SSLOK:   sslOK,
		Status:  "removido",
		Message: msg,
		Slug:    slug,
	}
}

func (m *Manager) limparVhostsRelacionados(fqdn, slug string) {
	vistos := map[string]bool{}
	dirs := []string{m.SitesDir, sitesEnabledDir}

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasSuffix(name, ".conf") {
				continue
			}
			site := strings.TrimSuffix(name, ".conf")
			if vistos[site] {
				continue
			}

			path := filepath.Join(dir, name)
			if entry.IsDir() {
				continue
			}
			if dir == sitesEnabledDir {
				if target, err := os.Readlink(path); err == nil {
					if filepath.IsAbs(target) {
						path = target
					} else {
						path = filepath.Join(m.SitesDir, filepath.Base(target))
					}
				}
			}

			raw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			content := string(raw)
			relacionado := strings.Contains(name, slug) ||
				strings.Contains(strings.ToLower(content), "servername "+fqdn) ||
				strings.Contains(strings.ToLower(content), "serveralias "+fqdn)
			if !relacionado {
				continue
			}

			vistos[site] = true
			_, _ = run("a2dissite", site)

			if strings.Contains(content, "ProxyPass") {
				_ = os.Remove(path)
			}
		}
	}
}

func certificadoParaFQDN(fqdn string) (chain string, key string, ok bool) {
	if chain, key, ok := certificadoLetsEncrypt(fqdn); ok {
		return chain, key, true
	}

	snakePaths := [][2]string{
		{"/etc/ssl/certs/ssl-cert-snakeoil.pem", "/etc/ssl/private/ssl-cert-snakeoil.key"},
		{"/etc/ssl/certs/ssl-cert-snakeoil.pem", "/etc/ssl/private/ssl-cert-snakeoil.pem"},
	}
	for _, p := range snakePaths {
		if arquivoExiste(p[0]) && arquivoExiste(p[1]) {
			return p[0], p[1], true
		}
	}
	return "", "", false
}

func certificadoLetsEncrypt(fqdn string) (chain string, key string, ok bool) {
	leChain := filepath.Join("/etc/letsencrypt/live", fqdn, "fullchain.pem")
	leKey := filepath.Join("/etc/letsencrypt/live", fqdn, "privkey.pem")
	if arquivoExiste(leChain) && arquivoExiste(leKey) {
		return leChain, leKey, true
	}
	return "", "", false
}

func limparCertificadoAutoassinado(disabledDir string) {
	certDir := filepath.Join(disabledDir, "certs")
	_ = os.Remove(filepath.Join(certDir, "fullchain.pem"))
	_ = os.Remove(filepath.Join(certDir, "privkey.pem"))
}

func (m *Manager) tentarEmitirCertificadoLE(fqdn, disabledDir string) {
	if _, _, ok := certificadoLetsEncrypt(fqdn); ok {
		return
	}
	if strings.TrimSpace(m.CertbotEmail) == "" {
		return
	}
	limparCertificadoAutoassinado(disabledDir)
	_, _ = run(
		"certbot", "certonly", "--webroot",
		"-w", "/var/www/html",
		"-d", fqdn,
		"--non-interactive",
		"--agree-tos",
		"--email", m.CertbotEmail,
		"--keep-until-expiring",
	)
}

func (m *Manager) garantirCertificadoDesativado(fqdn, disabledDir string) (chain string, key string, sslOK bool, err error) {
	if chain, key, ok := certificadoLetsEncrypt(fqdn); ok {
		limparCertificadoAutoassinado(disabledDir)
		return chain, key, true, nil
	}

	certDir := filepath.Join(disabledDir, "certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return "", "", false, err
	}
	chain = filepath.Join(certDir, "fullchain.pem")
	key = filepath.Join(certDir, "privkey.pem")
	if arquivoExiste(chain) && arquivoExiste(key) {
		return chain, key, false, nil
	}

	out, err := run(
		"openssl", "req", "-x509", "-nodes", "-days", "365", "-newkey", "rsa:2048",
		"-keyout", key, "-out", chain,
		"-subj", "/CN="+fqdn,
	)
	if err != nil {
		return "", "", false, fmt.Errorf("openssl: %s (%v)", out, err)
	}
	return chain, key, false, nil
}

func arquivoExiste(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func htmlDominioDesativado(fqdn string) string {
	return `<!DOCTYPE html>
<html lang="pt-br">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Domínio desativado</title>
  <style>
    body { font-family: system-ui, sans-serif; background: #0f172a; color: #e2e8f0; margin: 0; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
    .box { max-width: 520px; padding: 2rem; text-align: center; }
    h1 { font-size: 1.5rem; margin-bottom: .75rem; }
    p { color: #94a3b8; line-height: 1.5; }
    code { color: #cbd5e1; }
  </style>
</head>
<body>
  <div class="box">
    <h1>Domínio desativado</h1>
    <p>O endereço <code>` + fqdn + `</code> não está mais vinculado ao FranqueadoPro.</p>
    <p>Se você é o franqueado, acesse pelo painel oficial ou cadastre um novo domínio personalizado.</p>
  </div>
</body>
</html>`
}

func (m *Manager) RetentarSSL(fqdn string) Result {
	fqdn = strings.ToLower(strings.TrimSpace(fqdn))
	if fqdn == "" || !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$`).MatchString(fqdn) {
		return Result{OK: false, Status: "erro", Message: "FQDN invalido"}
	}
	slug := SlugFromFQDN(fqdn)

	if err := os.MkdirAll("/var/www/html/.well-known/acme-challenge", 0755); err != nil {
		return Result{OK: false, Status: "erro", Message: "mkdir acme-challenge: " + err.Error(), Slug: slug}
	}

	// Garante vhost HTTP unico com ServerName (sem SSL incompleto)
	_ = m.escreverVhostHTTP(fqdn, slug, false)
	_, _ = run("a2ensite", slug+".conf")
	// Desabilita SSL antigo incompleto que causa ambiguidade no certbot --apache
	_, _ = run("a2dissite", slug+"-le-ssl.conf")
	_, _ = run("apache2ctl", "configtest")
	_, _ = run("systemctl", "reload", "apache2")

	certOut := m.emitirCertificadoWebroot(fqdn)
	chain, key, ok := certificadoLetsEncrypt(fqdn)
	if !ok {
		return Result{
			OK:      true,
			SSLOK:   false,
			Status:  "pendente_dns",
			Message: "SSL ainda pendente: " + certOut,
			Slug:    slug,
		}
	}

	if err := m.ativarHTTPSComCert(fqdn, slug, chain, key); err != nil {
		return Result{
			OK:      true,
			SSLOK:   false,
			Status:  "pendente_dns",
			Message: "Certificado existe, falha ao ativar no Apache: " + err.Error(),
			Slug:    slug,
		}
	}

	return Result{
		OK:      true,
		SSLOK:   true,
		Status:  "ativo",
		Message: "SSL emitido/instalado com sucesso",
		Slug:    slug,
	}
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
