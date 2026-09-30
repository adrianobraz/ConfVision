package visdata

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	IDFranqueado     string `json:"id_franqueado"`
	NomeConfig       string `json:"nome_config"`
	Descricao        string `json:"descricao"`
	Ativo            bool   `json:"ativo"`
	Provedor         string `json:"provedor"`
	RemetenteNome    string `json:"remetente_nome"`
	RemetenteEmail   string `json:"remetente_email"`
	SmtpHost         string `json:"smtp_host"`
	SmtpPorta        int    `json:"smtp_porta"`
	SmtpSeguranca    string `json:"smtp_seguranca"`
	SmtpAutenticacao bool   `json:"smtp_autenticacao"`
	SmtpUsuario      string `json:"smtp_usuario"`
	SmtpTimeoutSeg   int    `json:"smtp_timeout_seg"`
	SenhaConfigurada bool   `json:"senha_configurada"`
}

type SMTPConfigEnvio struct {
	SMTPConfig
	SmtpSenha string `json:"smtp_senha"`
}

type SMTPTestPasso struct {
	Label string `json:"label"`
	OK    bool   `json:"ok"`
	Erro  string `json:"erro,omitempty"`
}

type SMTPTestResult struct {
	OK       bool            `json:"ok"`
	Passos   []SMTPTestPasso `json:"passos"`
	Mensagem string          `json:"mensagem"`
}

func GetSMTPConfig(ctx context.Context, idFranqueado string, incluirSenha bool) (SMTPConfigEnvio, error) {
	out := SMTPConfigEnvio{
		SMTPConfig: SMTPConfig{
			IDFranqueado: idFranqueado, SmtpPorta: 587, SmtpSeguranca: "starttls",
			SmtpTimeoutSeg: 30, Ativo: true,
		},
	}
	db, err := DB()
	if err != nil {
		return out, err
	}
	var senhaEnc sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT nome_config, descricao, ativo, provedor,
       remetente_nome, remetente_email,
       smtp_host, smtp_porta, smtp_seguranca, smtp_autenticacao,
       smtp_usuario, smtp_senha, smtp_timeout_seg
FROM ops_franqueado_smtp_config
WHERE id_franqueado = $1`, idFranqueado).Scan(
		&out.NomeConfig, &out.Descricao, &out.Ativo, &out.Provedor,
		&out.RemetenteNome, &out.RemetenteEmail,
		&out.SmtpHost, &out.SmtpPorta, &out.SmtpSeguranca, &out.SmtpAutenticacao,
		&out.SmtpUsuario, &senhaEnc, &out.SmtpTimeoutSeg,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.SenhaConfigurada = senhaEnc.Valid && strings.TrimSpace(senhaEnc.String) != ""
	if incluirSenha && out.SenhaConfigurada {
		out.SmtpSenha, err = decryptSmtpSecret(senhaEnc.String)
	}
	return out, err
}

func SMTPConfiguradoAtivo(ctx context.Context, idFranqueado string) (bool, error) {
	cfg, err := GetSMTPConfig(ctx, idFranqueado, false)
	if err != nil {
		return false, err
	}
	if !cfg.Ativo {
		return false, nil
	}
	if strings.TrimSpace(cfg.SmtpHost) == "" || strings.TrimSpace(cfg.RemetenteEmail) == "" {
		return false, nil
	}
	return cfg.SenhaConfigurada || !cfg.SmtpAutenticacao, nil
}

func SaveSMTPConfig(ctx context.Context, cfg SMTPConfigEnvio, novaSenha string) error {
	if strings.TrimSpace(cfg.IDFranqueado) == "" {
		return errors.New("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	senhaEnc := ""
	if strings.TrimSpace(novaSenha) != "" {
		senhaEnc, err = encryptSmtpSecret(novaSenha)
		if err != nil {
			return err
		}
	}
	if senhaEnc == "" {
		var old sql.NullString
		_ = db.QueryRowContext(ctx, `SELECT smtp_senha FROM ops_franqueado_smtp_config WHERE id_franqueado = $1`, cfg.IDFranqueado).Scan(&old)
		if old.Valid {
			senhaEnc = old.String
		}
	}
	porta := cfg.SmtpPorta
	if porta <= 0 {
		porta = 587
	}
	timeout := cfg.SmtpTimeoutSeg
	if timeout <= 0 {
		timeout = 30
	}
	seg := strings.ToLower(strings.TrimSpace(cfg.SmtpSeguranca))
	if seg == "" {
		seg = "starttls"
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_franqueado_smtp_config (
    id_franqueado, nome_config, descricao, ativo, provedor,
    remetente_nome, remetente_email,
    smtp_host, smtp_porta, smtp_seguranca, smtp_autenticacao,
    smtp_usuario, smtp_senha, smtp_timeout_seg, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,NOW())
ON CONFLICT (id_franqueado) DO UPDATE SET
    nome_config = EXCLUDED.nome_config,
    descricao = EXCLUDED.descricao,
    ativo = EXCLUDED.ativo,
    provedor = EXCLUDED.provedor,
    remetente_nome = EXCLUDED.remetente_nome,
    remetente_email = EXCLUDED.remetente_email,
    smtp_host = EXCLUDED.smtp_host,
    smtp_porta = EXCLUDED.smtp_porta,
    smtp_seguranca = EXCLUDED.smtp_seguranca,
    smtp_autenticacao = EXCLUDED.smtp_autenticacao,
    smtp_usuario = EXCLUDED.smtp_usuario,
    smtp_senha = EXCLUDED.smtp_senha,
    smtp_timeout_seg = EXCLUDED.smtp_timeout_seg,
    updated_at = NOW()`,
		cfg.IDFranqueado, cfg.NomeConfig, cfg.Descricao, cfg.Ativo, cfg.Provedor,
		cfg.RemetenteNome, cfg.RemetenteEmail,
		cfg.SmtpHost, porta, seg, cfg.SmtpAutenticacao,
		cfg.SmtpUsuario, senhaEnc, timeout,
	)
	return err
}

func TestarSMTP(ctx context.Context, cfg SMTPConfigEnvio) SMTPTestResult {
	res := SMTPTestResult{Passos: []SMTPTestPasso{}}
	host := strings.TrimSpace(cfg.SmtpHost)
	porta := cfg.SmtpPorta
	if porta <= 0 {
		porta = 587
	}
	if host == "" {
		res.Mensagem = "Informe o servidor SMTP."
		res.Passos = append(res.Passos, SMTPTestPasso{Label: "Servidor informado", OK: false, Erro: "host vazio"})
		return res
	}

	if _, err := net.LookupHost(host); err != nil {
		res.Passos = append(res.Passos, SMTPTestPasso{Label: "Servidor encontrado", OK: false, Erro: err.Error()})
		res.Mensagem = "Falha na resolucao DNS."
		return res
	}
	res.Passos = append(res.Passos, SMTPTestPasso{Label: "Servidor encontrado", OK: true})

	addr := fmt.Sprintf("%s:%d", host, porta)
	timeout := time.Duration(cfg.SmtpTimeoutSeg) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		res.Passos = append(res.Passos, SMTPTestPasso{Label: fmt.Sprintf("Porta %d acessivel", porta), OK: false, Erro: err.Error()})
		res.Mensagem = "Porta SMTP inacessivel."
		return res
	}
	res.Passos = append(res.Passos, SMTPTestPasso{Label: fmt.Sprintf("Porta %d acessivel", porta), OK: true})

	seg := strings.ToLower(strings.TrimSpace(cfg.SmtpSeguranca))
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		res.Passos = append(res.Passos, SMTPTestPasso{Label: "Conexao SMTP", OK: false, Erro: err.Error()})
		res.Mensagem = "Falha ao iniciar cliente SMTP."
		return res
	}
	defer client.Close()

	tlsOk := false
	switch seg {
	case "ssl", "tls":
		tlsCfg := &tls.Config{ServerName: host}
		if err := client.StartTLS(tlsCfg); err != nil {
			res.Passos = append(res.Passos, SMTPTestPasso{Label: "Conexao TLS estabelecida", OK: false, Erro: err.Error()})
			res.Mensagem = "Falha TLS."
			return res
		}
		tlsOk = true
	case "starttls":
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsCfg := &tls.Config{ServerName: host}
			if err := client.StartTLS(tlsCfg); err != nil {
				res.Passos = append(res.Passos, SMTPTestPasso{Label: "Conexao TLS estabelecida", OK: false, Erro: err.Error()})
				res.Mensagem = "Falha STARTTLS."
				return res
			}
			tlsOk = true
		}
	case "nenhuma", "none", "":
		tlsOk = true
	}
	if tlsOk {
		res.Passos = append(res.Passos, SMTPTestPasso{Label: "Conexao TLS estabelecida", OK: true})
	}

	if cfg.SmtpAutenticacao {
		senha := strings.TrimSpace(cfg.SmtpSenha)
		if senha == "" && cfg.SenhaConfigurada && cfg.IDFranqueado != "" {
			stored, err := GetSMTPConfig(ctx, cfg.IDFranqueado, true)
			if err == nil {
				senha = stored.SmtpSenha
			}
		}
		auth := smtp.PlainAuth("", strings.TrimSpace(cfg.SmtpUsuario), senha, host)
		if err := client.Auth(auth); err != nil {
			res.Passos = append(res.Passos, SMTPTestPasso{Label: "Usuario autenticado", OK: false, Erro: err.Error()})
			res.Mensagem = "Falha na autenticacao SMTP."
			return res
		}
		res.Passos = append(res.Passos, SMTPTestPasso{Label: "Usuario autenticado", OK: true})
	} else {
		res.Passos = append(res.Passos, SMTPTestPasso{Label: "Autenticacao desabilitada", OK: true})
	}

	res.OK = true
	res.Mensagem = "Conexao SMTP funcionando."
	return res
}

func EnviarEmailSMTP(ctx context.Context, cfg SMTPConfigEnvio, destinatario, assunto, corpoHTML string) error {
	if strings.TrimSpace(destinatario) == "" {
		return errors.New("destinatario obrigatorio")
	}
	full, err := GetSMTPConfig(ctx, cfg.IDFranqueado, true)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.SmtpHost) != "" {
		full.SmtpHost = cfg.SmtpHost
		full.SmtpPorta = cfg.SmtpPorta
		full.SmtpSeguranca = cfg.SmtpSeguranca
		full.SmtpAutenticacao = cfg.SmtpAutenticacao
		full.SmtpUsuario = cfg.SmtpUsuario
		if strings.TrimSpace(cfg.SmtpSenha) != "" {
			full.SmtpSenha = cfg.SmtpSenha
		}
		full.RemetenteEmail = cfg.RemetenteEmail
		full.RemetenteNome = cfg.RemetenteNome
	}
	if !full.Ativo {
		return errors.New("smtp inativo")
	}
	from := strings.TrimSpace(full.RemetenteEmail)
	if from == "" {
		return errors.New("remetente nao configurado")
	}
	if assunto == "" {
		assunto = "Teste SMTP FranqueadoPro"
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, destinatario, assunto, corpoHTML)
	return sendMailSMTP(full, destinatario, []byte(msg))
}

func sendMailSMTP(cfg SMTPConfigEnvio, destinatario string, msg []byte) error {
	host := strings.TrimSpace(cfg.SmtpHost)
	porta := cfg.SmtpPorta
	if porta <= 0 {
		porta = 587
	}
	addr := fmt.Sprintf("%s:%d", host, porta)
	var auth smtp.Auth
	if cfg.SmtpAutenticacao {
		senha := cfg.SmtpSenha
		if senha == "" {
			var err error
			senha, err = decryptSmtpSecret(cfg.SmtpSenha)
			if err != nil {
				return err
			}
		}
		auth = smtp.PlainAuth("", strings.TrimSpace(cfg.SmtpUsuario), senha, host)
	}
	from := strings.TrimSpace(cfg.RemetenteEmail)
	seg := strings.ToLower(strings.TrimSpace(cfg.SmtpSeguranca))
	if seg == "ssl" || seg == "tls" {
		return sendMailTLS(addr, host, auth, from, destinatario, msg)
	}
	return smtp.SendMail(addr, auth, from, []string{destinatario}, msg)
}

func sendMailTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
