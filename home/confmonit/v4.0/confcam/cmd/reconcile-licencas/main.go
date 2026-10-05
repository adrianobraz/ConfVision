// Ferramenta de validacao: reconcilia licencas pendentes vs faturas pagas (ConfCam + Xano leitura).
//
// Uso:
//
//	go run ./cmd/reconcile-licencas -id_franqueado=2026082511344085584850570
//	go run ./cmd/reconcile-licencas -todos
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"confvision/src/config"
	"confvision/src/modulos/visdata"
	"confvision/src/xanopro"

	"github.com/joho/godotenv"
)

func dumpFaturaItens(idFranqueado string) {
	raw, err := xanopro.Post("/fp_fatura_minhas", map[string]any{
		"id_franqueado": idFranqueado,
		"status":        "paga",
		"incluir_itens": true,
	})
	if err != nil {
		log.Fatal(err)
	}
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(b))
}

func main() {
	idFra := flag.String("id_franqueado", "", "Franqueado alvo")
	todos := flag.Bool("todos", false, "Reconciliar todos com licenca pendente")
	verbose := flag.Bool("verbose", false, "Mostrar erros de ativacao")
	dumpItens := flag.Bool("dump-itens", false, "Listar itens vis_licenca das faturas (debug)")
	flag.Parse()

	_ = godotenv.Load()
	config.ConfigurarApp()

	if !config.VisPostgresEnabled {
		log.Fatal("VIS_POSTGRES_ENABLED necessario")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if *todos {
		porFra, err := visdata.ReconciliarLicencasPagasTodosPendentes(ctx)
		if err != nil {
			log.Fatalf("reconciliar todos: %v", err)
		}
		total := 0
		for id, n := range porFra {
			if n < 0 {
				fmt.Printf("%s: ERRO\n", id)
				continue
			}
			fmt.Printf("%s: ativadas=%d\n", id, n)
			if n > 0 {
				total += n
			}
		}
		fmt.Printf("total ativadas: %d\n", total)
		os.Exit(0)
	}

	id := strings.TrimSpace(*idFra)
	if id == "" {
		log.Fatal("informe -id_franqueado ou -todos")
	}

	if *dumpItens {
		dumpFaturaItens(id)
		return
	}

	var n int
	var err error
	var verr []string
	if *verbose {
		n, verr, err = visdata.ReconciliarLicencasPagasFranqueadoVerbose(ctx, id)
	} else {
		n, err = visdata.ReconciliarLicencasPagasFranqueado(ctx, id)
	}
	if err != nil {
		log.Fatalf("reconciliar: %v", err)
	}
	fmt.Printf("id_franqueado=%s ativadas=%d\n", id, n)
	for _, e := range verr {
		fmt.Println("erro:", e)
	}
	if n == 0 {
		diag, derr := visdata.DiagnoseReconciliarLicencas(id)
		if derr != nil {
			log.Printf("diagnostico: %v", derr)
			return
		}
		fmt.Printf("diagnostico: %+v\n", diag)
		if diag["incluir_itens_necessario"] == true {
			fmt.Println("hint: publicar 2286_fp_fatura_minhas (incluir_itens) no Xano e repetir")
		}
	}
}
