package alert

import (
	"testing"
	"time"

	"confvision-infrastructure-monitor/internal/model"
)

func TestOcorrenciaAbertaNaoDuplica(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	levels := map[string]string{"cpu": Alerta}
	first, state := Evaluate(nil, now, levels, nil)
	if len(first) != 1 || first[0].Estado != Alerta {
		t.Fatalf("%+v", first)
	}
	second, state := Evaluate(state, now.Add(time.Second), levels, nil)
	if len(second) != 1 || !second[0].Desde.Equal(first[0].Desde) {
		t.Fatalf("duplicou %+v", second)
	}
	_ = state
}

func TestRecuperacao(t *testing.T) {
	now := time.Now().UTC()
	_, state := Evaluate(nil, now, map[string]string{"cpu": Alerta}, nil)
	got, _ := Evaluate(state, now.Add(time.Second), map[string]string{"cpu": Normal}, nil)
	if len(got) != 1 || got[0].Estado != Recuperacao {
		t.Fatalf("%+v", got)
	}
}

func TestLevelNaoUsaZeroComoAlerta(t *testing.T) {
	if Level(0, 70, 85) != Normal {
		t.Fatal("zero virou alerta")
	}
	if Level(90, 70, 85) != Alerta {
		t.Fatal("limiar")
	}
	reading := model.Unavailable("cpu_pct", "ausente")
	if reading.Valor != nil {
		t.Fatal("valor ficticio")
	}
}
