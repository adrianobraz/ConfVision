// Reparo: estorna repasses REP pagos (7, 9) e pos-estorno parceiro (6, 8) no piloto.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"apifunction/config"
	"apifunction/pgcredito"
	"apifunction/pgatendimento"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if os.Getenv("CONFSERVICE_URL") == "" {
		_ = os.Setenv("CONFSERVICE_URL", "http://185.130.61.4:2020")
	}
	config.Carregar()
	dsn := strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	if dsn == "" {
		dsn = "postgres://confmonit:20080328DriziN@191.96.156.116:5432/confmonit?sslmode=disable"
	}
	if err := pgcredito.Abrir(dsn); err != nil {
		fmt.Fprintln(os.Stderr, "postgres:", err)
		os.Exit(1)
	}

	idFra := "2026072204185539042285696"
	xano := strings.TrimRight(strings.TrimSpace(os.Getenv("XANO_API_FINANCEIRO")), "/")
	wk := strings.TrimSpace(os.Getenv("WORKER_SECRET"))

	for _, fp := range []int{7, 9} {
		if xano == "" || wk == "" {
			fmt.Printf("fatura REP %d: configure XANO_API_FINANCEIRO e WORKER_SECRET\n", fp)
			continue
		}
		body, _ := json.Marshal(map[string]any{
			"worker_key": wk, "fatura_id": fp,
			"motivo": "Reparo estorno cascata piloto",
		})
		req, _ := http.NewRequest(http.MethodPost, xano+"/fp_repasse_estornar_pagamento", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("fatura REP %d err: %v\n", fp, err)
			continue
		}
		raw, _ := readBody(res)
		fmt.Printf("fatura REP %d HTTP %d: %s\n", fp, res.StatusCode, strings.TrimSpace(string(raw)))
		res.Body.Close()
	}

	ctx := context.Background()
	for _, fp := range []int{6, 8} {
		r, err := pgatendimento.ReverterPosEstorno(ctx, idFra, fp)
		fmt.Printf("pos-estorno fatura %d: %+v err=%v\n", fp, r, err)
	}
}

func readBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(res.Body)
	return buf.Bytes(), err
}
