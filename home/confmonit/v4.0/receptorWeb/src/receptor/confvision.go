package receptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"receptorWeb/src/auxiliar"
	"receptorWeb/src/config"
	evento "receptorWeb/src/eventos"
	"receptorWeb/src/respApp"
	"strings"
)

type ReceptorConfVision struct {
	IDFranqueado  string `json:"idFranqueado"`
	Conta         string `json:"conta"`
	Particao      string `json:"particao"`
	ZonaUser      string `json:"zonauser"`
	VisEventoID   int64  `json:"vis_evento_id"`
	SenhaRecSerial string `json:"senha"`
}

// ReceberEventoConfVision recebe deteccao analitica do worker e abre processo no terminal.
func ReceberEventoConfVision(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		fmt.Println("Informando Erro Receptor Web ConfVision =>", erro)
		respApp.RespErro(w, http.StatusBadRequest, erro)
		return
	}

	var d ReceptorConfVision
	if erro := json.Unmarshal(corpo, &d); erro != nil {
		fmt.Println("Informando Erro Receptor Web ConfVision =>", erro)
		respApp.RespErro(w, http.StatusBadRequest, erro)
		return
	}

	if d.SenhaRecSerial != config.SenhaWeb {
		fmt.Println("Informando Erro Receptor Web ConfVision => senha invalida")
		respApp.RespErro(w, http.StatusUnauthorized, errors.New("senha invalida"))
		return
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		fmt.Println("Informando Erro Receptor Web ConfVision =>", err)
		respApp.RespErro(w, http.StatusBadRequest, err)
		return
	}
	defer db.Close()

	res, erro := evento.RecebeEventoConfVision(
		db,
		strings.TrimSpace(d.IDFranqueado),
		d.Conta,
		d.Particao,
		d.ZonaUser,
		d.VisEventoID,
	)
	if erro != nil {
		fmt.Println("Informando Erro Receptor Web ConfVision =>", erro.Error())
		respApp.RespErro(w, http.StatusBadRequest, erro)
		return
	}

	respApp.RespDados(w, http.StatusOK, res)
}
