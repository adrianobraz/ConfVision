package confvision

import (
	"confvision/src/modulos/visdata"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func ServirImagemPublica(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}
	if !imagemRateLimitAllow(r) {
		http.Error(w, "limite de requisicoes excedido", http.StatusTooManyRequests)
		return
	}

	codigo := strings.TrimSpace(mux.Vars(r)["codigo"])
	if codigo == "" {
		http.NotFound(w, r)
		return
	}

	eventoID, ok := visdata.ParseCodigoImagemPublica(codigo)
	if !ok {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	snapURL, err := visdata.SnapshotURLImagemLiberada(ctx, eventoID, codigo)
	if err != nil || snapURL == "" {
		http.NotFound(w, r)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, snapURL, nil)
	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[IMAGEM] fetch evento=%d url=%s: %v", eventoID, snapURL, err)
		http.Error(w, "imagem indisponivel", http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		log.Printf("[IMAGEM] fetch HTTP %d evento=%d", res.StatusCode, eventoID)
		http.NotFound(w, r)
		return
	}

	ct := res.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, res.Body); err != nil {
		log.Printf("[IMAGEM] stream evento=%d: %v", eventoID, err)
	}
}

func ServirImagemPublicaHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok " + time.Now().UTC().Format(time.RFC3339)))
}
