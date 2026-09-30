package pgfinmirror

import (
	"context"
	"fmt"
	"strings"
	"time"

	"apifunction/pgcredito"
)

type HookEntity struct {
	Tipo  string         `json:"tipo"`
	Acao  string         `json:"acao"`
	Dados map[string]any `json:"dados"`
}

type HookResult struct {
	Processados int   `json:"processados"`
	Ignorados   int   `json:"ignorados"`
	DuracaoMs   int64 `json:"duracao_ms"`
}

func ApplyHook(ctx context.Context, entidades []HookEntity) (HookResult, error) {
	start := time.Now()
	res := HookResult{}

	db, err := pgcredito.DB()
	if err != nil {
		return res, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer tx.Rollback()

	for _, ent := range entidades {
		tipo := strings.ToLower(strings.TrimSpace(ent.Tipo))
		acao := strings.ToLower(strings.TrimSpace(ent.Acao))
		if acao == "" {
			acao = "upsert"
		}
		if ent.Dados == nil {
			res.Ignorados++
			continue
		}

		var applyErr error
		switch tipo {
		case "fatura":
			if acao == "delete" {
				applyErr = deleteFatura(ctx, tx, intFromAny(ent.Dados["id"]))
			} else {
				applyErr = upsertFaturaOne(ctx, tx, ent.Dados)
			}
		case "pagamento":
			if acao == "delete" {
				applyErr = deletePagamento(ctx, tx, intFromAny(ent.Dados["id"]))
			} else {
				applyErr = upsertPagamentoOne(ctx, tx, ent.Dados)
			}
		case "assinatura":
			if acao == "delete" {
				applyErr = deleteAssinatura(ctx, tx, intFromAny(ent.Dados["id"]))
			} else {
				applyErr = upsertAssinaturaOne(ctx, tx, ent.Dados)
			}
		case "caixa", "caixa_movimento":
			applyErr = upsertCaixaOne(ctx, tx, ent.Dados)
		case "conta_pagar":
			applyErr = upsertContaPagarOne(ctx, tx, ent.Dados)
		default:
			res.Ignorados++
			continue
		}
		if applyErr != nil {
			return res, fmt.Errorf("%s id=%v: %w", tipo, ent.Dados["id"], applyErr)
		}
		res.Processados++
	}

	if err := setSyncState(ctx, tx, "last_hook_at", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return res, err
	}
	if err := tx.Commit(); err != nil {
		return res, err
	}

	res.DuracaoMs = time.Since(start).Milliseconds()
	return res, nil
}
