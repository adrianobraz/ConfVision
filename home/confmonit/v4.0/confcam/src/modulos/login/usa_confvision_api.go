package login

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"confvision/src/auxiliar"
)

func apiUsaConfVision(w http.ResponseWriter, r *http.Request) {
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var entrada struct {
		FraId string `json:"fraId"`
	}

	if erro = json.Unmarshal(body, &entrada); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	usa, erro := getUsaConfVision(entrada.FraId)
	if erro != nil {
		if strings.Contains(erro.Error(), "nao encontrado") {
			auxiliar.RespostaErro(w, http.StatusNotFound, erro)
			return
		}
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaJSON(w, http.StatusOK, struct {
		Status string `json:"status"`
		Dados  struct {
			FraId           string `json:"fraId"`
			FraUsaConfVision string `json:"fraUsaConfVision"`
		} `json:"dados"`
	}{
		Status: "OK",
		Dados: struct {
			FraId           string `json:"fraId"`
			FraUsaConfVision string `json:"fraUsaConfVision"`
		}{
			FraId:           strings.TrimSpace(entrada.FraId),
			FraUsaConfVision: usa,
		},
	})
}
