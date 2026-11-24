package config

import "testing"

func TestBindPublicoSemTLSFalha(t *testing.T) {
	cfg := Config{Listen: "0.0.0.0:8090"}
	if err := cfg.ValidateListen(); err == nil {
		t.Fatal("bind publico sem tls foi aceito")
	}
}

func TestBindAtrasDoProxy(t *testing.T) {
	cfg := Config{Listen: "0.0.0.0:8090", BehindProxy: true}
	if err := cfg.ValidateListen(); err != nil {
		t.Fatal(err)
	}
}

func TestLoopbackSemTLS(t *testing.T) {
	cfg := Config{Listen: "127.0.0.1:8090"}
	if err := cfg.ValidateListen(); err != nil {
		t.Fatal(err)
	}
}
