package dispositivoV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func insere(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Envia as credenciais para o app solicitante
	respApp.Dados(w, http.StatusOK, obj.ID_Dispositivo)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj)
}

func getDadosByIdFisico1(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GetDadosByIdFisico1(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj)
}

func getDadosByIdFisico2(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GetDadosByIdFisico2(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj)
}

func getMsgAtendenteById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GetMsgAtendenteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj.MsgAtendente)
}

func setMsgAtendenteById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.SetMsgAtendenteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func getArmadoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GetArmadoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj.Armado)
}

// workerGetArmadoById — leitura de Armado (+ fabricante) para o worker ConfVision.
// Não altera getArmadoById (usado por outros sistemas com JWT).
func workerGetArmadoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var obj Dispositivo
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := obj.GetArmadoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	idFabricante := ""
	if err := obj.GetDadosById(); err == nil {
		idFabricante = obj.ID_Fabricante
	}

	respApp.Dados(w, http.StatusOK, map[string]string{
		"armado":       obj.Armado,
		"idFabricante": idFabricante,
	})
}

func setArmadoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.SetArmadoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func inverteArmadoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.InverteArmadoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func getAtivoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func setAtivoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.SetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func inverteAtivoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.InverteAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}
	respApp.OK(w)
}

func gravarUltimoEventoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GravarUltimoEventoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func deletaById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.DeletaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func getNomeClienteById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var nome string
	// Executa o comando
	if err := obj.GetNomeClienteById(&nome); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, nome)
}

func getIdClienteById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var idCliente string
	// Executa o comando
	if err := obj.GetIdClienteById(&idCliente); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, idCliente)
}

func getIdFranqueadoById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var idFranqueado string
	// Executa o comando
	if err := obj.GetIdFranqueadoById(&idFranqueado); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, idFranqueado)
}

func getIdRepresentanteById(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var idRepresentante string
	// Executa o comando
	if err := obj.GetIdRepresentanteById(&idRepresentante); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, idRepresentante)
}

func listarSemComunicacaoByIdFranqueado(w http.ResponseWriter, r *http.Request) {

	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Dispositivo
	// Executa o comando
	if err := obj.ListarSemComunicacaoByIdFranqueado(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func listarByIdCliente(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Dispositivo
	// Executa o comando
	if err := obj.ListarByIdCliente(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func listarByIdFranqueado(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Dispositivo
	// Executa o comando
	if err := obj.ListarByIdFranqueado(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func gerarContaByIdFranqueado(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.GerarContaByIdFranqueado(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj.Conta)
}

func vericaContaByIdFranquado(w http.ResponseWriter, r *http.Request) {
	// Reucpera o email e senha enviado no corpo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Cria obj para receber os dados
	var obj Dispositivo

	// Carrega os dados recebido no objeto
	if err := json.Unmarshal(body, &obj); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Executa o comando
	if err := obj.VericaContaByIdFranquado(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, obj.NomeCliente)
}
