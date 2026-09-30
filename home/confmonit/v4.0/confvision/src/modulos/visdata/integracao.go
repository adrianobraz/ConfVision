package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"confvision/src/config"
)

type integracaoConfig struct {
	ID                     int
	IDFranqueado           string
	Sistema                string
	Nome                   string
	Ativo                  bool
	EventosURL             string
	AuthTipo               string
	AuthUser               string
	AuthPass               string
	AuthToken              string
	EmpresaCodigo          string
	EventoCodigo           string
	SetorPadrao            string
	ParticaoPadrao         string
	IdentificacaoPadrao    string
	CodigoIntegracaoImagem string
	EnviarImagem           bool
	DuplaComunicacao       bool
	WebhookInboundAtivo    bool
	CodigoEventoArmar      string
	CodigoEventoDesarmar   string
}

var integracaoSistemasValidos = map[string]bool{
	"nenhum": true, "moni": true, "dguard": true, "segware": true, "confmonit": true,
}

func normalizeIntegracaoSistema(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if integracaoSistemasValidos[s] {
		return s
	}
	return "nenhum"
}

func normalizeDuplaComunicacao(sistema string, dupla bool) bool {
	s := normalizeIntegracaoSistema(sistema)
	if s == "nenhum" || s == "confmonit" {
		return false
	}
	return dupla
}

func integracaoDuplaComunicacaoHabilitada(sistema string) bool {
	s := normalizeIntegracaoSistema(sistema)
	return s == "moni" || s == "dguard" || s == "segware"
}

func defaultIntegracaoFranqueado(idFranqueado string) map[string]any {
	return map[string]any{
		"id":                       nil,
		"id_franqueado":            idFranqueado,
		"sistema":                  "nenhum",
		"nome":                     "",
		"ativo":                    false,
		"eventos_url":              "",
		"auth_tipo":                "none",
		"auth_user":                "",
		"auth_pass":                "",
		"auth_token":               "",
		"empresa_codigo":           "",
		"evento_codigo":            "",
		"setor_padrao":             "",
		"particao_padrao":          "",
		"identificacao_padrao":     "",
		"codigo_integracao_imagem": "046",
		"enviar_imagem":            true,
		"dupla_comunicacao":        false,
		"webhook_inbound_ativo":    false,
		"codigo_evento_armar":      "130",
		"codigo_evento_desarmar":   "131",
	}
}

func GetIntegracaoFranqueado(ctx context.Context, idFranqueado string) (map[string]any, error) {
	if strings.TrimSpace(idFranqueado) == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `
SELECT id, created_at, updated_at, id_franqueado, sistema, nome, ativo, eventos_url,
       auth_tipo, auth_user, auth_pass, auth_token, empresa_codigo, evento_codigo,
       setor_padrao, particao_padrao, identificacao_padrao, codigo_integracao_imagem, enviar_imagem, dupla_comunicacao,
       webhook_inbound_ativo, codigo_evento_armar, codigo_evento_desarmar
FROM vis_integracao_config
WHERE id_franqueado = $1
ORDER BY updated_at DESC, id DESC
LIMIT 1`, idFranqueado)

	var integracao map[string]any
	item, err := scanIntegracaoRow(row)
	if err == sql.ErrNoRows {
		integracao = defaultIntegracaoFranqueado(idFranqueado)
	} else if err != nil {
		return nil, err
	} else {
		integracao = item
	}

	meta, err := buildIntegracaoMeta(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	return map[string]any{"integracao": integracao, "meta": meta}, nil
}

func buildIntegracaoMeta(ctx context.Context, idFranqueado string) (map[string]any, error) {
	temLicenca := false
	var licencaErro string
	if t, err := FranqueadoTemTerminal(ctx, idFranqueado); err != nil {
		licencaErro = err.Error()
	} else {
		temLicenca = t
	}
	meta := map[string]any{
		"imagem_public_base_url":    config.ImagemPublicBaseURL,
		"imagem_secret_configurado": config.ImagemPublicSecret != "",
		"confmonit_servidor_ok": config.TerminalNotifyEnabled &&
			config.ReceptorWebURL != "" &&
			config.ReceptorWebSenha != "",
		"confmonit_licenca_ok": temLicenca,
		"dispatch_habilitado":  config.IntegracaoDispatchEnabled,
	}
	if licencaErro != "" {
		meta["confmonit_licenca_erro"] = licencaErro
	}
	return meta, nil
}

func SaveIntegracaoFranqueado(ctx context.Context, input map[string]any) (map[string]any, error) {
	idFra := strings.TrimSpace(strVal(input, "id_franqueado"))
	if idFra == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	sistema := normalizeIntegracaoSistema(strVal(input, "sistema"))
	input["sistema"] = sistema

	db, err := DB()
	if err != nil {
		return nil, err
	}

	// Nenhum: persiste registro inativo (nao apaga) para o dispatch respeitar a escolha.
	if sistema == "nenhum" {
		input["ativo"] = false
		input["dupla_comunicacao"] = false
	}

	dupla := normalizeDuplaComunicacao(sistema, boolDefault(input, "dupla_comunicacao", false))
	if dupla {
		temLicenca, err := FranqueadoTemTerminal(ctx, idFra)
		if err != nil {
			return nil, fmt.Errorf("verificar licenca terminal: %w", err)
		}
		if !temLicenca {
			return nil, fmt.Errorf("dupla comunicacao requer licenca de terminal ConfMonit")
		}
	}
	input["dupla_comunicacao"] = dupla

	if sistema == "confmonit" {
		temLicenca, err := FranqueadoTemTerminal(ctx, idFra)
		if err != nil {
			return nil, fmt.Errorf("verificar licenca terminal: %w", err)
		}
		if !temLicenca {
			return nil, fmt.Errorf("franqueado sem licenca de terminal ConfMonit")
		}
	}

	var existingID int
	err = db.QueryRowContext(ctx, `
SELECT id FROM vis_integracao_config
WHERE id_franqueado = $1
ORDER BY updated_at DESC, id DESC
LIMIT 1`, idFra).Scan(&existingID)
	hasExisting := err == nil
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if hasExisting {
		if _, err := UpdateIntegracao(ctx, existingID, input); err != nil {
			return nil, err
		}
		_, _ = db.ExecContext(ctx, `DELETE FROM vis_integracao_config WHERE id_franqueado = $1 AND id != $2`, idFra, existingID)
	} else {
		if _, err := CreateIntegracao(ctx, input); err != nil {
			return nil, err
		}
	}
	return GetIntegracaoFranqueado(ctx, idFra)
}

func TestIntegracaoFranqueado(ctx context.Context, idFranqueado string, input map[string]any) (map[string]any, error) {
	cfg, found, err := loadIntegracaoFranqueadoConfig(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	sistemaForm := strings.ToLower(strings.TrimSpace(strVal(input, "sistema")))
	if !found || cfg == nil {
		if sistemaForm == "moni" && strings.TrimSpace(strVal(input, "eventos_url")) != "" {
			return TestIntegracaoMoni(ctx, 0, input)
		}
		return map[string]any{
			"sucesso": false, "mensagem": "nenhuma integracao configurada",
		}, nil
	}
	if cfg.Sistema == "nenhum" {
		if sistemaForm == "moni" && strings.TrimSpace(strVal(input, "eventos_url")) != "" {
			return TestIntegracaoMoni(ctx, cfg.ID, input)
		}
		return map[string]any{
			"sucesso": false, "mensagem": "integracao desativada (nenhum)",
		}, nil
	}
	if !cfg.Ativo && strings.ToLower(cfg.Sistema) != "moni" {
		return map[string]any{
			"sucesso": false, "mensagem": "integracao inativa",
		}, nil
	}

	switch strings.ToLower(firstNonEmpty(sistemaForm, cfg.Sistema)) {
	case "moni":
		primary, err := TestIntegracaoMoni(ctx, cfg.ID, input)
		if err != nil {
			return nil, err
		}
		return mergeTesteDuplaComunicacao(ctx, idFranqueado, cfg.ID, firstNonEmpty(sistemaForm, cfg.Sistema), input, primary)
	case "confmonit":
		return TestIntegracaoConfmonit(ctx, idFranqueado, cfg.ID)
	case "dguard", "segware":
		primary := map[string]any{
			"sucesso": false, "mensagem": "integracao " + cfg.Sistema + " em desenvolvimento", "sistema": cfg.Sistema,
		}
		return mergeTesteDuplaComunicacao(ctx, idFranqueado, cfg.ID, firstNonEmpty(sistemaForm, cfg.Sistema), input, primary)
	default:
		return map[string]any{
			"sucesso": false, "mensagem": "sistema nao suportado para teste",
		}, nil
	}
}

func mergeTesteDuplaComunicacao(ctx context.Context, idFranqueado string, integracaoID int, sistema string, input, primary map[string]any) (map[string]any, error) {
	sistema = normalizeIntegracaoSistema(sistema)
	ativo := boolDefault(input, "ativo", true)
	dupla := normalizeDuplaComunicacao(sistema, boolDefault(input, "dupla_comunicacao", false))
	if !dupla || !ativo || !integracaoDuplaComunicacaoHabilitada(sistema) {
		return primary, nil
	}

	terminal, err := TestIntegracaoConfmonit(ctx, idFranqueado, integracaoID)
	if err != nil {
		return nil, err
	}

	out := map[string]any{}
	for k, v := range primary {
		out[k] = v
	}
	out["dupla_comunicacao"] = true
	out["terminal"] = terminal

	primaryOK := primary != nil && primary["sucesso"] == true
	terminalOK := terminal != nil && terminal["sucesso"] == true
	out["sucesso"] = primaryOK || terminalOK

	msgs := []string{}
	if primary != nil {
		if m, ok := primary["mensagem"].(string); ok && m != "" {
			msgs = append(msgs, strings.ToUpper(nullStrMap(primary, "sistema"))+": "+m)
		}
	}
	if terminal != nil {
		if m, ok := terminal["mensagem"].(string); ok && m != "" {
			msgs = append(msgs, "CONFMONIT: "+m)
		}
	}
	if len(msgs) > 0 {
		out["mensagem"] = strings.Join(msgs, " · ")
	}
	return out, nil
}

func nullStrMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func ListIntegracoesByFranqueado(ctx context.Context, idFranqueado string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, updated_at, id_franqueado, sistema, nome, ativo, eventos_url,
       auth_tipo, auth_user, auth_pass, auth_token, empresa_codigo, evento_codigo,
       setor_padrao, particao_padrao, identificacao_padrao, codigo_integracao_imagem, enviar_imagem, dupla_comunicacao,
       webhook_inbound_ativo, codigo_evento_armar, codigo_evento_desarmar
FROM vis_integracao_config
WHERE id_franqueado = $1
ORDER BY sistema ASC, id ASC`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []map[string]any
	for rows.Next() {
		item, err := scanIntegracaoRow(rows)
		if err != nil {
			return nil, err
		}
		lista = append(lista, item)
	}
	return map[string]any{"dados": lista}, nil
}

func GetIntegracao(ctx context.Context, id int) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	row := db.QueryRowContext(ctx, `
SELECT id, created_at, updated_at, id_franqueado, sistema, nome, ativo, eventos_url,
       auth_tipo, auth_user, auth_pass, auth_token, empresa_codigo, evento_codigo,
       setor_padrao, particao_padrao, identificacao_padrao, codigo_integracao_imagem, enviar_imagem, dupla_comunicacao,
       webhook_inbound_ativo, codigo_evento_armar, codigo_evento_desarmar
FROM vis_integracao_config WHERE id = $1`, id)
	return scanIntegracaoRow(row)
}

func scanIntegracaoRow(scanner interface {
	Scan(dest ...any) error
}) (map[string]any, error) {
	var id int
	var created, updated sql.NullTime
	var idFra, sistema, nome, eventosURL, authTipo, authUser, authPass, authToken sql.NullString
	var empresa, evento, setor, part, ident, codImg sql.NullString
	var codArmar, codDesarmar sql.NullString
	var ativo, enviarImagem, duplaComunicacao, webhookInbound sql.NullBool
	err := scanner.Scan(&id, &created, &updated, &idFra, &sistema, &nome, &ativo, &eventosURL,
		&authTipo, &authUser, &authPass, &authToken, &empresa, &evento, &setor, &part, &ident, &codImg, &enviarImagem, &duplaComunicacao,
		&webhookInbound, &codArmar, &codDesarmar)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":                       id,
		"created_at":               nullTime(created),
		"updated_at":               nullTime(updated),
		"id_franqueado":            nullStr(idFra),
		"sistema":                  nullStr(sistema),
		"nome":                     nullStr(nome),
		"ativo":                    nullBool(ativo),
		"eventos_url":              nullStr(eventosURL),
		"auth_tipo":                nullStr(authTipo),
		"auth_user":                nullStr(authUser),
		"auth_pass":                nullStr(authPass),
		"auth_token":               nullStr(authToken),
		"empresa_codigo":           nullStr(empresa),
		"evento_codigo":            nullStr(evento),
		"setor_padrao":             nullStr(setor),
		"particao_padrao":          nullStr(part),
		"identificacao_padrao":     nullStr(ident),
		"codigo_integracao_imagem": nullStr(codImg),
		"enviar_imagem":            nullBool(enviarImagem),
		"dupla_comunicacao":        nullBool(duplaComunicacao),
		"webhook_inbound_ativo":    nullBool(webhookInbound),
		"codigo_evento_armar":      nullStr(codArmar),
		"codigo_evento_desarmar":   nullStr(codDesarmar),
	}, nil
}

func CreateIntegracao(ctx context.Context, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	sistema := strings.ToLower(strings.TrimSpace(strVal(input, "sistema")))
	if sistema == "" {
		return nil, fmt.Errorf("sistema obrigatorio")
	}
	idFra := strVal(input, "id_franqueado")
	if idFra == "" {
		return nil, fmt.Errorf("id_franqueado obrigatorio")
	}
	var id int
	err = db.QueryRowContext(ctx, `
INSERT INTO vis_integracao_config (
    id_franqueado, sistema, nome, ativo, eventos_url, auth_tipo, auth_user, auth_pass, auth_token,
    empresa_codigo, evento_codigo, setor_padrao, particao_padrao, identificacao_padrao,
    codigo_integracao_imagem, enviar_imagem, dupla_comunicacao,
    webhook_inbound_ativo, codigo_evento_armar, codigo_evento_desarmar, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,NOW())
RETURNING id`,
		idFra, sistema, strVal(input, "nome"), boolDefault(input, "ativo", true),
		strVal(input, "eventos_url"), defaultStr(strVal(input, "auth_tipo"), "none"),
		strVal(input, "auth_user"), strVal(input, "auth_pass"), strVal(input, "auth_token"),
		strVal(input, "empresa_codigo"), strVal(input, "evento_codigo"),
		strVal(input, "setor_padrao"), strVal(input, "particao_padrao"), strVal(input, "identificacao_padrao"),
		defaultStr(strVal(input, "codigo_integracao_imagem"), "046"),
		boolDefault(input, "enviar_imagem", true),
		boolDefault(input, "dupla_comunicacao", false),
		boolDefault(input, "webhook_inbound_ativo", false),
		defaultStr(strVal(input, "codigo_evento_armar"), "130"),
		defaultStr(strVal(input, "codigo_evento_desarmar"), "131"),
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetIntegracao(ctx, id)
}

func UpdateIntegracao(ctx context.Context, id int, input map[string]any) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	n := 1
	for _, pair := range []struct {
		key string
		val any
	}{
		{"sistema", normalizeIntegracaoSistema(strVal(input, "sistema"))},
		{"nome", strVal(input, "nome")},
		{"ativo", boolDefault(input, "ativo", true)},
		{"eventos_url", strVal(input, "eventos_url")},
		{"auth_tipo", strVal(input, "auth_tipo")},
		{"auth_user", strVal(input, "auth_user")},
		{"auth_pass", strVal(input, "auth_pass")},
		{"auth_token", strVal(input, "auth_token")},
		{"empresa_codigo", strVal(input, "empresa_codigo")},
		{"evento_codigo", strVal(input, "evento_codigo")},
		{"setor_padrao", strVal(input, "setor_padrao")},
		{"particao_padrao", strVal(input, "particao_padrao")},
		{"identificacao_padrao", strVal(input, "identificacao_padrao")},
		{"codigo_integracao_imagem", strVal(input, "codigo_integracao_imagem")},
		{"enviar_imagem", boolDefault(input, "enviar_imagem", true)},
		{"dupla_comunicacao", boolDefault(input, "dupla_comunicacao", false)},
		{"webhook_inbound_ativo", boolDefault(input, "webhook_inbound_ativo", false)},
		{"codigo_evento_armar", defaultStr(strVal(input, "codigo_evento_armar"), "130")},
		{"codigo_evento_desarmar", defaultStr(strVal(input, "codigo_evento_desarmar"), "131")},
	} {
		if _, ok := input[pair.key]; ok {
			sets = append(sets, fmt.Sprintf("%s = $%d", pair.key, n))
			args = append(args, pair.val)
			n++
		}
	}
	if len(sets) == 1 {
		return GetIntegracao(ctx, id)
	}
	args = append(args, id)
	q := fmt.Sprintf("UPDATE vis_integracao_config SET %s WHERE id = $%d", strings.Join(sets, ", "), n)
	if _, err := db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return GetIntegracao(ctx, id)
}

func DeleteIntegracao(ctx context.Context, id int) error {
	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `DELETE FROM vis_integracao_config WHERE id = $1`, id)
	return err
}

func ListIntegracaoLog(ctx context.Context, idFranqueado string, limit, offset int, dataDe, dataAte, clienteNome string) (map[string]any, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}

	where := "WHERE l.id_franqueado = $1"
	args := []any{idFranqueado}
	n := 2

	if strings.TrimSpace(dataDe) != "" {
		where += fmt.Sprintf(" AND l.created_at >= $%d", n)
		args = append(args, strings.TrimSpace(dataDe))
		n++
	}
	if strings.TrimSpace(dataAte) != "" {
		where += fmt.Sprintf(" AND l.created_at <= $%d", n)
		args = append(args, strings.TrimSpace(dataAte))
		n++
	}

	if termo := strings.TrimSpace(clienteNome); termo != "" {
		ids, err := SearchClienteIDsByNome(ctx, idFranqueado, termo)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return map[string]any{
				"dados":         []map[string]any{},
				"itemsReceived": 0,
				"has_more":      false,
			}, nil
		}
		ph := make([]string, len(ids))
		for i, id := range ids {
			ph[i] = fmt.Sprintf("$%d", n)
			args = append(args, id)
			n++
		}
		where += fmt.Sprintf(
			" AND COALESCE(NULLIF(TRIM(e.id_cliente), ''), NULLIF(TRIM(c.id_cliente), '')) IN (%s)",
			strings.Join(ph, ","),
		)
	}

	args = append(args, limit, offset)
	limitPh := fmt.Sprintf("$%d", n)
	offsetPh := fmt.Sprintf("$%d", n+1)

	q := fmt.Sprintf(`
SELECT l.id, l.created_at, l.id_franqueado, l.vis_integracao_id, l.vis_evento_id, l.sistema,
       l.sucesso, l.http_status, l.mensagem, l.payload_resumo,
       COALESCE(NULLIF(TRIM(c.nome), ''), '') AS camera_nome,
       COALESCE(NULLIF(TRIM(e.id_cliente), ''), NULLIF(TRIM(c.id_cliente), ''), '') AS id_cliente
FROM vis_integracao_log l
LEFT JOIN vis_evento e ON e.id = l.vis_evento_id
LEFT JOIN vis_camera c ON c.id = e.vis_camera_id
%s
ORDER BY l.created_at DESC
LIMIT %s OFFSET %s`, where, limitPh, offsetPh)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clienteNomes := map[string]string{}
	var lista []map[string]any
	for rows.Next() {
		var id, httpStatus sql.NullInt64
		var created sql.NullTime
		var idFra, sistema, msg, resumo, cameraNome, idCliente sql.NullString
		var integID, eventoID sql.NullInt64
		var sucesso sql.NullBool
		if err := rows.Scan(&id, &created, &idFra, &integID, &eventoID, &sistema,
			&sucesso, &httpStatus, &msg, &resumo, &cameraNome, &idCliente); err != nil {
			return nil, err
		}
		idFraStr := sqlString(idFra)
		idCliStr := sqlString(idCliente)
		clienteNome := ""
		if idCliStr != "" {
			if cached, ok := clienteNomes[idCliStr]; ok {
				clienteNome = cached
			} else {
				nome, err := GetClienteNome(ctx, firstNonEmpty(idFraStr, idFranqueado), idCliStr)
				if err != nil {
					return nil, err
				}
				clienteNome = nome
				clienteNomes[idCliStr] = nome
			}
		}
		lista = append(lista, map[string]any{
			"id": id.Int64, "created_at": nullTime(created), "id_franqueado": nullStr(idFra),
			"vis_integracao_id": nullInt(integID), "vis_evento_id": nullInt(eventoID),
			"sistema": nullStr(sistema), "sucesso": nullBool(sucesso),
			"http_status": nullInt(httpStatus), "mensagem": nullStr(msg),
			"payload_resumo": nullStr(resumo),
			"camera_nome": nullStr(cameraNome),
			"id_cliente": idCliStr, "cliente_nome": clienteNome,
		})
	}
	return map[string]any{
		"dados":         lista,
		"itemsReceived": len(lista),
		"has_more":      len(lista) >= limit,
	}, nil
}

func insertIntegracaoLog(ctx context.Context, idFra string, integID, eventoID int, sistema string,
	sucesso bool, httpStatus int, msg, resumo string) {
	db, err := DB()
	if err != nil {
		return
	}
	var integ any = integID
	if integID <= 0 {
		integ = nil
	}
	var evt any = eventoID
	if eventoID <= 0 {
		evt = nil
	}
	_, _ = db.ExecContext(ctx, `
INSERT INTO vis_integracao_log (
    id_franqueado, vis_integracao_id, vis_evento_id, sistema, sucesso, http_status, mensagem, payload_resumo
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		idFra, integ, evt, sistema, sucesso, httpStatus, truncateStr(msg, 500), truncateStr(resumo, 500))
}

func loadIntegracaoConfig(ctx context.Context, id int) (*integracaoConfig, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var cfg integracaoConfig
	var nome, eventosURL, authTipo, authUser, authPass, authToken sql.NullString
	var empresa, evento, setor, part, ident, codImg sql.NullString
	var codArmar, codDesarmar sql.NullString
	var ativo, enviarImagem, duplaComunicacao, webhookInbound sql.NullBool
	err = db.QueryRowContext(ctx, `
SELECT id, id_franqueado, sistema, nome, ativo, eventos_url, auth_tipo, auth_user, auth_pass, auth_token,
       empresa_codigo, evento_codigo, setor_padrao, particao_padrao, identificacao_padrao,
       codigo_integracao_imagem, enviar_imagem, dupla_comunicacao,
       webhook_inbound_ativo, codigo_evento_armar, codigo_evento_desarmar
FROM vis_integracao_config WHERE id = $1`, id).Scan(
		&cfg.ID, &cfg.IDFranqueado, &cfg.Sistema, &nome, &ativo, &eventosURL,
		&authTipo, &authUser, &authPass, &authToken, &empresa, &evento, &setor, &part, &ident, &codImg, &enviarImagem, &duplaComunicacao,
		&webhookInbound, &codArmar, &codDesarmar,
	)
	if err != nil {
		return nil, err
	}
	cfg.Nome = sqlString(nome)
	cfg.Ativo = ativo.Valid && ativo.Bool
	cfg.EventosURL = sqlString(eventosURL)
	cfg.AuthTipo = defaultStr(sqlString(authTipo), "none")
	cfg.AuthUser = sqlString(authUser)
	cfg.AuthPass = sqlString(authPass)
	cfg.AuthToken = sqlString(authToken)
	cfg.EmpresaCodigo = sqlString(empresa)
	cfg.EventoCodigo = sqlString(evento)
	cfg.SetorPadrao = sqlString(setor)
	cfg.ParticaoPadrao = sqlString(part)
	cfg.IdentificacaoPadrao = sqlString(ident)
	cfg.CodigoIntegracaoImagem = defaultStr(sqlString(codImg), "046")
	cfg.EnviarImagem = !enviarImagem.Valid || enviarImagem.Bool
	cfg.DuplaComunicacao = duplaComunicacao.Valid && duplaComunicacao.Bool
	cfg.WebhookInboundAtivo = webhookInbound.Valid && webhookInbound.Bool
	cfg.CodigoEventoArmar = defaultStr(sqlString(codArmar), "130")
	cfg.CodigoEventoDesarmar = defaultStr(sqlString(codDesarmar), "131")
	return &cfg, nil
}

func loadIntegracaoFranqueadoConfig(ctx context.Context, idFranqueado string) (*integracaoConfig, bool, error) {
	if strings.TrimSpace(idFranqueado) == "" {
		return nil, false, nil
	}
	db, err := DB()
	if err != nil {
		return nil, false, err
	}
	var cfg integracaoConfig
	var nome, eventosURL, authTipo, authUser, authPass, authToken sql.NullString
	var empresa, evento, setor, part, ident, codImg sql.NullString
	var codArmar, codDesarmar sql.NullString
	var ativo, enviarImagem, duplaComunicacao, webhookInbound sql.NullBool
	err = db.QueryRowContext(ctx, `
SELECT id, id_franqueado, sistema, nome, ativo, eventos_url, auth_tipo, auth_user, auth_pass, auth_token,
       empresa_codigo, evento_codigo, setor_padrao, particao_padrao, identificacao_padrao,
       codigo_integracao_imagem, enviar_imagem, dupla_comunicacao,
       webhook_inbound_ativo, codigo_evento_armar, codigo_evento_desarmar
FROM vis_integracao_config
WHERE id_franqueado = $1
ORDER BY updated_at DESC, id DESC
LIMIT 1`, idFranqueado).Scan(
		&cfg.ID, &cfg.IDFranqueado, &cfg.Sistema, &nome, &ativo, &eventosURL,
		&authTipo, &authUser, &authPass, &authToken, &empresa, &evento, &setor, &part, &ident, &codImg, &enviarImagem, &duplaComunicacao,
		&webhookInbound, &codArmar, &codDesarmar,
	)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	cfg.Nome = sqlString(nome)
	cfg.Sistema = normalizeIntegracaoSistema(cfg.Sistema)
	cfg.Ativo = ativo.Valid && ativo.Bool
	cfg.EventosURL = sqlString(eventosURL)
	cfg.AuthTipo = defaultStr(sqlString(authTipo), "none")
	cfg.AuthUser = sqlString(authUser)
	cfg.AuthPass = sqlString(authPass)
	cfg.AuthToken = sqlString(authToken)
	cfg.EmpresaCodigo = sqlString(empresa)
	cfg.EventoCodigo = sqlString(evento)
	cfg.SetorPadrao = sqlString(setor)
	cfg.ParticaoPadrao = sqlString(part)
	cfg.IdentificacaoPadrao = sqlString(ident)
	cfg.CodigoIntegracaoImagem = defaultStr(sqlString(codImg), "046")
	cfg.EnviarImagem = !enviarImagem.Valid || enviarImagem.Bool
	cfg.DuplaComunicacao = duplaComunicacao.Valid && duplaComunicacao.Bool
	cfg.WebhookInboundAtivo = webhookInbound.Valid && webhookInbound.Bool
	cfg.CodigoEventoArmar = defaultStr(sqlString(codArmar), "130")
	cfg.CodigoEventoDesarmar = defaultStr(sqlString(codDesarmar), "131")
	return &cfg, true, nil
}

func defaultStr(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}

func truncateStr(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func eventoImagemJaLiberada(ctx context.Context, eventoID int) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var ts sql.NullTime
	err = db.QueryRowContext(ctx, `
SELECT imagem_liberada_em FROM vis_evento WHERE id = $1`, eventoID).Scan(&ts)
	if err != nil {
		if err == sql.ErrNoRows {
		 return false, nil
		}
		return false, err
	}
	return ts.Valid, nil
}
