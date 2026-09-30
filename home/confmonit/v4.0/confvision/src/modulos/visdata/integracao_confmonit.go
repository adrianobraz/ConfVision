package visdata

import (
	"confvision/src/config"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type terminalTesteRow struct {
	EventoID   int
	Conta      string
	Particao   string
	ZonaUser   string
	CameraNome string
}

func TestIntegracaoConfmonit(ctx context.Context, idFranqueado string, integracaoID int) (map[string]any, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	if idFranqueado == "" {
		return map[string]any{
			"sucesso": false, "mensagem": "id_franqueado obrigatorio", "sistema": "confmonit",
		}, nil
	}

	temLicenca, err := FranqueadoTemTerminal(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	if !temLicenca {
		return map[string]any{
			"sucesso": false, "mensagem": "franqueado sem licenca de terminal ConfMonit", "sistema": "confmonit",
		}, nil
	}
	if !config.TerminalNotifyEnabled {
		return map[string]any{
			"sucesso": false, "mensagem": "TERMINAL_NOTIFY_ENABLED=false no servidor Go", "sistema": "confmonit",
		}, nil
	}
	if config.ReceptorWebURL == "" || config.ReceptorWebSenha == "" {
		return map[string]any{
			"sucesso": false, "mensagem": "RECEPTOR_WEB_URL ou RECEPTOR_WEB_SENHA nao configurado no servidor", "sistema": "confmonit",
		}, nil
	}

	row, err := loadEventoTesteTerminal(ctx, idFranqueado)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return map[string]any{
			"sucesso": false,
			"mensagem": "nenhum evento ConfVision encontrado; provoque uma deteccao analitica antes do teste",
			"sistema":  "confmonit",
		}, nil
	}

	conta := padDigits(row.Conta, 4)
	particao := padDigits(row.Particao, 2)
	zona := normalizeZonaUser(row.ZonaUser)
	if conta == "" || conta == "0000" {
		return map[string]any{
			"sucesso": false,
			"mensagem": "ultimo evento sem conta valida; verifique cadastro da camera",
			"sistema":  "confmonit",
		}, nil
	}

	payload := map[string]any{
		"idFranqueado":  idFranqueado,
		"conta":         conta,
		"particao":      particao,
		"zonauser":      zona,
		"vis_evento_id": row.EventoID,
		"senha":         config.ReceptorWebSenha,
		"codigo":        config.TerminalContactID,
	}

	resumoObj := map[string]any{
		"conta": conta, "particao": particao, "zonauser": zona,
		"vis_evento_id": row.EventoID, "camera": row.CameraNome,
	}
	resumo, _ := json.Marshal(resumoObj)

	idEvento, idProcesso, err := notifyReceptorWithRetry(ctx, payload)
	if err != nil {
		insertIntegracaoLog(ctx, idFranqueado, integracaoID, row.EventoID, "confmonit", false, 0, err.Error(), string(resumo))
		return map[string]any{
			"sucesso": false, "mensagem": err.Error(), "sistema": "confmonit",
			"payload": payload, "vis_evento_id": row.EventoID,
		}, nil
	}
	if idEvento == "" && idProcesso == "" {
		msg := "receptor respondeu sem idEvento/idProcesso"
		insertIntegracaoLog(ctx, idFranqueado, integracaoID, row.EventoID, "confmonit", false, 200, msg, string(resumo))
		return map[string]any{
			"sucesso": false, "mensagem": msg, "sistema": "confmonit",
			"payload": payload, "vis_evento_id": row.EventoID,
		}, nil
	}

	insertIntegracaoLog(ctx, idFranqueado, integracaoID, row.EventoID, "confmonit", true, 200, "ok", string(resumo))
	return map[string]any{
		"sucesso": true, "mensagem": "ok", "sistema": "confmonit",
		"payload": payload, "vis_evento_id": row.EventoID,
		"id_evento": idEvento, "id_processo": idProcesso,
	}, nil
}

// dispatchConfmonitIntegracao envia evento ao terminal ConfMonit e registra vis_integracao_log.
func dispatchConfmonitIntegracao(ctx context.Context, eventoID int, cfg integracaoConfig) error {
	if !config.TerminalNotifyEnabled {
		insertIntegracaoLog(ctx, cfg.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0,
			"TERMINAL_NOTIFY_ENABLED=false", "")
		return nil
	}
	if config.ReceptorWebURL == "" || config.ReceptorWebSenha == "" {
		insertIntegracaoLog(ctx, cfg.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0,
			"RECEPTOR_WEB_URL ou RECEPTOR_WEB_SENHA nao configurado", "")
		return fmt.Errorf("receptor web nao configurado")
	}

	row, err := loadEventoDispatchRow(ctx, eventoID)
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

	temLicenca, err := FranqueadoTemTerminal(ctx, row.IDFranqueado)
	if err != nil {
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0, err.Error(), "")
		return err
	}
	if !temLicenca {
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0,
			"franqueado sem licenca de terminal ConfMonit", "")
		return nil
	}

	conta := padDigits(row.Conta, 4)
	particao := padDigits(row.Particao, 2)
	zona := normalizeZonaUser(row.ZonaUser)
	if row.IDFranqueado == "" || conta == "" || conta == "0000" {
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0,
			"franqueado/conta ausentes", "")
		return nil
	}

	payload := map[string]any{
		"idFranqueado":  row.IDFranqueado,
		"conta":         conta,
		"particao":      particao,
		"zonauser":      zona,
		"vis_evento_id": eventoID,
		"senha":         config.ReceptorWebSenha,
		"codigo":        config.TerminalContactID,
	}
	resumoObj := map[string]any{
		"conta": conta, "particao": particao, "zonauser": zona, "vis_evento_id": eventoID,
	}
	resumo, _ := json.Marshal(resumoObj)

	idEvento, idProcesso, err := notifyReceptorWithRetry(ctx, payload)
	if err != nil {
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0, err.Error(), string(resumo))
		return err
	}
	if idEvento == "" && idProcesso == "" {
		msg := "receptor respondeu sem idEvento/idProcesso"
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 200, msg, string(resumo))
		return fmt.Errorf("%s", msg)
	}

	db, err := DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
UPDATE vis_evento
SET id_evento = COALESCE(NULLIF($2, ''), id_evento),
    id_processo = COALESCE(NULLIF($3, ''), id_processo)
WHERE id = $1 AND (id_evento IS NULL OR id_evento = '')`,
		eventoID, idEvento, idProcesso)
	if err != nil {
		insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", false, 0,
			"falha ao atualizar ids terminal: "+err.Error(), string(resumo))
		return fmt.Errorf("atualizar ids terminal: %w", err)
	}

	insertIntegracaoLog(ctx, row.IDFranqueado, cfg.ID, eventoID, "confmonit", true, 200, "ok", string(resumo))
	log.Printf("[INTEGRACAO] confmonit ok evento=%d idEvento=%s processo=%s dupla=%v",
		eventoID, idEvento, idProcesso, cfg.DuplaComunicacao)
	return nil
}

func loadEventoTesteTerminal(ctx context.Context, idFranqueado string) (*terminalTesteRow, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}

	row, err := queryEventoTesteTerminal(ctx, db, idFranqueado, true)
	if err != nil {
		return nil, err
	}
	if row != nil {
		return row, nil
	}
	return queryEventoTesteTerminal(ctx, db, idFranqueado, false)
}

func queryEventoTesteTerminal(ctx context.Context, db *sql.DB, idFranqueado string, requireSnapshot bool) (*terminalTesteRow, error) {
	snapshotFilter := ""
	if requireSnapshot {
		snapshotFilter = ` AND COALESCE(NULLIF(TRIM(e.snapshot_url), ''), '') <> ''`
	}

	var row terminalTesteRow
	var conta, part, zona, camNome sql.NullString
	err := db.QueryRowContext(ctx, fmt.Sprintf(`
SELECT e.id, e.conta, e.particao,
       COALESCE(NULLIF(TRIM(c.zonauser), ''), NULLIF(TRIM(e.canal), ''), '001') AS zonauser,
       COALESCE(NULLIF(TRIM(c.nome), ''), '') AS camera_nome
FROM vis_evento e
LEFT JOIN vis_camera c ON c.id = e.vis_camera_id
WHERE e.id_franqueado = $1%s
ORDER BY e.id DESC
LIMIT 1`, snapshotFilter), idFranqueado).Scan(
		&row.EventoID, &conta, &part, &zona, &camNome,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	row.Conta = sqlString(conta)
	row.Particao = sqlString(part)
	row.ZonaUser = sqlString(zona)
	row.CameraNome = sqlString(camNome)
	return &row, nil
}
