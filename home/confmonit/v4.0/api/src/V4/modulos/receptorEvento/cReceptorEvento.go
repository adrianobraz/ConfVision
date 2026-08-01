package receptorEvento

import (
	"api/src/V4/respApp"
	"net/http"
)

func carregarDados(w http.ResponseWriter, r *http.Request) {

	// Criar obj para receber o conteudo
	var rec ReceptorEvento

	var lista []DadosCarregamento

	// Executar a operação
	if err := rec.carregarDados(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) < 1 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}

}
