package visdata

import (
	"context"
	"testing"
	"time"
)

func TestAplicarAtivacaoLicencasDeFaturasFiltraTipoConfvision(t *testing.T) {
	ctx := context.Background()
	called := 0
	orig := ativarLicencaPagamentoFn
	t.Cleanup(func() { ativarLicencaPagamentoFn = orig })

	ativarLicencaPagamentoFn = func(ctx context.Context, licID, faturaID, pagamentoID int, pagoEm time.Time) (map[string]any, error) {
		called++
		if licID != 42 || faturaID != 101 {
			t.Fatalf("unexpected ids lic=%d fat=%d", licID, faturaID)
		}
		return map[string]any{"id": licID, "status": "disponivel"}, nil
	}

	faturas := []map[string]any{
		{
			"id": 100, "tipo": "contrato_franqueado", "pago_em": "2026-10-05T12:00:00Z",
			"itens": []any{
				map[string]any{"ref_tipo": "vis_licenca", "ref_id": "42"},
			},
		},
		{
			"id": 101, "tipo": "confvision_venda", "pago_em": "2026-10-05T12:00:00Z",
			"itens": []any{
				map[string]any{"ref_tipo": "vis_licenca", "ref_id": float64(42)},
				map[string]any{"ref_tipo": "outro", "ref_id": "99"},
			},
		},
	}

	n, err := aplicarAtivacaoLicencasDeFaturas(ctx, faturas)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("ativadas=%d want 1 (so confvision_venda)", n)
	}
	if called != 1 {
		t.Fatalf("AtivarLicencaPagamento calls=%d want 1", called)
	}
}

func TestAplicarAtivacaoIgnoraItensInvalidos(t *testing.T) {
	ctx := context.Background()
	orig := ativarLicencaPagamentoFn
	t.Cleanup(func() { ativarLicencaPagamentoFn = orig })
	ativarLicencaPagamentoFn = func(ctx context.Context, licID, faturaID, pagamentoID int, pagoEm time.Time) (map[string]any, error) {
		return map[string]any{"id": licID}, nil
	}

	faturas := []map[string]any{
		{
			"id": 1, "tipo": "confvision_venda",
			"itens": []any{
				map[string]any{"ref_tipo": "vis_licenca", "ref_id": ""},
				map[string]any{"ref_tipo": "vis_licenca", "ref_id": "0"},
			},
		},
	}
	n, err := aplicarAtivacaoLicencasDeFaturas(ctx, faturas)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("ativadas=%d want 0", n)
	}
}
