package visdata

import (
	"context"
	"testing"
	"time"
)

func TestAplicarAtivacaoFallbackRefIDVazio(t *testing.T) {
	ctx := context.Background()
	orig := ativarLicencaPagamentoFn
	t.Cleanup(func() { ativarLicencaPagamentoFn = orig })

	var got []int
	ativarLicencaPagamentoFn = func(ctx context.Context, licID, faturaID, pagamentoID int, pagoEm time.Time) (map[string]any, error) {
		got = append(got, licID)
		return map[string]any{"id": licID}, nil
	}

	origList := listPendenteLicencaIDsFn
	t.Cleanup(func() { listPendenteLicencaIDsFn = origList })
	listPendenteLicencaIDsFn = func(ctx context.Context, idFranqueado string) ([]int, error) {
		return []int{10, 11, 12}, nil
	}

	faturas := []map[string]any{
		{
			"id": 50, "tipo": "confvision_venda", "id_franqueado": "fra1", "pago_em": "2026-10-05T12:00:00Z",
			"itens": []any{
				map[string]any{"ref_tipo": "vis_licenca", "ref_id": ""},
				map[string]any{"ref_tipo": "vis_licenca", "ref_id": ""},
			},
		},
	}

	n, errs := aplicarAtivacaoLicencasDeFaturasVerbose(ctx, faturas)
	if len(errs) != 0 {
		t.Fatalf("errs=%v", errs)
	}
	if n != 2 {
		t.Fatalf("ativadas=%d want 2", n)
	}
	if len(got) != 2 || got[0] != 10 || got[1] != 11 {
		t.Fatalf("licencas ativadas=%v want [10 11]", got)
	}
}
