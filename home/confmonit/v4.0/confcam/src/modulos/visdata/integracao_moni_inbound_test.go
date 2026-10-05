package visdata

import "testing"

func TestMoniCustomerCodeVariants(t *testing.T) {
	got := moniCustomerCodeVariants(530)
	want := []string{"0530", "530"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestMoniEmpresaCodeVariants(t *testing.T) {
	got := moniEmpresaCodeVariants("1")
	seen := map[string]bool{}
	for _, s := range got {
		seen[s] = true
	}
	for _, want := range []string{"001", "01", "1"} {
		if !seen[want] {
			t.Fatalf("missing variant %q in %v", want, got)
		}
	}
}

func TestMoniCompanyCodeVariantsUsesIntegracaoFallback(t *testing.T) {
	cfg := &integracaoConfig{EmpresaCodigo: "001"}
	got := moniCompanyCodeVariants(moniInboundEvent{}, cfg)
	if len(got) == 0 || got[0] != "001" {
		t.Fatalf("got %v want first 001", got)
	}
}

func TestResolveMoniInboundAcaoToggleCode(t *testing.T) {
	cfg := &integracaoConfig{
		CodigoEventoArmar:   "402",
		CodigoEventoDesarmar: "402",
	}
	ev := moniInboundEvent{EventSecondCode: "402"}
	acao, ok := resolveMoniInboundAcao(cfg, ev)
	if !ok || acao != "toggle" {
		t.Fatalf("got acao=%q ok=%v want toggle", acao, ok)
	}
}

func TestResolveMoniInboundAcaoDistinctCodes(t *testing.T) {
	cfg := &integracaoConfig{
		CodigoEventoArmar:   "130",
		CodigoEventoDesarmar: "131",
	}
	acaoArmar, ok := resolveMoniInboundAcao(cfg, moniInboundEvent{EventSecondCode: "130"})
	if !ok || acaoArmar != "ativar" {
		t.Fatalf("armar: got %q ok=%v", acaoArmar, ok)
	}
	acaoDesarmar, ok := resolveMoniInboundAcao(cfg, moniInboundEvent{EventSecondCode: "131"})
	if !ok || acaoDesarmar != "desativar" {
		t.Fatalf("desarmar: got %q ok=%v", acaoDesarmar, ok)
	}
}

func TestMoniInboundEstaAtivo(t *testing.T) {
	plano := "analitico_24h_foto"
	pausada := true
	retomada := false

	if moniInboundEstaAtivo([]map[string]any{
		{"plano": plano, "analitico_pausado": pausada},
	}) {
		t.Fatal("esperado inativo com camera pausada")
	}
	if !moniInboundEstaAtivo([]map[string]any{
		{"plano": plano, "analitico_pausado": retomada},
	}) {
		t.Fatal("esperado ativo com camera retomada")
	}
	if !moniInboundEstaAtivo([]map[string]any{
		{"plano": plano, "analitico_pausado": pausada},
		{"plano": plano, "analitico_pausado": retomada},
	}) {
		t.Fatal("esperado ativo se alguma camera retomada")
	}
}
