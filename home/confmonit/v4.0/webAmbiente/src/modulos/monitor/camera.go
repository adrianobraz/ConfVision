package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/config"
	"webAmbiente/src/resposta"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

func enrichSetoresMapa(setores []tipos.MapaSetor, idFranqueado string) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	for i := range setores {
		dados, _ := auxiliar.BuscarDadosSetorAlarme(setores[i].IdSetor)
		setores[i].TipoSetor = dados.TipoSetor
		setores[i].Camera = dados.Camera
		setores[i].Numero = dados.Numero
		setores[i].Particao = dados.Particao
		setores[i].Icone = auxiliar.ResolverIconeSalvo(setores[i].Icone, dados.TipoSetor)
		if idFranqueado != "" {
			setores[i].IdFranqueado = idFranqueado
		}
	}
}

func confVisionConfig(w http.ResponseWriter, r *http.Request) {
	resposta.JsonDados(w, http.StatusOK, map[string]string{
		"hlsBase": config.ConfVisionHlsBase,
	})
}

func cameraAoVivo(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IdSetor       string `json:"id_setor"`
		IdDispositivo string `json:"id_dispositivo"`
		IdFranqueado  string `json:"id_franqueado"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	idSetor := strings.ToUpper(strings.TrimSpace(req.IdSetor))
	if idSetor == "" {
		resposta.Erro(w, http.StatusBadRequest, fmt.Errorf("id_setor obrigatorio"))
		return
	}

	dados, err := auxiliar.BuscarDadosSetorAlarme(idSetor)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if !auxiliar.SetorTemCamera(dados.Camera) {
		resposta.JsonVazio(w)
		return
	}

	idDispositivo := strings.TrimSpace(req.IdDispositivo)
	if idDispositivo == "" {
		resposta.Erro(w, http.StatusBadRequest, fmt.Errorf("id_dispositivo obrigatorio"))
		return
	}

	cam, err := xano.BuscarCameraPorSetor(
		idSetor,
		idDispositivo,
		strings.TrimSpace(req.IdFranqueado),
		dados.Particao,
		dados.Numero,
	)
	if err != nil {
		resposta.Erro(w, http.StatusBadGateway, err)
		return
	}
	if cam == nil || cam.Id <= 0 {
		resposta.JsonVazio(w)
		return
	}

	hlsUrl := ""
	if config.ConfVisionHlsBase != "" {
		hlsUrl = fmt.Sprintf("%s/live/%d/index.m3u8", config.ConfVisionHlsBase, cam.Id)
	}

	resposta.JsonDados(w, http.StatusOK, map[string]interface{}{
		"camera": map[string]interface{}{
			"id":     cam.Id,
			"nome":   cam.Nome,
			"hlsUrl": hlsUrl,
		},
		"hlsBase": config.ConfVisionHlsBase,
	})
}
