package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"confvision-infrastructure-monitor/internal/collect"
	"confvision-infrastructure-monitor/internal/config"
	"confvision-infrastructure-monitor/internal/model"
)

func TestAPIExigeSegredoENaoPersisteHistorico(t *testing.T) {
	cfg := config.Config{ServerID: "vps-teste", ServerNome: "teste", Secret: "segredo-teste", Interval: time.Second}
	eng := collect.NewEngine(cfg)
	eng.Refresh(context.Background())
	srv := New(cfg, eng)
	handler := srv.Handler()

	sem := httptest.NewRecorder()
	handler.ServeHTTP(sem, httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil))
	if sem.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", sem.Code)
	}

	errado := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/overview", nil)
	req.Header.Set("Authorization", "Bearer outro")
	handler.ServeHTTP(errado, req)
	if errado.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", errado.Code)
	}

	ok := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/history", nil)
	req.Header.Set("Authorization", "Bearer segredo-teste")
	handler.ServeHTTP(ok, req)
	if ok.Code != http.StatusOK {
		t.Fatalf("status %d body %s", ok.Code, ok.Body.String())
	}
	var body model.HistoryResponse
	if err := json.Unmarshal(ok.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Persistido || body.Motivo == "" || len(body.Series) != 0 {
		t.Fatalf("%+v", body)
	}

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK || jsonContainsSecret(health.Body.String()) {
		t.Fatalf("health %s", health.Body.String())
	}
}

func TestCapacidadeSemCotaNaoProjeta(t *testing.T) {
	cfg := config.Config{ServerID: "vps-teste", Secret: "segredo-teste", Interval: time.Second}
	eng := collect.NewEngine(cfg)
	eng.Refresh(context.Background())
	snap, ok := eng.Snapshot()
	if !ok {
		t.Fatal("sem snapshot")
	}
	if snap.Capacidade.CPUProjetada.Origem != model.OrigemIndisponivel || snap.Capacidade.CPUProjetada.Valor != nil {
		t.Fatalf("%+v", snap.Capacidade.CPUProjetada)
	}
}

func jsonContainsSecret(body string) bool {
	return len(body) > 0 && (contains(body, "segredo-teste"))
}

func contains(s, part string) bool {
	return len(part) > 0 && len(s) >= len(part) && (s == part || len(s) > 0 && (func() bool {
		for i := 0; i+len(part) <= len(s); i++ {
			if s[i:i+len(part)] == part {
				return true
			}
		}
		return false
	})())
}
