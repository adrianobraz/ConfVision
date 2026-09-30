package visdata

import "testing"

func TestPrepareSetorQueryParams(t *testing.T) {
	part, zona := prepareSetorQueryParams("1", "1")
	if part != "01" {
		t.Fatalf("particao=%q want 01", part)
	}
	if zona != "001" {
		t.Fatalf("zonauser=%q want 001", zona)
	}

	part, zona = prepareSetorQueryParams("01", "001")
	if part != "01" || zona != "001" {
		t.Fatalf("particao=%q zonauser=%q", part, zona)
	}

	_, zona = prepareSetorQueryParams("01", "MURO")
	if zona != "MURO" {
		t.Fatalf("zonauser texto=%q want MURO", zona)
	}
}

func TestCameraStreamPathEmptyWithoutSecret(t *testing.T) {
	if CameraStreamPath(2) != "" {
		t.Fatal("expected empty path without RTMP_PUBLISH_SECRET in test env")
	}
}
