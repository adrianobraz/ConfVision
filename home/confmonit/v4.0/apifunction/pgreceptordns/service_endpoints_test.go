package pgreceptordns

import "testing"

func TestSubdominioPadraoPorModulo(t *testing.T) {
	cases := map[string]string{
		"JFL":       "jfl",
		"INTELBRAS": "intelbras",
		"VETTI":     "vetti",
		"COMPATEC":  "compatec",
		"OUTRO":     "",
	}
	for mod, want := range cases {
		if got := subdominioPadraoPorModulo(mod); got != want {
			t.Fatalf("modulo %q: got %q want %q", mod, got, want)
		}
	}
}

func TestFabricanteEhCamera(t *testing.T) {
	if !fabricanteEhCamera("CAMERA") || !fabricanteEhCamera(" camera ") {
		t.Fatal("deveria reconhecer fabricante camera")
	}
	if fabricanteEhCamera("JFL") {
		t.Fatal("JFL nao e camera")
	}
}

func TestHostParaEndpointPrioridade(t *testing.T) {
	s := &Service{zona: "dnsid.com.br"}
	reg := &Registro{Status: "ativo", FQDN: "minha.dnsid.com.br"}
	if got := s.hostParaEndpoint(reg, "JFL"); got != "minha.dnsid.com.br" {
		t.Fatalf("dns proprio: got %q", got)
	}
	if got := s.hostParaEndpoint(nil, "JFL"); got != "jfl.dnsid.com.br" {
		t.Fatalf("fallback JFL: got %q", got)
	}
	if got := s.hostParaEndpoint(nil, "INTELBRAS"); got != "intelbras.dnsid.com.br" {
		t.Fatalf("fallback Intelbras: got %q", got)
	}
}
