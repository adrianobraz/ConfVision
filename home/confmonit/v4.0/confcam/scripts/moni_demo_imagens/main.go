// Upload de imagens demo Moni (Contabo + vis_evento + links publicos).
//
// Uso:
//
//	cd home/confmonit/v4.0/confvision
//	go run ./scripts/moni_demo_imagens [pasta_com_1.jpg_2.jpg_3.jpg]
//
// Variaveis opcionais:
//   CONFVISION_API_URL=https://vision.confmonit2.com.br  (usa API remota se postgres local falhar)
//   VIS_WORKER_API_KEY=...
package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"confvision/src/config"
	"confvision/src/modulos/visdata"
	"confvision/src/storage"
)

const idFranqueado = "0"

var demoEventos = []int{990001, 990002, 990003}

func main() {
	config.ConfigurarApp()

	imgDir := "scripts/moni_demo_imagens/testdata"
	if len(os.Args) > 1 {
		imgDir = os.Args[1]
	}
	if err := os.MkdirAll(imgDir, 0o755); err != nil {
		fail("mkdir testdata: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	baseURL := config.ImagemPublicBaseURL
	if baseURL == "" {
		baseURL = "https://imagem.dnsid.com.br"
	}

	fmt.Println("=== Moni demo imagens ===")
	fmt.Printf("Franqueado: %s | Eventos: %v\n", idFranqueado, demoEventos)
	fmt.Printf("Pasta imagens: %s\n\n", imgDir)

	useAPI := strings.TrimSpace(os.Getenv("CONFVISION_API_URL"))
	if useAPI == "" {
		useAPI = "https://vision.confmonit2.com.br"
	}

	db, dbErr := visdata.DB()
	if dbErr != nil {
		fmt.Printf("Postgres local indisponivel (%v) — usando API %s\n\n", dbErr, useAPI)
		if err := runViaAPI(ctx, useAPI, imgDir, baseURL); err != nil {
			fail("%v", err)
		}
		return
	}

	for i, eventoID := range demoEventos {
		if err := processOne(ctx, db, imgDir, baseURL, i+1, eventoID); err != nil {
			fail("%v", err)
		}
	}

	fmt.Println("Pronto. Teste no navegador ou: curl -o out.jpg \"<Link Moni>\"")
}

func runViaAPI(ctx context.Context, apiBase, imgDir, baseURL string) error {
	key := strings.TrimSpace(os.Getenv("VIS_WORKER_API_KEY"))
	if key == "" {
		key = config.VisWorkerAPIKey
	}
	if key == "" {
		return fmt.Errorf("VIS_WORKER_API_KEY nao configurado")
	}
	client := &http.Client{Timeout: 45 * time.Second}
	apiBase = strings.TrimRight(apiBase, "/")

	for i, eventoID := range demoEventos {
		n := i + 1
		localPath := filepath.Join(imgDir, fmt.Sprintf("%d.jpg", n))
		if err := ensureJPEG(localPath, n); err != nil {
			return fmt.Errorf("imagem %s: %w", localPath, err)
		}
		raw, err := os.ReadFile(localPath)
		if err != nil {
			return fmt.Errorf("ler %s: %w", localPath, err)
		}

		snapURL, err := storage.UploadEventoSnapshot(ctx, idFranqueado, eventoID, raw)
		if err != nil {
			return fmt.Errorf("upload evento=%d: %w", eventoID, err)
		}

		hash := visdata.CodigoImagemPublica(eventoID)
		if hash == "" {
			return fmt.Errorf("hash vazio evento=%d", eventoID)
		}

		if err := upsertDemoEventoAPI(ctx, client, apiBase, key, eventoID, snapURL, hash); err != nil {
			fmt.Fprintf(os.Stderr, "AVISO api evento=%d: %v\n", eventoID, err)
			printSQL(eventoID, snapURL, hash)
		}

		printResult(n, localPath, eventoID, snapURL, hash, baseURL)
	}

	fmt.Println("Pronto. Teste no navegador ou: curl -o out.jpg \"<Link Moni>\"")
	fmt.Println("\nSe os links retornarem 404, faca deploy do binario confvision (build.ps1) e execute este script de novo,")
	fmt.Println("ou rode o SQL impresso acima no Postgres do Core4.")
	return nil
}

func upsertDemoEventoAPI(ctx context.Context, client *http.Client, apiBase, key string, eventoID int, snapURL, hash string) error {
	body, _ := json.Marshal(map[string]any{
		"id":                      eventoID,
		"id_franqueado":           idFranqueado,
		"snapshot_url":            snapURL,
		"codigo_imagem_publico":   hash,
	})
	status, raw, err := apiRequest(ctx, client, http.MethodPost, apiBase+"/vis_evento_demo_moni", key, body)
	if err != nil {
		return err
	}
	if status >= 200 && status < 300 {
		return nil
	}
	return fmt.Errorf("POST /vis_evento_demo_moni HTTP %d: %s", status, truncate(string(raw), 300))
}

func apiRequest(ctx context.Context, client *http.Client, method, url, key string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vis-Worker-Key", key)
	res, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res.StatusCode, raw, nil
}

func processOne(ctx context.Context, db *sql.DB, imgDir, baseURL string, n, eventoID int) error {
	localPath := filepath.Join(imgDir, fmt.Sprintf("%d.jpg", n))
	if err := ensureJPEG(localPath, n); err != nil {
		return fmt.Errorf("imagem %s: %w", localPath, err)
	}

	raw, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("ler %s: %w", localPath, err)
	}

	snapURL, err := storage.UploadEventoSnapshot(ctx, idFranqueado, eventoID, raw)
	if err != nil {
		return fmt.Errorf("upload evento=%d: %w", eventoID, err)
	}

	hash := visdata.CodigoImagemPublica(eventoID)
	if hash == "" {
		return fmt.Errorf("hash vazio evento=%d", eventoID)
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO vis_evento (
	id, id_franqueado, id_cliente, conta, particao, canal,
	tipo_deteccao, snapshot_url,
	codigo_imagem_publico, imagem_liberada_em
) VALUES (
	$1, $2, '0', '0000', '00', '001',
	'movimento', $3,
	$4, NOW()
)
ON CONFLICT (id) DO UPDATE SET
	snapshot_url = EXCLUDED.snapshot_url,
	codigo_imagem_publico = EXCLUDED.codigo_imagem_publico,
	imagem_liberada_em = NOW()`,
		eventoID, idFranqueado, snapURL, hash,
	)
	if err != nil {
		return fmt.Errorf("postgres evento=%d: %w", eventoID, err)
	}

	printResult(n, localPath, eventoID, snapURL, hash, baseURL)
	return nil
}

func printResult(n int, localPath string, eventoID int, snapURL, hash, baseURL string) {
	publicLink := baseURL + "/" + hash
	fmt.Printf("--- Imagem %d ---\n", n)
	fmt.Printf("  Arquivo local : %s\n", localPath)
	fmt.Printf("  Evento ID     : %d\n", eventoID)
	fmt.Printf("  Contabo URL   : %s\n", snapURL)
	fmt.Printf("  Link Moni     : %s\n", publicLink)
	fmt.Printf("  Complemento   : %s\n\n", visdata.ComplementoImagemMoni("046", hash))
}

func ensureJPEG(path string, n int) error {
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		return nil
	}
	colors := []color.RGBA{
		{R: 220, G: 60, B: 60, A: 255},
		{R: 60, G: 160, B: 80, A: 255},
		{R: 50, G: 100, B: 220, A: 255},
	}
	c := colors[(n-1)%len(colors)]
	img := image.NewRGBA(image.Rect(0, 0, 640, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func printSQL(eventoID int, snapURL, hash string) {
	fmt.Printf("  SQL (se API indisponivel):\n")
	fmt.Printf(`  INSERT INTO vis_evento (id, id_franqueado, id_cliente, conta, particao, canal, tipo_deteccao, snapshot_url, codigo_imagem_publico, imagem_liberada_em) VALUES (%d, '0', '0', '0000', '00', '001', 'movimento', '%s', '%s', NOW()) ON CONFLICT (id) DO UPDATE SET snapshot_url = EXCLUDED.snapshot_url, codigo_imagem_publico = EXCLUDED.codigo_imagem_publico, imagem_liberada_em = NOW();`+"\n\n", eventoID, snapURL, hash)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERRO: "+format+"\n", args...)
	os.Exit(1)
}
