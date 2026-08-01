package erroConexao

import (
	"api/src/V4/respApp"
	"net/http"
)

func listar(w http.ResponseWriter, r *http.Request) {

	var ec ErroConexao

	var lista []ErroConexao

	if err := ec.listar(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}

}

func limpar(w http.ResponseWriter, r *http.Request) {

	var ec ErroConexao

	if err := ec.limpar(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia resposta para o APP
	respApp.OK(w)
}
