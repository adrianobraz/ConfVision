package auxiliar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"franqueadopro/src/config"
	"franqueadopro/src/seguranca"
)

func MasterDoCookie(r *http.Request) string {
	cookie, err := seguranca.LerCookies(r)
	if err != nil {
		return ""
	}
	if cookie["master"] != "" {
		return cookie["master"]
	}
	return buscarMasterNaAPI(r, cookie["idUsuario"])
}

func PreencherEhMaster(r *http.Request, d *Pagina) {
	if d == nil {
		return
	}
	d.EhMaster = MasterDoCookie(r) == "S"
}

func buscarMasterNaAPI(r *http.Request, idUsuario string) string {
	if idUsuario == "" {
		return ""
	}

	body, _ := json.Marshal(map[string]string{"idUsuario": idUsuario})
	url := fmt.Sprintf("%s/v4/usuario/getDadosById", config.ApiUrl)
	response, err := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode >= 400 {
		return ""
	}

	var resp struct {
		Dados struct {
			Master string `json:"master"`
		} `json:"dados"`
	}
	if json.Unmarshal(raw, &resp) != nil {
		return ""
	}
	return resp.Dados.Master
}
