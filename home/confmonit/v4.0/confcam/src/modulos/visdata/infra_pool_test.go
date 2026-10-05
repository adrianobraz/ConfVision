package visdata

import "testing"

func TestCameraNeedsWorkerID(t *testing.T) {
	if CameraNeedsWorkerID("online", false, false) {
		t.Fatal("plano online nao exige worker")
	}
	if !CameraNeedsWorkerID("analitico_24h_foto", true, true) {
		t.Fatal("analitico exige worker")
	}
	if !CameraNeedsWorkerID("sensor", false, true) {
		t.Fatal("captura analitico exige worker")
	}
}

func TestInfraPoolActiveFlags(t *testing.T) {
	mtx := []map[string]any{{"ativo": false}, {"ativo": true}}
	if !infraPoolHasActiveMtx(mtx) {
		t.Fatal("expected active mtx row")
	}
	wrk := []map[string]any{
		{"ativo": true, "worker_tipo": "dvr"},
		{"ativo": true, "worker_tipo": "analitico"},
	}
	if !infraPoolHasActiveWorkerAnalitico(wrk) {
		t.Fatal("expected active analitico worker")
	}
	if infraPoolHasActiveWorkerAnalitico([]map[string]any{{"ativo": true, "worker_tipo": "motion"}}) {
		t.Fatal("motion should not count")
	}
}
