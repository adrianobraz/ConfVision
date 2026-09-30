// Reparo: marca ativação fatura 6 e gera repasses faturas 6/8 no piloto.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"apifunction/config"
	"apifunction/confservice"
	"apifunction/pgcredito"
	"apifunction/pgatendimento"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if os.Getenv("CONFSERVICE_URL") == "" {
		_ = os.Setenv("CONFSERVICE_URL", "http://185.130.61.4:2020")
	}
	config.Carregar()
	ctx := context.Background()
	dsn := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if dsn == "" {
		dsn = "postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
	}
	if err := pgcredito.Abrir(dsn); err != nil {
		fmt.Fprintln(os.Stderr, "postgres:", err)
		os.Exit(1)
	}
	idFra := "2026072204185539042285696"
	if len(os.Args) > 1 {
		idFra = os.Args[1]
	}

	// Fatura 6 — ativação parceiro padrão (estava cancelada apesar de paga no Xano)
	pagoAte := time.Now().AddDate(0, 0, 14)
	rec6, err := pgatendimento.MarcarAtivacaoPagaFatura(ctx, idFra, 6, "", pagoAte)
	if err != nil {
		fmt.Println("fatura 6 marcar:", err)
	} else if rec6 != nil {
		fmt.Printf("fatura 6 ativacao id=%d parceiro=%s status=%s\n", rec6.ID, rec6.IDParceiro, rec6.Status)
		if idP, piso, err := pgatendimento.MetaRepasseAtivacao(ctx, idFra, 6, rec6.IDParceiro, 0); err == nil {
			err = confservice.GerarRepassesParceiro(idFra, 6, []confservice.RepasseItem{{
				IDParceiro: idP, ValorPiso: piso, FaturaID: 6,
				RefTipo: pgatendimento.RefTipoParceiroConfig, RefID: idP,
			}})
			fmt.Printf("fatura 6 repasse parceiro=%s piso=%.2f err=%v\n", idP, piso, err)
		} else {
			fmt.Println("fatura 6 meta repasse:", err)
		}
	}

	// Fatura 8 — exceção parceiro e71ee186
	if idV, idP, piso, err := pgatendimento.MetaRepasseExcecao(ctx, "1", 0); err == nil {
		fmt.Printf("fatura 8 excecao vinculo=%s parceiro=%s piso=%.2f\n", idV, idP, piso)
		err = confservice.GerarRepassesParceiro(idFra, 8, []confservice.RepasseItem{{
			IDVinculo: idV, IDParceiro: idP, ValorPiso: piso, FaturaID: 8,
			RefTipo: "cs_parceiro_vinculo", RefID: idV,
		}})
		fmt.Println("fatura 8 repasse err=", err)
	} else {
		fmt.Println("fatura 8 meta:", err)
	}

	list, _ := pgatendimento.ListFaturasParceiroStatus(ctx, idFra)
	fmt.Println("\n=== status faturas ===")
	for _, it := range list {
		if it.FPFaturaID != 6 && it.FPFaturaID != 8 {
			continue
		}
		fmt.Printf("fp=%d tipo=%s parceiro=%s ops=%s fase=%s repasse=%s\n",
			it.FPFaturaID, it.Tipo, short(it.IDParceiro), it.StatusOps, it.Fase, it.RepasseStatus)
	}
}

func short(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 8 {
		return s
	}
	return s[:8] + "..."
}
