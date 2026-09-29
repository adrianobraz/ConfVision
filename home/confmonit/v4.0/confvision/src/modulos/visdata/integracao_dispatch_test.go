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
