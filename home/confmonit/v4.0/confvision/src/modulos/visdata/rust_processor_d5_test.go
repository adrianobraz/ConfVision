package visdata

import (
	"os"
	"testing"
)

func TestParseRustProcessorRegistry_pipeFormat(t *testing.T) {
	t.Setenv("RUST_PROCESSOR_BASE_URLS", " srv-a|https://rust-a.example/,srv-b|https://rust-b.example ")
	reg := parseRustProcessorRegistry()
	if len(reg) != 2 {
		t.Fatalf("len=%d want 2", len(reg))
	}
	if reg[0].ServidorID != "srv-a" || reg[0].BaseURL != "https://rust-a.example" {
		t.Fatalf("reg0=%+v", reg[0])
	}
	if reg[1].ServidorID != "srv-b" {
		t.Fatalf("reg1=%+v", reg[1])
	}
}

func TestFilterRegistryByServidor(t *testing.T) {
	reg := []rustProcessorRegistryEntry{
		{ServidorID: "h1", BaseURL: "https://a"},
		{ServidorID: "h2", BaseURL: "https://b"},
	}
	f := filterRegistryByServidor(reg, "h2")
	if len(f) != 1 || f[0].BaseURL != "https://b" {
		t.Fatalf("filter=%+v", f)
	}
	fAll := filterRegistryByServidor(reg, "")
	if len(fAll) != 2 {
		t.Fatalf("empty filter want 2 got %d", len(fAll))
	}
}

func TestParseRustProcessorRegistry_plainURL(t *testing.T) {
	os.Unsetenv("RUST_PROCESSOR_BASE_URLS")
	// default from d5DefaultProcessorURLs when empty - skip unset behavior
	t.Setenv("RUST_PROCESSOR_BASE_URLS", "https://only.example")
	reg := parseRustProcessorRegistry()
	if len(reg) != 1 || reg[0].ServidorID != "" || reg[0].BaseURL != "https://only.example" {
		t.Fatalf("reg=%+v", reg)
	}
}
