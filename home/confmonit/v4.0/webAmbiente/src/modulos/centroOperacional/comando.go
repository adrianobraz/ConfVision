package centroOperacional

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webAmbiente/src/config"
)

type comandoReq struct {
	IdDispositivo string `json:"idDispositivo"`
	Numero        int    `json:"numero"`
	Usuario       string `json:"usuario"`
	Acao          int    `json:"acao"`
	Senha         string `json:"senha"`
	SenhaWeb      string `json:"senhaWeb"`
}

func enviarComando(idDispositivo, senha, particao string, armar bool) error {
	if config.UrlComando == "" {
		return fmt.Errorf("URL_COMANDO nao configurada")
	}

	num, _ := strconv.Atoi(strings.TrimSpace(particao))
	if num <= 0 {
		num = 1
	}

	acao := 0
	if armar {
		acao = 1
	}

	payload := comandoReq{
		IdDispositivo: strings.TrimSpace(idDispositivo),
		Numero:        num,
		Usuario:       "000",
		Acao:          acao,
		Senha:         strings.TrimSpace(senha),
		SenhaWeb:      config.SenhaWebComando,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(config.UrlComando, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("comando remoto: %s", string(raw))
	}
	return nil
}
