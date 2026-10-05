package visdata

import (
	"database/sql"
	"testing"
)

func TestBoolValAtivoJSON(t *testing.T) {
	cases := []struct {
		key string
		val any
		want *bool
	}{
		{"ativo", true, ptrBool(true)},
		{"ativo", false, ptrBool(false)},
		{"ativo", float64(1), ptrBool(true)},
		{"ativo", float64(0), ptrBool(false)},
		{"ativo", "S", ptrBool(true)},
		{"ativo", "N", ptrBool(false)},
		{"ativo", "false", ptrBool(false)},
		{"ativo", "0", ptrBool(false)},
	}
	for _, c := range cases {
		got := boolVal(map[string]any{c.key: c.val}, c.key)
		if got == nil || c.want == nil {
			if got != c.want {
				t.Fatalf("val=%v got=%v want=%v", c.val, got, c.want)
			}
			continue
		}
		if *got != *c.want {
			t.Fatalf("val=%v got=%v want=%v", c.val, *got, *c.want)
		}
	}
}

func ptrBool(b bool) *bool { return &b }

func TestCvgEscopoClienteMultiDispositivo(t *testing.T) {
	one := []map[string]any{
		{"id": 1, "id_dispositivo": "A"},
		{"id": 2, "id_dispositivo": "A"},
	}
	if cvgEscopoClienteMultiDispositivo(one) {
		t.Fatal("mesmo dispositivo nao deve ser multi")
	}
	two := []map[string]any{
		{"id": 1, "id_dispositivo": "A"},
		{"id": 2, "id_dispositivo": "B"},
	}
	if !cvgEscopoClienteMultiDispositivo(two) {
		t.Fatal("dois dispositivos deve ser multi")
	}
}

func TestResolveCameraAtivoDeteccaoRespeitaInputFalse(t *testing.T) {
	input := map[string]any{"ativo": false}
	ativo, _, _, _ := resolveCameraAtivoDeteccao(
		"analitico_24h_foto",
		PlanoFlagsFrom("analitico_24h_foto"),
		input,
		sql.NullTime{},
	)
	if ativo {
		t.Fatal("ativo deve permanecer false quando informado no payload")
	}
}
