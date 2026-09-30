package visdata

import "context"

type SMTPConfigEnvio struct {
	IDFranqueado string `json:"id_franqueado"`
	SmtpHost     string `json:"smtp_host"`
	SmtpPorta    int    `json:"smtp_porta"`
	SmtpUsuario  string `json:"smtp_usuario"`
	SmtpSenha    string `json:"smtp_senha"`
	SmtpSeguranca string `json:"smtp_seguranca"`
	SmtpTimeoutSeg int  `json:"smtp_timeout_seg"`
	Ativo        bool   `json:"ativo"`
}

func GetSMTPConfig(ctx context.Context, idFranqueado string, interno bool) (SMTPConfigEnvio, error) {
	_ = ctx
	_ = interno
	return SMTPConfigEnvio{IDFranqueado: idFranqueado}, nil
}

func SMTPConfiguradoAtivo(ctx context.Context, idFranqueado string) (bool, error) {
	_ = ctx
	_ = idFranqueado
	return false, nil
}

func SaveSMTPConfig(ctx context.Context, cfg SMTPConfigEnvio, novaSenha string) error {
	_ = ctx
	_ = cfg
	_ = novaSenha
	return nil
}

func TestarSMTP(ctx context.Context, cfg SMTPConfigEnvio) map[string]any {
	_ = ctx
	_ = cfg
	return map[string]any{"ok": false, "mensagem": "smtp stub"}
}

func EnviarEmailSMTP(ctx context.Context, cfg SMTPConfigEnvio, dest, assunto, corpo string) error {
	_ = ctx
	_ = cfg
	_ = dest
	_ = assunto
	_ = corpo
	return nil
}
