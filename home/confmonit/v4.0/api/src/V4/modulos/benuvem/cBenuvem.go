package benuvemV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

/*
{
	"codigoCliente": "",
	"particao": "",
	"codigoFranqueado": "",
	"canais": [""]
}
*/

// OK gerarEvento gera um evento para inicio de gravação
func gerarEvento(w http.ResponseWriter, r *http.Request) {

	// Pega o conteudo vindo no corpo da requisicao
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}
	var be BeNuvem
	// Popula o objeto evt com valores vindo do corpo requisiçõa
	var evt BenuvemEvento
	if erro := json.Unmarshal(corpo, &evt); erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	if err := be.GerarEvento(evt); err != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	respApp.OK(w)

}

// OK finalizarEvento finaliza um evento gerado anteriormente
func finalizarEvento(w http.ResponseWriter, r *http.Request) {

	// Pega o conteudo vindo no corpo da requisicao
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	var be BeNuvem
	// Popula o objeto evt com valores vindo do corpo requisiçõa
	var e BenuvemEvento
	if erro := json.Unmarshal(corpo, &e); erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	if err := be.FinalizarEvento(e); err != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	respApp.OK(w)

}

func getRtpsImagens(w http.ResponseWriter, r *http.Request) {
	respApp.OK(w)

}

// pega a url da camera
func getUrlImagem(w http.ResponseWriter, r *http.Request) {
	//codCliente, particao, data, popup string, channels []string
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	var e BenuvemEvento
	if erro := json.Unmarshal(corpo, &e); erro != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	var resp struct {
		Url    string `json:"url"`
		Expira int    `json:"expira"`
	}
	var be BeNuvem
	if err := be.GetUrlCamera(e, &resp.Url, &resp.Expira); err != nil {
		respApp.Erro(w, http.StatusBadRequest, erro)
		return
	}

	respApp.Dados(w, http.StatusOK, resp)
}
