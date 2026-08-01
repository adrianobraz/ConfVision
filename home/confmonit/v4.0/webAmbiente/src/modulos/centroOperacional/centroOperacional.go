package centroOperacional

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"webAmbiente/src/config"
	"webAmbiente/src/modulos/home"
	"webAmbiente/src/modulos/monitor"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/centroOperacional/page",
		Metodo:   http.MethodGet,
		Controle: page,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/global/page",
		Metodo:   http.MethodGet,
		Controle: globalPage,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/mapas",
		Metodo:   http.MethodPost,
		Controle: mapasCO,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/eventos",
		Metodo:   http.MethodPost,
		Controle: eventos,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/mapasStatus",
		Metodo:   http.MethodPost,
		Controle: mapasStatus,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/configCliente",
		Metodo:   http.MethodPost,
		Controle: configClienteGet,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/configCliente/salvar",
		Metodo:   http.MethodPost,
		Controle: configClienteSalvar,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/dispositivoStatus",
		Metodo:   http.MethodPost,
		Controle: dispositivoStatus,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/comando",
		Metodo:   http.MethodPost,
		Controle: comando,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/setoresMapas",
		Metodo:   http.MethodPost,
		Controle: setoresMapas,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/processoDetalhe",
		Metodo:   http.MethodPost,
		Controle: processoDetalhe,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/finalizarProcesso",
		Metodo:   http.MethodPost,
		Controle: finalizarProcessoHandler,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/contadoresMapa",
		Metodo:   http.MethodPost,
		Controle: contadoresMapa,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/timeline/listar",
		Metodo:   http.MethodPost,
		Controle: timelineListar,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/timeline/salvar",
		Metodo:   http.MethodPost,
		Controle: timelineSalvar,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/rastro/listar",
		Metodo:   http.MethodPost,
		Controle: rastroListar,
		Seguro:   true,
	},
	{
		Uri:      "/centroOperacional/rastro/salvar",
		Metodo:   http.MethodPost,
		Controle: rastroSalvar,
		Seguro:   true,
	},
}

func page(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/centroOperacional/centroOperacional.html",
		"public/templates/components/pagina.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "erro ao carregar template", http.StatusInternalServerError)
		return
	}

	idCliente := strings.TrimSpace(r.URL.Query().Get("idCliente"))
	nomeCliente := strings.TrimSpace(r.URL.Query().Get("nome"))
	idFranqueado := strings.TrimSpace(r.URL.Query().Get("idFranqueado"))

	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Centro Operacional", config.TituloSite),
		NavbarTitulo: "Centro Operacional",
		IdCliente:    idCliente,
		NomeCliente:  nomeCliente,
		IdFranqueado: idFranqueado,
		ModoGlobal:   "N",
	}
	_ = page.ExecuteTemplate(w, "centroOperacional.html", d)
}

func globalPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/centroOperacional/centroOperacional.html",
		"public/templates/components/pagina.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "erro ao carregar template", http.StatusInternalServerError)
		return
	}

	op, _ := seguranca.LerOperador(r)
	titulo := "Todos os clientes"
	switch op["userTipo"] {
	case "FRA":
		if op["nomeVinculo"] != "" {
			titulo = op["nomeVinculo"]
		}
	case "REP":
		titulo = "Representante — todos os franqueados"
	case "CEN":
		titulo = "Central — todos os franqueados"
	}

	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Centro Operacional", config.TituloSite),
		NavbarTitulo: "Centro Operacional",
		NomeCliente:  titulo,
		ModoGlobal:   "S",
		UserTipo:     op["userTipo"],
	}
	_ = page.ExecuteTemplate(w, "centroOperacional.html", d)
}

func mapasCO(w http.ResponseWriter, r *http.Request) {
	op, err := seguranca.LerOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	lista, err := home.ListarMapasOperador(op)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(lista) == 0 {
		resposta.JsonVazio(w)
		return
	}
	resposta.JsonDados(w, http.StatusOK, lista)
}

type reqCliente struct {
	IdCliente    string `json:"idCliente"`
	IdFranqueado string `json:"idFranqueado"`
}

func lerReqCliente(w http.ResponseWriter, r *http.Request) (reqCliente, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return reqCliente{}, false
	}
	var req reqCliente
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return reqCliente{}, false
	}
	if strings.TrimSpace(req.IdCliente) == "" {
		resposta.Erro(w, http.StatusBadRequest, fmt.Errorf("idCliente obrigatorio"))
		return reqCliente{}, false
	}
	idFranq, err := seguranca.ResolverIdFranqueado(r, req.IdFranqueado, req.IdCliente)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return reqCliente{}, false
	}
	req.IdFranqueado = idFranq
	return req, true
}

func eventos(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdCliente    string `json:"idCliente"`
		IdFranqueado string `json:"idFranqueado"`
		Global       bool   `json:"global"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var lista []ProcessoFila
	if req.Global || strings.TrimSpace(req.IdCliente) == "" {
		op, err := seguranca.LerOperador(r)
		if err != nil {
			resposta.Erro(w, http.StatusUnauthorized, err)
			return
		}
		lista, err = listarProcessosOperador(op, 100)
		if err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
	} else {
		idFranq, err := seguranca.ResolverIdFranqueado(r, req.IdFranqueado, req.IdCliente)
		if err != nil {
			resposta.Erro(w, http.StatusUnauthorized, err)
			return
		}
		lista, err = listarProcessosCliente(req.IdCliente, idFranq, 50)
		if err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
	}

	if len(lista) == 0 {
		resposta.JsonVazio(w)
		return
	}
	resposta.JsonDados(w, http.StatusOK, lista)
}

func mapasStatus(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Ids []int `json:"ids"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	st, err := monitor.StatusMapasBatch(req.Ids)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonDados(w, http.StatusOK, st)
}

func configClienteGet(w http.ResponseWriter, r *http.Request) {
	req, ok := lerReqCliente(w, r)
	if !ok {
		return
	}
	cfg, err := xano.BuscarCoClienteConfig(req.IdCliente, req.IdFranqueado)
	if err != nil || cfg.Id == 0 {
		resposta.JsonVazio(w)
		return
	}
	resposta.JsonDados(w, http.StatusOK, cfg)
}

func configClienteSalvar(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdCliente    string `json:"idCliente"`
		IdFranqueado string `json:"idFranqueado"`
		CorAvatar    string `json:"corAvatar"`
		Iniciais     string `json:"iniciais"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	idFranq, err := seguranca.ResolverIdFranqueado(r, req.IdFranqueado, req.IdCliente)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	cfg, err := xano.SalvarCoClienteConfig(req.IdCliente, idFranq, req.CorAvatar, req.Iniciais)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonDados(w, http.StatusOK, cfg)
}

func dispositivoStatus(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdDispositivo string `json:"idDispositivo"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	armado, senha, part, err := statusDispositivo(req.IdDispositivo)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonDados(w, http.StatusOK, map[string]string{
		"armado":   armado,
		"senha":    senha,
		"particao": part,
	})
}

func comando(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdDispositivo string `json:"idDispositivo"`
		Acao          string `json:"acao"`
		Senha         string `json:"senha"`
		Particao      string `json:"particao"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	armar := strings.EqualFold(strings.TrimSpace(req.Acao), "armar")
	if err := enviarComando(req.IdDispositivo, req.Senha, req.Particao, armar); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonOK(w)
}

func setoresMapas(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Ids []int `json:"ids"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	type setorRef struct {
		IdSetor       string `json:"idSetor"`
		IdDispositivo string `json:"idDispositivo"`
		ZonaUser      string `json:"zonaUser"`
		Particao      string `json:"particao"`
	}

	out := make(map[string][]setorRef)
	for _, idMapa := range req.Ids {
		setores, err := xano.ListarSetoresPorMapa(idMapa)
		if err != nil {
			continue
		}
		lista := make([]setorRef, 0, len(setores))
		for _, s := range setores {
			lista = append(lista, setorRef{
				IdSetor:       s.IdSetor,
				IdDispositivo: s.IdDispositivo,
				ZonaUser:      s.Numero,
				Particao:      s.Particao,
			})
		}
		out[strconv.Itoa(idMapa)] = lista
	}
	resposta.JsonDados(w, http.StatusOK, out)
}

func processoDetalhe(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdProcesso   string `json:"idProcesso"`
		IdFranqueado string `json:"idFranqueado"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	idFranq, err := seguranca.ResolverIdFranqueado(r, req.IdFranqueado, "")
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	lista, err := listarEventosAgrupados(req.IdProcesso, idFranq)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(lista) == 0 {
		resposta.JsonVazio(w)
		return
	}
	resposta.JsonDados(w, http.StatusOK, lista)
}

func finalizarProcessoHandler(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdProcesso   string `json:"idProcesso"`
		IdCliente    string `json:"idCliente"`
		Descricao    string `json:"descricao"`
		IdOperador   string `json:"idOperador"`
		NomeOperador string `json:"nomeOperador"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.IdCliente) != "" {
		if _, err := seguranca.ResolverIdFranqueado(r, "", req.IdCliente); err != nil {
			resposta.Erro(w, http.StatusUnauthorized, err)
			return
		}
	}
	if err := finalizarProcesso(req.IdProcesso, req.IdCliente, req.Descricao, req.IdOperador, req.NomeOperador); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonOK(w)
}

func contadoresMapa(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		MapaId int `json:"mapaId"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req reqBody
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if req.MapaId <= 0 {
		resposta.Erro(w, http.StatusBadRequest, fmt.Errorf("mapaId obrigatorio"))
		return
	}
	dados, err := contadoresDoMapa(req.MapaId)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonDados(w, http.StatusOK, dados)
}
