// test-create-camera — exercita visdata.CreateCamera (mesmo fluxo do POST /api/cameras).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"confvision/src/config"
	"confvision/src/conexao"
	"confvision/src/modulos/visdata"

	"github.com/joho/godotenv"
)

func main() {
	licID := flag.Int("vis_licenca_id", 390, "licenca disponivel")
	idFra := flag.String("id_franqueado", "2026082511344085584850570", "franqueado")
	nome := flag.String("nome", "", "nome camera")
	debug := flag.Bool("debug", false, "testar conexao Go antes")
	flag.Parse()

	_ = godotenv.Load()
	config.ConfigurarApp()
	if !config.VisPostgresEnabled {
		log.Fatal("VIS_POSTGRES_ENABLED necessario")
	}

	if *debug {
		db, err := conexao.ConectarPostgres()
		if err != nil {
			log.Fatalf("ConectarPostgres: %v", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatalf("Ping: %v", err)
		}
		fmt.Println("ConectarPostgres + Ping OK")
		_ = db.Close()
	}

	n := *nome
	if n == "" {
		n = fmt.Sprintf("teste cadastro agent %s", time.Now().Format("20060102-150405"))
	}

	input := map[string]any{
		"ativo":             false,
		"nome":              n,
		"id_franqueado":     *idFra,
		"id_cliente":        "2026100511085794635792043",
		"id_dispositivo":    "2026100511092418859823640",
		"conta":             "0001",
		"particao":          "01",
		"canal":             "",
		"setor":             "teste",
		"protocolo":         "wifi",
		"status":            "offline",
		"vis_licenca_id":    *licID,
		"zonauser":          "002",
		"captura_analitico": true,
		"captura_sensor":    false,
		"confianca_min":     0.5,
		"cooldown_seg":      30,
		"deteccao_humano":   true,
		"deteccao_veiculo":  false,
		"modo_deteccao":     "dentro",
		"somente_armado":    false,
		"id_setor":          "2026100502365068090338143",
	}

	ctx := context.Background()
	out, err := visdata.CreateCamera(ctx, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CreateCamera ERRO (apos humanize): %v\n", err)
		os.Exit(1)
	}
	raw, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println("CreateCamera OK:")
	fmt.Println(string(raw))
}
