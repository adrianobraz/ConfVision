package visdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
}

func ListIntegracoesByFranqueado(ctx context.Context, idFranqueado string) (map[string]any, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, updated_at, id_franqueado, sistema, nome, ativo, eventos_url,
       auth_tipo, auth_user, auth_pass, auth_token, empresa_codigo, evento_codigo,
       setor_padrao, particao_padrao, identificacao_padrao, codigo_integracao_imagem, enviar_imagem
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
       setor_padrao, particao_padrao, identificacao_padrao, codigo_integracao_imagem, enviar_imagem
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
	var ativo, enviarImagem sql.NullBool
	err := scanner.Scan(&id, &created, &updated, &idFra, &sistema, &nome, &ativo, &eventosURL,
		&authTipo, &authUser, &authPass, &authToken, &empresa, &evento, &setor, &part, &ident, &codImg, &enviarImagem)
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
    codigo_integracao_imagem, enviar_imagem, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NOW())
RETURNING id`,
		idFra, sistema, strVal(input, "nome"), boolDefault(input, "ativo", true),
		strVal(input, "eventos_url"), defaultStr(strVal(input, "auth_tipo"), "none"),
		strVal(input, "auth_user"), strVal(input, "auth_pass"), strVal(input, "auth_token"),
		strVal(input, "empresa_codigo"), strVal(input, "evento_codigo"),
		strVal(input, "setor_padrao"), strVal(input, "particao_padrao"), strVal(input, "identificacao_padrao"),
		defaultStr(strVal(input, "codigo_integracao_imagem"), "046"),
		boolDefault(input, "enviar_imagem", true),
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
	_ = offset
	_ = dataDe
	_ = dataAte
	_ = clienteNome
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, created_at, id_franqueado, vis_integracao_id, vis_evento_id, sistema,
       sucesso, http_status, mensagem, payload_resumo
FROM vis_integracao_log
WHERE id_franqueado = $1
ORDER BY created_at DESC
LIMIT $2`, idFranqueado, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []map[string]any
	for rows.Next() {
		var id, httpStatus sql.NullInt64
		var created sql.NullTime
		var idFra, sistema, msg, resumo sql.NullString
		var integID, eventoID sql.NullInt64
		var sucesso sql.NullBool
		if err := rows.Scan(&id, &created, &idFra, &integID, &eventoID, &sistema,
			&sucesso, &httpStatus, &msg, &resumo); err != nil {
			return nil, err
		}
		lista = append(lista, map[string]any{
			"id": id.Int64, "created_at": nullTime(created), "id_franqueado": nullStr(idFra),
			"vis_integracao_id": nullInt(integID), "vis_evento_id": nullInt(eventoID),
			"sistema": nullStr(sistema), "sucesso": nullBool(sucesso),
			"http_status": nullInt(httpStatus), "mensagem": nullStr(msg),
			"payload_resumo": nullStr(resumo),
		})
	}
	return map[string]any{"dados": lista}, nil
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
	var ativo, enviarImagem sql.NullBool
	err = db.QueryRowContext(ctx, `
SELECT id, id_franqueado, sistema, nome, ativo, eventos_url, auth_tipo, auth_user, auth_pass, auth_token,
       empresa_codigo, evento_codigo, setor_padrao, particao_padrao, identificacao_padrao,
       codigo_integracao_imagem, enviar_imagem
FROM vis_integracao_config WHERE id = $1`, id).Scan(
		&cfg.ID, &cfg.IDFranqueado, &cfg.Sistema, &nome, &ativo, &eventosURL,
		&authTipo, &authUser, &authPass, &authToken, &empresa, &evento, &setor, &part, &ident, &codImg, &enviarImagem,
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
	return &cfg, nil
}

func listIntegracoesAtivasMoni(ctx context.Context, idFranqueado string) ([]integracaoConfig, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
SELECT id, id_franqueado, sistema, nome, ativo, eventos_url, auth_tipo, auth_user, auth_pass, auth_token,
       empresa_codigo, evento_codigo, setor_padrao, particao_padrao, identificacao_padrao,
       codigo_integracao_imagem, enviar_imagem
FROM vis_integracao_config
WHERE id_franqueado = $1 AND ativo = TRUE AND sistema = 'moni'`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []integracaoConfig
	for rows.Next() {
		var cfg integracaoConfig
		var nome, eventosURL, authTipo, authUser, authPass, authToken sql.NullString
		var empresa, evento, setor, part, ident, codImg sql.NullString
		var ativo, enviarImagem sql.NullBool
		if err := rows.Scan(&cfg.ID, &cfg.IDFranqueado, &cfg.Sistema, &nome, &ativo, &eventosURL,
			&authTipo, &authUser, &authPass, &authToken, &empresa, &evento, &setor, &part, &ident, &codImg, &enviarImagem); err != nil {
			return nil, err
		}
		cfg.Nome = sqlString(nome)
		cfg.Ativo = true
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
		out = append(out, cfg)
	}
	return out, nil
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

func MoniInboundPOST(ctx context.Context, payload map[string]any) (int, []byte, error) {
	_ = ctx
	_ = payload
	return 501, []byte(`{"erro":"integracao moni inbound nao disponivel"}`), nil
}
