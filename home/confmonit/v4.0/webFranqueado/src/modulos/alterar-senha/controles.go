package alterarSenha

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"webFranqueado/src/auxiliar"
	"webFranqueado/src/config"
	"webFranqueado/src/seguranca"

	"github.com/joho/godotenv"
)

// CarregarPaginaMenuPrincipal carrega a pagina do menu principal
func CarregarAlterarSenha(w http.ResponseWriter, r *http.Request) {
	res, erro := godotenv.Read()
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if res["MANUTENCAO"] == "S" {
		var d auxiliar.Pagina

		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "emManutencao.html", d)
	} else {
		var d auxiliar.Pagina

		// Carrega title do site
		d.TituloSite = config.TituloSite

		auxiliar.ExecutarTemplate(w, "alterar-senha.html", d)
	}
}

func AlterarSenha(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Monta a url para requisição na API
	url := fmt.Sprintf(`%s/v4/usuario/alterarSenhaById`, config.ApiUrl)

	// Faz a requisição na API
	response, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	// Verifica se o status code da resposta esta na faixa dos erros
	if response.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, response)
		return
	}

	// Pega o json no copro da resposta
	body, erro = io.ReadAll(response.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	auxiliar.RespostaAPP(w, body)

}
