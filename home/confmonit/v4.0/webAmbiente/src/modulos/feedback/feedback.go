package feedback

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"webAmbiente/src/auxiliar"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
)

var Rotas = []tipos.Rota{
	{Uri: "/feedback/enviar", Metodo: http.MethodPost, Controle: enviar, Seguro: true},
}

func enviar(w http.ResponseWriter, r *http.Request) {
	op, err := seguranca.LerOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	idUsuario := strings.TrimSpace(op["idOperador"])
	if idUsuario == "" {
		resposta.Erro(w, http.StatusUnauthorized, errors.New("sessao invalida"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		Tipo      string `json:"tipo"`
		Descricao string `json:"descricao"`
		URLPagina string `json:"urlPagina"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	tipo := strings.ToLower(strings.TrimSpace(req.Tipo))
	descricao := strings.TrimSpace(req.Descricao)
	urlPagina := strings.TrimSpace(req.URLPagina)

	if tipo != "bug" && tipo != "melhoria" && tipo != "ideia" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("tipo invalido"))
		return
	}
	if len(descricao) < 5 {
		resposta.Erro(w, http.StatusBadRequest, errors.New("descricao obrigatorio"))
		return
	}

	idFranqueado := ""
	if strings.ToUpper(strings.TrimSpace(op["userTipo"])) == "FRA" {
		idFranqueado = strings.TrimSpace(op["idVinculo"])
	}

	idFeedback := fmt.Sprintf("%020d", time.Now().UnixNano()%1e18)
	if len(idFeedback) > 20 {
		idFeedback = idFeedback[len(idFeedback)-20:]
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO cm_feedback_cliente (
			ID_Feedback, Software, Tipo, Descricao, URL_Pagina,
			ID_Usuario, ID_Franqueado
		) VALUES (?, 'webambiente', ?, ?, NULLIF(?, ''), ?, NULLIF(?, ''))
	`, idFeedback, tipo, descricao, urlPagina, idUsuario, idFranqueado)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "OK",
		"dados":  map[string]string{"idFeedback": idFeedback},
	})
}
