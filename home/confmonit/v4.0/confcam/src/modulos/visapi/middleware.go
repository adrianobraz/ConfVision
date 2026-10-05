package visapi

import (
	"confvision/src/config"
	"net/http"
	"strings"
)

func workerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(config.VisWorkerAPIKey)
		if key == "" {
			next(w, r)
			return
		}

		got := strings.TrimSpace(r.Header.Get("X-Vis-Worker-Key"))
		if got == "" {
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				got = strings.TrimSpace(auth[7:])
			}
		}

		if got != key {
			receptorKey := strings.TrimSpace(config.VisReceptorBearer)
			if receptorKey == "" || got != receptorKey {
				http.Error(w, `{"erro":"nao autorizado"}`, http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	}
}
