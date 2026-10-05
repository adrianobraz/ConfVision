package visdata

import "testing"

func TestIntegracaoSistemaPermiteDispatch(t *testing.T) {
	cases := []struct {
		found  bool
		sistema string
		ativo  bool
		want   bool
	}{
		{false, "", false, false},
		{true, "nenhum", false, false},
		{true, "nenhum", true, false},
		{true, "confmonit", false, false},
		{true, "confmonit", true, true},
		{true, "moni", true, true},
		{true, "moni", false, false},
	}
	for _, c := range cases {
		got := integracaoSistemaPermiteDispatch(c.found, c.sistema, c.ativo)
		if got != c.want {
			t.Fatalf("found=%v sistema=%q ativo=%v got=%v want=%v", c.found, c.sistema, c.ativo, got, c.want)
		}
	}
}

func TestNormalizeDuplaComunicacao(t *testing.T) {
	cases := []struct {
		sistema string
		dupla   bool
		want    bool
	}{
		{"nenhum", true, false},
		{"confmonit", true, false},
		{"moni", false, false},
		{"moni", true, true},
		{"dguard", true, true},
		{"segware", true, true},
	}
	for _, c := range cases {
		got := normalizeDuplaComunicacao(c.sistema, c.dupla)
		if got != c.want {
			t.Fatalf("sistema=%q dupla=%v got=%v want=%v", c.sistema, c.dupla, got, c.want)
		}
	}
}

func TestIntegracaoDuplaComunicacaoHabilitada(t *testing.T) {
	if !integracaoDuplaComunicacaoHabilitada("moni") {
		t.Fatal("moni deveria habilitar dupla")
	}
	if integracaoDuplaComunicacaoHabilitada("confmonit") {
		t.Fatal("confmonit nao deveria habilitar dupla")
	}
}
