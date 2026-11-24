package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"confvision-infrastructure-monitor/internal/collect"
	"confvision-infrastructure-monitor/internal/config"
	"confvision-infrastructure-monitor/internal/model"
)

type Server struct {
	cfg config.Config
	eng *collect.Engine
}

func New(cfg config.Config, eng *collect.Engine) *Server {
	return &Server{cfg: cfg, eng: eng}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/overview", s.auth(s.overview))
	mux.HandleFunc("GET /api/v1/servers", s.auth(s.servers))
	mux.HandleFunc("GET /api/v1/servers/{id}", s.auth(s.server))
	mux.HandleFunc("GET /api/v1/containers", s.auth(s.containers))
	mux.HandleFunc("GET /api/v1/services", s.auth(s.services))
	mux.HandleFunc("GET /api/v1/processes", s.auth(s.processes))
	mux.HandleFunc("GET /api/v1/cameras", s.auth(s.cameras))
	mux.HandleFunc("GET /api/v1/history", s.auth(s.history))
	mux.HandleFunc("GET /api/v1/capacity", s.auth(s.capacity))
	mux.HandleFunc("GET /api/v1/alerts", s.auth(s.alerts))
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	snap, ok := s.eng.Snapshot()
	status := "sem_coleta"
	if ok {
		status = "ok"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      status,
		"coletado_em": snap.ColetadoEm,
	})
}

func (s *Server) overview(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current())
}

func (s *Server) servers(w http.ResponseWriter, _ *http.Request) {
	snap := s.current()
	writeJSON(w, http.StatusOK, []model.Server{snap.Servidor})
}

func (s *Server) server(w http.ResponseWriter, r *http.Request) {
	snap := s.current()
	if r.PathValue("id") != snap.Servidor.ID {
		writeJSON(w, http.StatusNotFound, map[string]string{"erro": "servidor nao coletado"})
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) containers(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current().Containers)
}

func (s *Server) services(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current().Servicos)
}

func (s *Server) processes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current().Processos)
}

func (s *Server) cameras(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current().Cameras)
}

func (s *Server) capacity(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current().Capacidade)
}

func (s *Server) alerts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.current().Alertas)
}

func (s *Server) history(w http.ResponseWriter, _ *http.Request) {
	snap := s.current()
	writeJSON(w, http.StatusOK, model.HistoryResponse{
		Persistido:   false,
		Motivo:       "o monitor nao persiste historico; a amostra atual nao e serie historica",
		AmostraAtual: &snap,
		Series:       []struct{}{},
	})
}

func (s *Server) current() model.Snapshot {
	snap, ok := s.eng.Snapshot()
	if ok {
		return snap
	}
	return model.Snapshot{
		Servidor:            model.Server{ID: s.cfg.ServerID, Nome: s.cfg.ServerNome},
		HistoricoPersistido: false,
		Coleta:              []model.Reading{model.Unavailable("coleta", "ainda nao houve coleta")},
	}
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorized(bearer(r), s.cfg.Secret) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"erro": "nao autorizado"})
			return
		}
		next(w, r)
	}
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

func authorized(got, want string) bool {
	if want == "" || got == "" {
		return false
	}
	a := sha256.Sum256([]byte(got))
	b := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("nivel=erro evento=json")
	}
}
