package licenca

import "testing"

func TestParseEstadoAddonsObjetoVazio(t *testing.T) {
	raw := []byte(`{
		"liberado": true,
		"motivo": "ok",
		"assinatura": {
			"plano": "lite",
			"status": "ativa",
			"valido_ate": "2026-08-10",
			"addons_json": {},
			"addons_pendentes_json": {}
		},
		"addons_efetivos": [],
		"addons_pendentes": {},
		"addons_contratados": {},
		"modulos_json": {},
		"limites_json": {"clientes_max": 50}
	}`)

	est, err := parseEstado(raw, "franqueadopro")
	if err != nil {
		t.Fatalf("parseEstado erro: %v", err)
	}
	if !est.Liberado {
		t.Fatalf("esperado liberado=true, got motivo=%s", est.Motivo)
	}
	if est.Plano != "lite" {
		t.Fatalf("plano=%s", est.Plano)
	}
}
