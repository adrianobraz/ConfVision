package usuariosV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func insere(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.ID_Usuario)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use)
}

func getDadosByEmail1(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetDadosByEmail1(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use)
}

func getIdVinculoByEmail1(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetIdVinculoByEmail1(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.ID_Vinculo)
}

func getEmail1Livre(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	var livre string
	// Executa o comendo
	if err := use.GetEmail1Livre(&livre); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, livre)
}

func getEmail2Livre(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var livre string
	// Executa o comendo
	if err := use.GetEmail2Livre(&livre); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, livre)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func deletaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func deletaAllByVinculo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.DeletaAllByVinculo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func listar(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria uma lista para receber os dados
	var lista []Usuario

	// Executa o comendo
	if err := use.Listar(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se a lista esta vazia
	if len(lista) <= 0 {
		// Envia resposta para o APP
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria uma lista para receber os dados
	var lista []Usuario

	// Executa o comendo
	if err := use.ListarByVinculo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se a lista esta vazia
	if len(lista) <= 0 {
		// Envia resposta para o APP
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculoCentral(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var use Usuario
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Usuario
	if err := use.ListarByVinculoCentral(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculoCentralPorReferencia(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var use Usuario
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Usuario
	if err := use.ListarByVinculoCentralPorReferencia(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) <= 0 {
		respApp.Vazio(w)
	} else {
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculoToMaster(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria uma lista para receber os dados
	var lista []Usuario

	// Executa o comendo
	if err := use.ListarByVinculoToMaster(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se a lista esta vazia
	if len(lista) <= 0 {
		// Envia resposta para o APP
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculoToNotMaster(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria uma lista para receber os dados
	var lista []Usuario

	// Executa o comendo
	if err := use.ListarByVinculoToNotMaster(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se a lista esta vazia
	if len(lista) <= 0 {
		// Envia resposta para o APP
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculoToAtivo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria uma lista para receber os dados
	var lista []Usuario

	// Executa o comendo
	if err := use.ListarByVinculoToAtivo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se a lista esta vazia
	if len(lista) <= 0 {
		// Envia resposta para o APP
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func listarByVinculoToNotAtivo(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria uma lista para receber os dados
	var lista []Usuario

	// Executa o comendo
	if err := use.ListarByVinculoToNotAtivo(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Verifica se a lista esta vazia
	if len(lista) <= 0 {
		// Envia resposta para o APP
		respApp.Vazio(w)
	} else {
		// Envia resposta para o APP
		respApp.Dados(w, http.StatusOK, lista)
	}
}

func getUsuarioAtivaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetUsuarioAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.Ativo)
}

func setUsuarioAtivaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.SetUsuarioAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.Ativo)
}

func inverteUsuarioAtivaById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("aqui")
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.InverteUsuarioAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.Ativo)
}

func getTerminalAtivaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetTerminalAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.UsuarioTeminal)
}

func setTerminalAtivaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.SetTerminalAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.UsuarioTeminal)
}

func inverteTerminalAtivaById(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.InverteTerminalAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.UsuarioTeminal)
}

func getWebAtivaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetWebAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.UsuarioWeb)
}

func setWebAtivaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.SetWebAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.UsuarioWeb)
}

func inverteWebAtivaById(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.InverteWebAtivaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.UsuarioWeb)
}

// func getTipoById(w http.ResponseWriter, r *http.Request) {
// 	// Recupera o conteudo da requisição
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Cria um objeto para receber o conteudo
// 	var use Usuario

// 	// Carrega o conteudo no objeto
// 	if err := json.Unmarshal(body, &use); err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Executa o comendo
// 	if err := use.GetTipoById(); err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Responde para o APP
// 	respApp.Dados(w, http.StatusOK, use.Tipo)
// }

// func setTipoById(w http.ResponseWriter, r *http.Request) {
// 	// Recupera o conteudo da requisição
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Cria um objeto para receber o conteudo
// 	var use Usuario

// 	// Carrega o conteudo no objeto
// 	if err := json.Unmarshal(body, &use); err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Executa o comendo
// 	if err := use.SetTipoById(); err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Responde para o APP
// 	respApp.Dados(w, http.StatusOK, use.Tipo)
// }

// func masterAtiva(w http.ResponseWriter, r *http.Request) {
// 	// Recupera o conteudo da requisição
// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Cria um objeto para receber o conteudo
// 	var use Usuario

// 	// Carrega o conteudo no objeto
// 	if err := json.Unmarshal(body, &use); err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Executa o comendo
// 	if err := use.MasterAtivaById(); err != nil {
// 		respApp.Erro(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	// Responde para o APP
// 	respApp.Dados(w, http.StatusOK, use.MasterAtivaById())
// }

func validaEmail1(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.Email1)
}

func getAtivarEnviarEmailById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.GetAtivarEnviarEmailById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.EnviarEmail)
}

func setAtivarEnviarEmailById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.SetAtivarEnviarEmailById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.EnviarEmail)
}

func inverteAtivarEnviarEmailById(w http.ResponseWriter, r *http.Request) {

	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.InverteAtivarEnviarEmailById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.Dados(w, http.StatusOK, use.EnviarEmail)
}

func resetarSenhaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.ResetarSenhaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}

func alterarSenhaById(w http.ResponseWriter, r *http.Request) {
	// Recupera o conteudo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria um objeto para receber o conteudo
	var use Usuario

	// Carrega o conteudo no objeto
	if err := json.Unmarshal(body, &use); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comendo
	if err := use.AlterarSenhaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Responde para o APP
	respApp.OK(w)
}
