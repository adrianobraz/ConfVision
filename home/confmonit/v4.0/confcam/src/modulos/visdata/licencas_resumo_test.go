package visdata

import (
	"context"
	"testing"
)

func TestResumoLicencasByFranqueadoRequerFranqueado(t *testing.T) {
	_, err := ResumoLicencasByFranqueado(context.Background(), "", "")
	if err == nil {
		t.Fatal("esperava erro sem id_franqueado")
	}
}
