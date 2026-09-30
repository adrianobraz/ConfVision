package auth

import (
	"net/http/httptest"
	"testing"
)

func TestRequireEscopoBreakglassUsaHeaderCentral(t *testing.T) {
	rva := "a1b2c3d4e5f6789012345678abcdef01"
	req := httptest.NewRequest("POST", "/financeiro/catalogo/listar", nil)
	req.Header.Set("X-Breakglass", "S")
	req.Header.Set("X-Adm-Central", rva)
	req.Header.Set("X-Adm-Token", "adm_test")

	sess, err := RequireEscopo(req)
	if err != nil {
		t.Fatalf("RequireEscopo: %v", err)
	}
	if sess.IDCentralUUID != rva {
		t.Fatalf("esperava IDCentralUUID=%s, obteve %s", rva, sess.IDCentralUUID)
	}
	if sess.UserTipo != "CEN" {
		t.Fatalf("esperava userTipo CEN, obteve %s", sess.UserTipo)
	}
}

func TestIDCentralFinanceiroRejeitaCentralLegada(t *testing.T) {
	s := SessaoAdm{IDCentralUUID: "CENTRAL", IDCentralCatalogo: "CENTRAL"}
	if id := s.IDCentralFinanceiro(); id != "" {
		t.Fatalf("nao deveria aceitar CENTRAL como UUID financeiro, obteve %q", id)
	}
}
