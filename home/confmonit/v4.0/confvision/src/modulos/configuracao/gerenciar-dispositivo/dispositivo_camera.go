package gerenciarDispositivo

import (
	"bytes"
	"confvision/src/auxiliar"
	"confvision/src/config"
	"confvision/src/seguranca"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const fabricanteCameraID = "7"

type payloadCameraInserir struct {
	IDCliente     string `json:"idCliente"`
	IDFabricante  string `json:"IdFabricante"`
	IDModelo      string `json:"idModelo"`
	Nome          string `json:"nome"`
	IDFranqueado  string `json:"idFranqueado"`
	ProvedorVideo string `json:"provedorVideo"`
}

type payloadCameraAlterar struct {
	IDDispositivo string `json:"idDispositivo"`
	IDFabricante  string `json:"IdFabricante"`
	IDModelo      string `json:"idModelo"`
	Nome          string `json:"nome"`
}

type respostaAPIDadosString struct {
	Status string `json:"status"`
	Dados  string `json:"dados"`
}

type dispositivoAPI struct {
	IDDispositivo     string `json:"idDispositivo"`
	IDCliente         string `json:"idCliente"`
	IDFabricante      string `json:"idFabricante"`
	IDModelo          string `json:"idModelo"`
	Tipo              string `json:"tipo"`
	Nome              string `json:"nome"`
	Particao          string `json:"particao"`
	Conta             string `json:"conta"`
	IDFisico1         string `json:"idFisico1"`
	IDFisico2         string `json:"idFisico2"`
	KeepAlive         string `json:"keepAlive"`
	MsgAtendente      string `json:"msgAtendente"`
	Senha             string `json:"senha"`
	SenhaVerbal       string `json:"senhaVerbal"`
	ContraSenhaVerbal string `json:"contraSenhaVerbal"`
}

type respostaAPIDadosDispositivo struct {
	Status string         `json:"status"`
	Dados  dispositivoAPI `json:"dados"`
}

func GerenciarDispositivoInserirCamera(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var req payloadCameraInserir
	if erro := json.Unmarshal(corpo, &req); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	req.IDCliente = strings.TrimSpace(req.IDCliente)
	req.IDModelo = strings.TrimSpace(req.IDModelo)
	req.Nome = strings.TrimSpace(req.Nome)
	req.IDFranqueado = strings.TrimSpace(req.IDFranqueado)
	req.IDFabricante = strings.TrimSpace(req.IDFabricante)
	if req.IDFabricante == "" {
		req.IDFabricante = fabricanteCameraID
	}

	if req.IDCliente == "" || req.IDModelo == "" || req.Nome == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("cliente, modelo e nome sao obrigatorios"))
		return
	}
	if req.IDFranqueado == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("idFranqueado e obrigatorio"))
		return
	}
	if req.IDFabricante != fabricanteCameraID {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("fabricante deve ser CAMERA (7)"))
		return
	}

	conta, erro := gerarContaCamera(r, req.IDFranqueado)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	payloadAPI := map[string]interface{}{
		"idCliente":         req.IDCliente,
		"idFabricante":      req.IDFabricante,
		"idModelo":          req.IDModelo,
		"nome":              strings.ToUpper(req.Nome),
		"particao":          "1",
		"conta":             conta,
		"idFisico1":         "N/A",
		"idFisico2":         "N/A",
		"keepAlive":         "0",
		"tipo":              "MASTER",
		"senhaVerbal":       "",
		"contraSenhaVerbal": "",
		"msgAtendente":      "",
		"senha":             "",
	}

	bodyAPI, erro := json.Marshal(payloadAPI)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/insere", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(bodyAPI))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpoResp, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	idDispositivo, erro := extrairIdDispositivoInserido(corpoResp)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	provedorVideo := normalizarProvedorVideo(req.ProvedorVideo)
	if erro := gravarProvedorVideo(idDispositivo, provedorVideo); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaAPP(w, corpoResp)
}

func GerenciarDispositivoAlterarCamera(w http.ResponseWriter, r *http.Request) {
	corpo, erro := io.ReadAll(r.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var req payloadCameraAlterar
	if erro := json.Unmarshal(corpo, &req); erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	req.IDDispositivo = strings.TrimSpace(req.IDDispositivo)
	req.IDModelo = strings.TrimSpace(req.IDModelo)
	req.Nome = strings.TrimSpace(req.Nome)
	req.IDFabricante = strings.TrimSpace(req.IDFabricante)
	if req.IDFabricante == "" {
		req.IDFabricante = fabricanteCameraID
	}

	if req.IDDispositivo == "" || req.IDModelo == "" || req.Nome == "" {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("idDispositivo, modelo e nome sao obrigatorios"))
		return
	}
	if req.IDFabricante != fabricanteCameraID {
		auxiliar.RespostaErro(w, http.StatusBadRequest, errors.New("fabricante deve ser CAMERA (7)"))
		return
	}

	atual, erro := buscarDispositivoCamera(r, req.IDDispositivo)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	payloadAPI := map[string]interface{}{
		"idDispositivo":     req.IDDispositivo,
		"idFabricante":      req.IDFabricante,
		"idModelo":          req.IDModelo,
		"nome":              strings.ToUpper(req.Nome),
		"particao":          valorOuPadrao(atual.Particao, "1"),
		"conta":             valorOuPadrao(atual.Conta, "0000"),
		"idFisico1":         valorOuPadrao(atual.IDFisico1, "N/A"),
		"idFisico2":         valorOuPadrao(atual.IDFisico2, "N/A"),
		"keepAlive":         valorOuPadrao(atual.KeepAlive, "0"),
		"tipo":              valorOuPadrao(atual.Tipo, "MASTER"),
		"senhaVerbal":       atual.SenhaVerbal,
		"contraSenhaVerbal": atual.ContraSenhaVerbal,
		"msgAtendente":      atual.MsgAtendente,
		"senha":             atual.Senha,
	}

	bodyAPI, erro := json.Marshal(payloadAPI)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	url := fmt.Sprintf("%s/v4/dispositivo/alteraById", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(bodyAPI))
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		auxiliar.TratarStatusCodeDeErro(w, resp)
		return
	}

	corpoResp, erro := io.ReadAll(resp.Body)
	if erro != nil {
		auxiliar.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	auxiliar.RespostaAPP(w, corpoResp)
}

func gerarContaCamera(r *http.Request, idFranqueado string) (string, error) {
	body, erro := json.Marshal(map[string]string{"idFranqueado": idFranqueado})
	if erro != nil {
		return "", erro
	}

	url := fmt.Sprintf("%s/v4/dispositivo/gerarContaByIdFranqueado", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		return "", erro
	}
	defer resp.Body.Close()

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		return "", erro
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("falha ao gerar conta: %s", strings.TrimSpace(string(corpo)))
	}

	var resposta respostaAPIDadosString
	if erro := json.Unmarshal(corpo, &resposta); erro != nil {
		return "", erro
	}

	conta := strings.TrimSpace(resposta.Dados)
	if conta == "" || strings.Contains(strings.ToLower(conta), "limite") {
		return "", errors.New("nao foi possivel gerar conta alarme para o dispositivo camera")
	}

	return conta, nil
}

func buscarDispositivoCamera(r *http.Request, idDispositivo string) (*dispositivoAPI, error) {
	body, erro := json.Marshal(map[string]string{"idDispositivo": idDispositivo})
	if erro != nil {
		return nil, erro
	}

	url := fmt.Sprintf("%s/v4/dispositivo/getDadosById", config.ApiUrl)
	resp, erro := seguranca.RequisiacaoAutenticada(r, http.MethodPost, url, bytes.NewBuffer(body))
	if erro != nil {
		return nil, erro
	}
	defer resp.Body.Close()

	corpo, erro := io.ReadAll(resp.Body)
	if erro != nil {
		return nil, erro
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("falha ao buscar dispositivo: %s", strings.TrimSpace(string(corpo)))
	}

	var resposta respostaAPIDadosDispositivo
	if erro := json.Unmarshal(corpo, &resposta); erro != nil {
		return nil, erro
	}
	if strings.TrimSpace(resposta.Dados.IDDispositivo) == "" {
		return nil, errors.New("dispositivo nao encontrado")
	}

	return &resposta.Dados, nil
}

func valorOuPadrao(valor, padrao string) string {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return padrao
	}
	return valor
}
