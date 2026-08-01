package receptorV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func recebeEvento(w http.ResponseWriter, r *http.Request) {
	// Recuperar o conteudo do corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Criar obj para receber o conteudo
	var rec Receptor
	// Carregar o conteudo no obj
	if err := json.Unmarshal(body, &rec); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executar a operação
	if err := rec.RecebeEvento(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}
