package visdata

import (
	"confvision/src/config"
	"testing"
)

func TestCodigoImagemPublicaRoundTrip(t *testing.T) {
	config.ImagemPublicSecret = "test-secret-imagem-2026"

	codigo := CodigoImagemPublica(8588)
	if codigo == "" {
		t.Fatal("codigo vazio")
	}
	if len(codigo) < imagemHashMinLength {
		t.Fatalf("codigo curto: %q", codigo)
	}
	id, ok := ParseCodigoImagemPublica(codigo)
	if !ok || id != 8588 {
		t.Fatalf("decode falhou: id=%d ok=%v", id, ok)
	}
}

func TestComplementoImagemMoni(t *testing.T) {
	got := ComplementoImagemMoni("046", "abc123def456")
	want := "046|1|1|abc123def456?"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
