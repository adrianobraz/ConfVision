package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	base := "https://xpcy-oyme-lno7.b2.xano.io/api:-WvTZ3QM"

	loginBody := map[string]string{"usuario": "usuario@teste.rep", "senha": "usuario123"}
	loginJSON, _ := json.Marshal(loginBody)
	resp, err := http.Post(base+"/fp_admin_login", "application/json", bytes.NewReader(loginJSON))
	if err != nil {
		panic(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var login map[string]any
	_ = json.Unmarshal(body, &login)
	token, _ := login["admin_token"].(string)
	rep, _ := login["idVinculo"].(string)
	cenCatalogo, _ := login["idCentralCatalogo"].(string)
	fmt.Printf("LOGIN idVinculo=%s idCentralCatalogo=%s\n", rep, cenCatalogo)

	dashBody := map[string]string{"admin_token": token, "competencia": "2026-08"}
	dashJSON, _ := json.Marshal(dashBody)
	resp2, _ := http.Post(base+"/fp_fin_dashboard", "application/json", bytes.NewReader(dashJSON))
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()

	var dash map[string]any
	_ = json.Unmarshal(body2, &dash)
	gov, _ := json.Marshal(dash["governanca_alerta"])
	fmt.Println("DASHBOARD governanca_alerta:", string(gov))

	urls := []string{
		fmt.Sprintf("http://185.130.61.4:20001/financeiro/governanca/estado?user_tipo=REP&id_vinculo=%s&id_central=%s&breakglass=false", rep, cenCatalogo),
		fmt.Sprintf("http://185.130.61.4:20001/financeiro/governanca/estado?user_tipo=REP&id_vinculo=%s&id_central=CENTRAL&breakglass=false", rep),
	}
	for _, u := range urls {
		r, err := http.Get(u)
		if err != nil {
			fmt.Println("GOV ERR", err)
			continue
		}
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println("GOV", u, string(b))
	}

	payJSON, _ := json.Marshal(map[string]any{
		"admin_token": token, "admin_usuario": "usuario@teste.rep",
		"fatura_id": 121, "metodo": "manual", "observacao": "diag go",
	})
	resp3, _ := http.Post(base+"/fp_fatura_registrar_pagamento", "application/json", bytes.NewReader(payJSON))
	body3, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	fmt.Println("PAY:", string(body3))
}
