package licenca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
)

// BloquearInsercaoSeLimite retorna true se a insercao deve ser bloqueada (resposta ja enviada).
func BloquearInsercaoSeLimite(w http.ResponseWriter, r *http.Request, recurso string, quantidadeAtual int) bool {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return false
	}
	est, _ := Verificar(cookie["idFranqueado"], "franqueadopro")
	ok, msg := PodeInserir(est, recurso, quantidadeAtual)
	if !ok {
		RespostaLimiteExcedido(w, msg)
		return true
	}
	return false
}

func ContarClientesFranqueado(r *http.Request, idFranqueado string) (int, error) {
	payload, _ := json.Marshal(map[string]string{"idFranqueado": idFranqueado})
	url := fmt.Sprintf("%s/v4/cliente/listarByIdFranqueado", config.ApiUrl)
	resp, err := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return ContarItensJSON(corpo), nil
}

func ContarDispositivosFranqueado(r *http.Request, idFranqueado string) (int, error) {
	url := fmt.Sprintf("%s/dispositivo-listar/%s", config.ApiUrl, idFranqueado)
	resp, err := seguranca.RequisiacaoAutenticada(r, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return ContarItensJSON(corpo), nil
}
