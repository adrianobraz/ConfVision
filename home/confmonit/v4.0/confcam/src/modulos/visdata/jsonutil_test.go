package visdata

import "testing"

func TestIntValPostgresInt64(t *testing.T) {
	m := map[string]any{"vis_mediamtx_node_id": int64(3)}
	if got := intVal(m, "vis_mediamtx_node_id"); got != 3 {
		t.Fatalf("intVal int64: got %d want 3", got)
	}
}

func TestIntValMissing(t *testing.T) {
	if got := intVal(nil, "x"); got != 0 {
		t.Fatalf("intVal nil map: got %d want 0", got)
	}
	if got := intVal(map[string]any{}, "x"); got != 0 {
		t.Fatalf("intVal missing key: got %d want 0", got)
	}
}
