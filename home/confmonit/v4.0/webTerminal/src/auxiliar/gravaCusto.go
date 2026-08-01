package aux

import (
	"encoding/json"
	"io"
	"net/http"
	"terminal/src/tipos"
)

var RotasGravaCusto = []tipos.Rota{
	{
		Uri:    "/auxiliar/gravaCusto",
		Metodo: http.MethodPost,
		Funcao: gravaCusto,
	},
}

func gravaCusto(w http.ResponseWriter, r *http.Request) {

	type objeto struct {
		ID_Tarifacao string `json:"idTarifacao"`
		ID_Vinculo   string `json:"idVinculo"`

		TipoOperacao string `json:"tipoOperacao"`
		DadoOperacao string `json:"dadoOperacao"`
		Aux          string `json:"aux"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// tipo ATENDIMENTO, LIGACAO, EMAIL, SMS e Atendimento
	if obj.ID_Tarifacao == "" {
		obj.ID_Tarifacao = GeradorDeId()
	}

	// Abre um canal de conexao
	db, erro := Conectar()
	if erro != nil {
		RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	stm, erro := db.Prepare(`
		INSERT INTO tarifacao(
			tarifacao.ID_Tarifacao, 
			tarifacao.ID_Vinculo,
			tarifacao.TipoOperacao, 
			tarifacao.DadoOperacao,
			tarifacao.Aux
		) VALUES (?,?,?,?,?)
	`)
	if erro != nil {
		RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		obj.ID_Tarifacao,
		obj.ID_Vinculo,
		obj.TipoOperacao,
		obj.DadoOperacao,
		obj.Aux,
	); erro != nil {
		RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	RespostaJsonOK(w)
}
