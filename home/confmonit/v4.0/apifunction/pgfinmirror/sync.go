package pgfinmirror

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"apifunction/config"
	"apifunction/pgcredito"
	"apifunction/xano"
)

const (
	metaTableFatura     = 121
	metaTablePagamento  = 123
	metaTableAssinatura = 120
	metaTableCaixa      = 132
	metaTableContaPagar = 133
)

type SyncResult struct {
	Faturas     int    `json:"faturas"`
	Pagamentos  int    `json:"pagamentos"`
	Assinaturas int    `json:"assinaturas"`
	Caixa       int    `json:"caixa"`
	ContasPagar int    `json:"contas_pagar"`
	DuracaoMs   int64  `json:"duracao_ms"`
	Erro        string `json:"erro,omitempty"`
}

func SyncFromMeta(ctx context.Context) (SyncResult, error) {
	start := time.Now()
	res := SyncResult{}

	cli := xano.NewMeta(config.XanoMetaBaseURL, config.XanoMetaAccessToken, config.XanoMetaWorkspaceID)
	if cli == nil || !cli.Enabled() {
		return res, fmt.Errorf("configure XANO_META_ACCESS_TOKEN no apifunction")
	}

	db, err := pgcredito.DB()
	if err != nil {
		return res, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	rawFaturas, err := cli.ListTableContent(metaTableFatura)
	if err != nil {
		return res, fmt.Errorf("meta faturas: %w", err)
	}
	if err := upsertFaturas(ctx, tx, rawFaturas); err != nil {
		return res, err
	}
	res.Faturas = len(rawFaturas)

	rawPags, err := cli.ListTableContent(metaTablePagamento)
	if err != nil {
		return res, fmt.Errorf("meta pagamentos: %w", err)
	}
	if err := upsertPagamentos(ctx, tx, rawPags); err != nil {
		return res, err
	}
	res.Pagamentos = len(rawPags)

	rawAss, err := cli.ListTableContent(metaTableAssinatura)
	if err != nil {
		return res, fmt.Errorf("meta assinaturas: %w", err)
	}
	if err := upsertAssinaturas(ctx, tx, rawAss); err != nil {
		return res, err
	}
	res.Assinaturas = len(rawAss)

	rawCaixa, err := cli.ListTableContent(metaTableCaixa)
	if err != nil {
		return res, fmt.Errorf("meta caixa: %w", err)
	}
	if err := upsertCaixa(ctx, tx, rawCaixa); err != nil {
		return res, err
	}
	res.Caixa = len(rawCaixa)

	rawCP, err := cli.ListTableContent(metaTableContaPagar)
	if err != nil {
		return res, fmt.Errorf("meta contas pagar: %w", err)
	}
	if err := upsertContasPagar(ctx, tx, rawCP); err != nil {
		return res, err
	}
	res.ContasPagar = len(rawCP)

	if err := setSyncState(ctx, tx, "last_sync_at", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return res, err
	}
	if err := tx.Commit(); err != nil {
		return res, err
	}

	res.DuracaoMs = time.Since(start).Milliseconds()
	return res, nil
}

func IsReady(ctx context.Context) (bool, string, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return false, "postgres nao configurado", err
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fp_fin_fatura`).Scan(&n); err != nil {
		return false, "espelho nao inicializado", err
	}
	if n == 0 {
		return false, "espelho vazio — aguardando primeiro sync", nil
	}
	last, _ := GetSyncState(ctx, "last_sync_at")
	if last == "" {
		return false, "sync nunca concluido", nil
	}
	t, err := time.Parse(time.RFC3339, last)
	if err != nil {
		return true, "", nil
	}
	if time.Since(t) > 48*time.Hour {
		return false, "espelho desatualizado (>48h)", nil
	}
	return true, "", nil
}

func GetSyncState(ctx context.Context, chave string) (string, error) {
	db, err := pgcredito.DB()
	if err != nil {
		return "", err
	}
	var val sql.NullString
	err = db.QueryRowContext(ctx, `SELECT valor FROM fp_fin_sync_state WHERE chave = $1`, chave).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val.String), nil
}

func setSyncState(ctx context.Context, tx *sql.Tx, chave, valor string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO fp_fin_sync_state (chave, valor, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (chave) DO UPDATE SET valor = EXCLUDED.valor, updated_at = NOW()
	`, chave, valor)
	return err
}

func LogShadowDiff(ctx context.Context, escopoTipo, escopoID, competencia, campo, pgVal, xanoVal string) error {
	db, err := pgcredito.DB()
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO fp_fin_shadow_diff (escopo_tipo, escopo_id, competencia, campo, valor_postgres, valor_xano)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, escopoTipo, escopoID, competencia, campo, pgVal, xanoVal)
	return err
}
