package clienteV4

import (
	"api/src/V4/respApp"
	"encoding/json"
	"io"
	"net/http"
)

func logar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.Logar(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli)
}

func webLogar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.WebLogar(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.Token)
}

func insere(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.Insere(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.ID_Cliente)
}

func getDadosById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.GetDadosById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli)
}

func alteraById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.AlteraById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func deleteById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.DeleteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
func preDeleteById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.PreDeleteById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func deleteAllByVinculo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.DeleteAllByVinculo(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func listarByIdFranqueado(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []Cliente
	if err := cli.ListarByIdFranqueado(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func listarComDispByIdFranqueado(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []clienteDisp
	if err := cli.ListarComDispByIdFranqueado(&lista); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(lista) > 0 {
		respApp.Dados(w, http.StatusOK, lista)
	} else {
		respApp.Vazio(w)
	}
}

func getAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.GetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.Ativo)
}

func setAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.SetAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func inverterAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.InverterAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.Ativo)
}

func getEmailAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.GetEmailAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.EnvioEmail)
}

func setEmailAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.SetEmailAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func inverterEmailAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.InverterEmailAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.EnvioEmail)
}

func getSmsAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.GetSmsAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.EnvioSms)
}

func setSmsAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.SetSmsAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func inverterSmsAtivoById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.InverterSmsAtivoById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.Ativo)
}

func resetarSenha(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.ResetarSenhaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func alterarSenhaById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.AlterarSenhaById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func getEmail1LivreByEmail1(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.GetEmail1LivreByEmail1(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.Dados(w, http.StatusOK, cli.Nome)
}

func setEmail1ById(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.SetEmailById(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}

func setDispPadraoBtnPanico(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	var cli Cliente

	if err := json.Unmarshal(body, &cli); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	if err := cli.SetDispPadraoBtnPanico(); err != nil {
		respApp.Erro(w, http.StatusBadRequest, err)
		return
	}

	respApp.OK(w)
}
