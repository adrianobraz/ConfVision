package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"
	"webAmbiente/src/config"
	"webAmbiente/src/resposta"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/monitor/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/monitor/carregar",
		Metodo:   http.MethodPost,
		Controle: carregarMapa,
		Seguro:   true,
	},
	{
		Uri:      "/monitor/status",
		Metodo:   http.MethodPost,
		Controle: statusMapa,
		Seguro:   true,
	},
	{
		Uri:      "/monitor/confvision/config",
		Metodo:   http.MethodGet,
		Controle: confVisionConfig,
		Seguro:   true,
	},
	{
		Uri:      "/monitor/camera",
		Metodo:   http.MethodPost,
		Controle: cameraAoVivo,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/monitor/monitor.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Monitor", config.TituloSite),
		NavbarTitulo: "Monitoramento",
	}
	if err := page.ExecuteTemplate(w, "monitor.html", d); err != nil {
		fmt.Println(err)
	}
}

func carregarMapa(w http.ResponseWriter, r *http.Request) {
	id, err := parseMapaAmbienteId(r)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	mapa, setores, err := xano.CarregarMapaComSetores(id)
	if err != nil {
		setores, err = xano.ListarSetoresPorMapa(id)
		if err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
		mapa = tipos.MapaAmbiente{Id: id}
	}

	enrichSetoresMapa(setores, mapa.IdFranqueado)

	resposta.JsonDados(w, http.StatusOK, map[string]interface{}{
		"mapa":    mapa,
		"setores": setores,
	})
}

func statusMapa(w http.ResponseWriter, r *http.Request) {
	id, err := parseMapaAmbienteId(r)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	setores, err := calcularStatusMapa(id)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.JSON(w, http.StatusOK, map[string]interface{}{
		"setores": setores,
	})
}

func parseMapaAmbienteId(r *http.Request) (int, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}

	var req struct {
		MapaAmbienteId int    `json:"mapa_ambiente_id"`
		IdMapa         int    `json:"idMapa"`
		Id             int    `json:"id"`
		MapaAmbienteIdS string `json:"mapa_ambiente_id_str"`
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return 0, err
		}
	}

	if req.MapaAmbienteId > 0 {
		return req.MapaAmbienteId, nil
	}
	if req.IdMapa > 0 {
		return req.IdMapa, nil
	}
	if req.Id > 0 {
		return req.Id, nil
	}
	if req.MapaAmbienteIdS != "" {
		return strconv.Atoi(req.MapaAmbienteIdS)
	}
	return 0, fmt.Errorf("mapa_ambiente_id obrigatorio")
}
