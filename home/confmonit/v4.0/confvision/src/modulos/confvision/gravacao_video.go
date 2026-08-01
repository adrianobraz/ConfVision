package confvision

import (
	"confvision/src/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type gravacaoStorageResumo struct {
	S3Endpoint string `json:"s3_endpoint"`
	S3Bucket   string `json:"s3_bucket"`
}

func ProxyGravacaoSegmentoVideo(w http.ResponseWriter, r *http.Request) {
	idFranqueado := strings.TrimSpace(r.URL.Query().Get("id_franqueado"))
	videoURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if idFranqueado == "" || videoURL == "" {
		http.Error(w, "id_franqueado e url são obrigatórios", http.StatusBadRequest)
		return
	}

	parsed, err := url.Parse(videoURL)
	if err != nil || parsed.Scheme != "https" {
		http.Error(w, "URL de vídeo inválida", http.StatusBadRequest)
		return
	}

	ok, err := gravacaoURLAutorizada(idFranqueado, videoURL)
	if err != nil {
		http.Error(w, "Erro ao validar storage", http.StatusBadGateway)
		return
	}
	if !ok {
		http.Error(w, "URL não autorizada para este franqueado", http.StatusForbidden)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, videoURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if rangeHdr := r.Header.Get("Range"); rangeHdr != "" {
		req.Header.Set("Range", rangeHdr)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Falha ao buscar vídeo no storage", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "ETag", "Last-Modified"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "video/mp4")
	}
	w.Header().Set("Cache-Control", "private, max-age=3600")

	if r.URL.Query().Get("download") == "1" {
		nome := sanitizarNomeArquivoVideo(r.URL.Query().Get("nome"))
		if nome == "" {
			nome = "gravacao.mp4"
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", nome))
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func sanitizarNomeArquivoVideo(nome string) string {
	nome = strings.TrimSpace(nome)
	if nome == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range nome {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	limpo := strings.Trim(b.String(), "-.")
	if limpo == "" {
		return ""
	}
	if !strings.HasSuffix(strings.ToLower(limpo), ".mp4") {
		limpo += ".mp4"
	}
	return limpo
}

func gravacaoURLAutorizada(idFranqueado, videoURL string) (bool, error) {
	storages, err := listarGravacaoStorageFranqueado(idFranqueado)
	if err != nil {
		return false, err
	}
	if len(storages) == 0 {
		return false, nil
	}

	lowerURL := strings.ToLower(videoURL)
	for _, st := range storages {
		bucket := strings.ToLower(strings.TrimSpace(st.S3Bucket))
		endpoint := strings.ToLower(strings.TrimSpace(st.S3Endpoint))
		if bucket != "" && strings.Contains(lowerURL, bucket) {
			return true, nil
		}
		if endpoint != "" {
			host := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
			host = strings.TrimSuffix(host, "/")
			if host != "" && strings.Contains(lowerURL, host) {
				return true, nil
			}
		}
	}
	return false, nil
}

func listarGravacaoStorageFranqueado(idFranqueado string) ([]gravacaoStorageResumo, error) {
	path := fmt.Sprintf("%s/vis_gravacao_storage_by_franqueado?id_franqueado=%s&status=ativo",
		config.XanoBaseUrl, url.QueryEscape(idFranqueado))

	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("storage api status %d", resp.StatusCode)
	}

	var envelope struct {
		Dados json.RawMessage `json:"dados"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}

	var lista []gravacaoStorageResumo
	if len(envelope.Dados) > 0 && envelope.Dados[0] == '[' {
		if err := json.Unmarshal(envelope.Dados, &lista); err != nil {
			return nil, err
		}
		return lista, nil
	}

	var unico gravacaoStorageResumo
	if err := json.Unmarshal(envelope.Dados, &unico); err == nil && (unico.S3Bucket != "" || unico.S3Endpoint != "") {
		return []gravacaoStorageResumo{unico}, nil
	}

	return lista, nil
}
