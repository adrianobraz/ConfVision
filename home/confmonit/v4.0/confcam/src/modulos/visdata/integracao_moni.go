package visdata

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	moniComplementoImagemPrefixo = "046"
	moniIdentificacaoFallback    = "E"
	moniDemoHashTeste            = "6rorjlqej4wy"
)

// Hashes demo em imagem.dnsid.com.br (vis_evento 990001–990003).
var moniDemoHashes = []string{"yw4zz83ag4x9", "6rorjlqej4wy", "qa5m9dx7p4el"}

type eventoIntegracaoRow struct {
	ID            int
	IDFranqueado  string
	IDCliente     string
	IDDispositivo string
	Conta         string
	Particao      string
	ZonaUser      string
	TipoDeteccao  string
	SnapshotURL   string
	CameraNome    string
}

func moniPrefixoImagem(cfg integracaoConfig) string {
	p := strings.TrimSpace(cfg.CodigoIntegracaoImagem)
	if p == "" {
		return moniComplementoImagemPrefixo
	}
	return p
}

func moniIdentificacao(cfg integracaoConfig, cameraNome string) string {
	if id := strings.TrimSpace(cfg.IdentificacaoPadrao); id != "" {
		return id
	}
	if n := strings.TrimSpace(cameraNome); n != "" {
		return n
	}
	return moniIdentificacaoFallback
}

func moniSetorValor(cfg integracaoConfig, zonaUser string) any {
	raw := strings.TrimSpace(cfg.SetorPadrao)
	if raw == "" {
		raw = normalizeZonaUser(zonaUser)
	}
	return moniJSONScalar(raw, 1)
}

func moniEmpresaValor(empresaCodigo string) any {
	return moniJSONScalar(strings.TrimSpace(empresaCodigo), 1)
}

func resolveEmpresaCodigo(ctx context.Context, cfg integracaoConfig, idFranqueado, idCliente string) (string, error) {
	codigo, err := GetClienteCodEmpresa(ctx, idFranqueado, idCliente)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(codigo) != "" {
		return codigo, nil
	}
	return strings.TrimSpace(cfg.EmpresaCodigo), nil
}

func moniParticaoValor(cfg integracaoConfig, particao string) string {
	if p := strings.TrimSpace(cfg.ParticaoPadrao); p != "" {
		return padDigits(p, 2)
	}
	if p := padDigits(particao, 2); p != "" && p != "00" {
		return p
	}
	return "01"
}

// moniJSONScalar envia int quando numerico (Moni local espera numero em empresa/setor).
func moniJSONScalar(raw string, defaultInt int) any {
	s := strings.TrimSpace(raw)
	if s == "" {
		return defaultInt
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return s
}

func buildMoniPayload(cfg integracaoConfig, row *eventoIntegracaoRow, complemento string, empresaCodigo string) map[string]any {
	clienteCodigo := padDigits(row.Conta, 4)
	if clienteCodigo == "" || clienteCodigo == "0000" {
		clienteCodigo = "0001"
	}
	return map[string]any{
		"empresa":       moniEmpresaValor(empresaCodigo),
		"cliente":       clienteCodigo,
		"particao":      moniParticaoValor(cfg, row.Particao),
		"evento":        defaultStr(cfg.EventoCodigo, "130"),
		"identificacao": moniIdentificacao(cfg, row.CameraNome),
		"setor":         moniSetorValor(cfg, row.ZonaUser),
		"complemento":   complemento,
	}
}

func dispatchMoniIntegracao(ctx context.Context, eventoID int, cfg integracaoConfig) error {
	if strings.TrimSpace(cfg.EventosURL) == "" {
		return fmt.Errorf("eventos_url vazio")
	}

	row, err := loadEventoIntegracaoRow(ctx, eventoID)
	if err != nil {
		return err
	}
	if row == nil {
		return nil
	}
	if strings.EqualFold(row.TipoDeteccao, "sensor") {
		return nil
	}
	if strings.TrimSpace(row.SnapshotURL) == "" {
		return nil
	}

	jaLiberada, err := eventoImagemJaLiberada(ctx, eventoID)
	if err != nil {
		return err
	}
	if jaLiberada {
		return nil
	}

	clienteCodigo, err := GetClienteCodigoInterno(ctx, row.IDFranqueado, row.IDCliente)
	if err != nil {
		return err
	}
	if clienteCodigo == "" {
		clienteCodigo = padDigits(row.Conta, 4)
	}
	if clienteCodigo == "" || clienteCodigo == "0000" {
		log.Printf("[MONI] evento=%d ignorado: codigo cliente ausente", eventoID)
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "moni", false, 0,
			"codigo_interno do cliente nao configurado", "")
		return nil
	}
	row.Conta = clienteCodigo

	empresaCodigo, err := resolveEmpresaCodigo(ctx, cfg, row.IDFranqueado, row.IDCliente)
	if err != nil {
		return err
	}

	complemento := ""
	codigoImagem := ""
	if cfg.EnviarImagem {
		codigoImagem = CodigoImagemPublica(eventoID)
		if codigoImagem == "" {
			return fmt.Errorf("IMAGEM_PUBLIC_SECRET nao configurado")
		}
		complemento = ComplementoImagemMoni(moniPrefixoImagem(cfg), codigoImagem)
	}

	payload := buildMoniPayload(cfg, row, complemento, empresaCodigo)

	httpStatus, err := postMoniGerarEvento(ctx, cfg, payload)
	resumo, _ := json.Marshal(map[string]any{
		"cliente": clienteCodigo, "empresa": empresaCodigo, "complemento": complemento, "evento": cfg.EventoCodigo,
	})
	if err != nil {
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "moni", false, httpStatus, err.Error(), string(resumo))
		return err
	}

	if cfg.EnviarImagem && codigoImagem != "" {
		if err := LiberarImagemEvento(ctx, eventoID, codigoImagem, cfg.ID); err != nil {
			insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "moni", false, httpStatus,
				"evento enviado mas falha ao liberar imagem: "+err.Error(), string(resumo))
			return err
		}
	}

	insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "moni", true, httpStatus, "ok", string(resumo))
	log.Printf("[MONI] ok evento=%d integracao=%d cliente=%s complemento=%s", eventoID, cfg.ID, clienteCodigo, complemento)
	return nil
}

func postMoniGerarEvento(ctx context.Context, cfg integracaoConfig, payload map[string]any) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.EventosURL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	applyIntegracaoAuth(req, cfg)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		msg := strings.TrimSpace(string(raw))
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return res.StatusCode, fmt.Errorf("HTTP %d: %s", res.StatusCode, msg)
	}
	return res.StatusCode, nil
}

func applyIntegracaoAuth(req *http.Request, cfg integracaoConfig) {
	switch strings.ToLower(strings.TrimSpace(cfg.AuthTipo)) {
	case "basic":
		req.SetBasicAuth(cfg.AuthUser, cfg.AuthPass)
	case "bearer":
		token := strings.TrimSpace(cfg.AuthToken)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "apikey", "api_key":
		token := strings.TrimSpace(cfg.AuthToken)
		if token != "" {
			req.Header.Set("X-API-Key", token)
		}
	}
}

func applyMoniTestOverrides(cfg *integracaoConfig, input map[string]any) {
	if cfg == nil || len(input) == 0 {
		return
	}
	if _, ok := input["eventos_url"]; ok {
		cfg.EventosURL = strings.TrimSpace(strVal(input, "eventos_url"))
	}
	if _, ok := input["auth_tipo"]; ok {
		cfg.AuthTipo = defaultStr(strings.TrimSpace(strVal(input, "auth_tipo")), "none")
	}
	if _, ok := input["auth_user"]; ok {
		cfg.AuthUser = strVal(input, "auth_user")
	}
	if _, ok := input["auth_pass"]; ok {
		cfg.AuthPass = strVal(input, "auth_pass")
	}
	if _, ok := input["auth_token"]; ok {
		cfg.AuthToken = strVal(input, "auth_token")
	}
	if _, ok := input["empresa_codigo"]; ok {
		cfg.EmpresaCodigo = strVal(input, "empresa_codigo")
	}
	if _, ok := input["evento_codigo"]; ok {
		cfg.EventoCodigo = strVal(input, "evento_codigo")
	}
	if _, ok := input["identificacao_padrao"]; ok {
		cfg.IdentificacaoPadrao = strVal(input, "identificacao_padrao")
	}
	if _, ok := input["setor_padrao"]; ok {
		cfg.SetorPadrao = strVal(input, "setor_padrao")
	}
	if _, ok := input["particao_padrao"]; ok {
		cfg.ParticaoPadrao = strVal(input, "particao_padrao")
	}
	if _, ok := input["codigo_integracao_imagem"]; ok {
		cfg.CodigoIntegracaoImagem = defaultStr(strVal(input, "codigo_integracao_imagem"), "046")
	}
	if b := boolVal(input, "enviar_imagem"); b != nil {
		cfg.EnviarImagem = *b
	}
}

func TestIntegracaoMoni(ctx context.Context, integracaoID int, input map[string]any) (map[string]any, error) {
	cfg := &integracaoConfig{
		Sistema:                "moni",
		EnviarImagem:           true,
		CodigoIntegracaoImagem: "046",
		EventoCodigo:           "130",
		IdentificacaoPadrao:    moniIdentificacaoFallback,
		ParticaoPadrao:         "01",
		SetorPadrao:            "1",
	}
	if integracaoID > 0 {
		loaded, err := loadIntegracaoConfig(ctx, integracaoID)
		if err != nil {
			return nil, err
		}
		cfg = loaded
	}
	if !strings.EqualFold(cfg.Sistema, "moni") {
		return nil, fmt.Errorf("teste disponivel apenas para integracao moni")
	}
	applyMoniTestOverrides(cfg, input)

	if strings.TrimSpace(cfg.EventosURL) == "" {
		return map[string]any{
			"sucesso": false, "mensagem": "URL de eventos obrigatoria",
		}, nil
	}

	clienteTeste := strings.TrimSpace(strVal(input, "cliente_teste"))
	if clienteTeste == "" {
		clienteTeste = "0001"
	}

	complemento := ""
	if cfg.EnviarImagem {
		complemento = ComplementoImagemMoni(moniPrefixoImagem(*cfg), moniDemoHashTeste)
	}

	row := &eventoIntegracaoRow{
		Conta:      clienteTeste,
		Particao:   cfg.ParticaoPadrao,
		CameraNome: "",
		ZonaUser:   cfg.SetorPadrao,
	}
	payload := buildMoniPayload(*cfg, row, complemento, cfg.EmpresaCodigo)

	httpStatus, err := postMoniGerarEvento(ctx, *cfg, payload)
	if err != nil {
		return map[string]any{
			"sucesso": false, "http_status": httpStatus, "mensagem": err.Error(),
			"payload": payload, "demo_imagem": "https://imagem.dnsid.com.br/" + moniDemoHashTeste,
		}, nil
	}
	return map[string]any{
		"sucesso": true, "http_status": httpStatus, "mensagem": "ok", "payload": payload,
		"demo_imagem": "https://imagem.dnsid.com.br/" + moniDemoHashTeste,
		"demo_hashes": moniDemoHashes,
	}, nil
}

func loadEventoIntegracaoRow(ctx context.Context, eventoID int) (*eventoIntegracaoRow, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var row eventoIntegracaoRow
	var idFra, idCli, idDisp, conta, part, tipo, snap sql.NullString
	var zona, camNome sql.NullString
	err = db.QueryRowContext(ctx, `
SELECT e.id, e.id_franqueado, e.id_cliente, e.id_dispositivo, e.conta, e.particao, e.tipo_deteccao,
       e.snapshot_url,
       COALESCE(NULLIF(TRIM(c.zonauser), ''), NULLIF(TRIM(e.canal), ''), '001') AS zonauser,
       COALESCE(NULLIF(TRIM(c.nome), ''), '') AS camera_nome
FROM vis_evento e
LEFT JOIN vis_camera c ON c.id = e.vis_camera_id
WHERE e.id = $1`, eventoID).Scan(
		&row.ID, &idFra, &idCli, &idDisp, &conta, &part, &tipo, &snap, &zona, &camNome,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	row.IDFranqueado = sqlString(idFra)
	row.IDCliente = sqlString(idCli)
	row.IDDispositivo = sqlString(idDisp)
	row.Conta = sqlString(conta)
	row.Particao = sqlString(part)
	row.TipoDeteccao = sqlString(tipo)
	row.SnapshotURL = sqlString(snap)
	row.ZonaUser = sqlString(zona)
	row.CameraNome = sqlString(camNome)
	return &row, nil
}
