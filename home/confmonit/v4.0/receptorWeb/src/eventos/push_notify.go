package evento

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func notifyTerminalPush(objEvt *Evento) {
	url := strings.TrimSpace(os.Getenv("TERMINAL_PUSH_URL"))
	if url == "" {
		url = "https://terminalmovel.confmonit2.com.br/interno/push/processo"
	}
	secret := strings.TrimSpace(os.Getenv("TERMINAL_PUSH_SECRET"))

	payload := map[string]string{
		"idProcesso":     objEvt.idProcesso,
		"idFranqueado":   objEvt.idFranqueado,
		"nomeCliente":    objEvt.nomeCliente,
		"nomeFranqueado": objEvt.nomeFranqueado,
		"nivel":          objEvt.nivel,
		"codigo":         objEvt.codigo,
		"descricao":      objEvt.descricao,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("[push] marshal:", err)
		return
	}

	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Println("[push] request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("X-Terminal-Push-Secret", secret)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println("[push] envio:", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		log.Println("[push] status:", res.StatusCode)
	}
}
