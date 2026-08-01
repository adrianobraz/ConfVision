package editor

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/config"
	"webAmbiente/src/resposta"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/editor/page",
		Metodo:   http.MethodGet,
		Controle: editorPage,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/editor/carregar",
		Metodo:   http.MethodPost,
		Controle: carregar,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/editor/dispositivos",
		Metodo:   http.MethodPost,
		Controle: dispositivos,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/editor/setores",
		Metodo:   http.MethodPost,
		Controle: setores,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/editor/salvar",
		Metodo:   http.MethodPost,
		Controle: salvar,
		Seguro:   true,
		Master:   true,
	},
}

func editorPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/editor/editor.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Editor de setores", config.TituloSite),
		NavbarTitulo: "Editor de setores",
	}
	if err := page.ExecuteTemplate(w, "editor.html", d); err != nil {
		fmt.Println(err)
	}
}

func carregar(w http.ResponseWriter, r *http.Request) {
	id, err := parseMapaId(r)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	mapa, setores, err := xano.CarregarMapaComSetores(id)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	posicoes := make([]map[string]interface{}, 0, len(setores))
	for _, s := range setores {
		label := s.Label
		if label == "" {
			label = s.SetorNome
		}
		tipoSetor, _ := auxiliar.BuscarTipoSetor(s.IdSetor)
		icone := auxiliar.ResolverIconeSalvo(s.Icone, tipoSetor)
		posicoes = append(posicoes, map[string]interface{}{
			"id":               s.Id,
			"id_setor":         s.IdSetor,
			"id_dispositivo":   s.IdDispositivo,
			"label":            label,
			"setor_nome":       s.SetorNome,
			"dispositivo_nome": s.DispositivoNome,
			"pos_x":            s.PosX,
			"pos_y":            s.PoxY,
			"tipo_setor":       tipoSetor,
			"icone":            icone,
		})
	}

	resposta.JsonDados(w, http.StatusOK, map[string]interface{}{
		"mapa":     mapa,
		"posicoes": posicoes,
	})
}

func dispositivos(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IdCliente string `json:"idCliente"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if req.IdCliente == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("idCliente obrigatorio"))
		return
	}

	lista, err := listarDispositivos(req.IdCliente)
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

func setores(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IdDispositivo string `json:"idDispositivo"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if req.IdDispositivo == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("idDispositivo obrigatorio"))
		return
	}

	lista, err := listarSetores(req.IdDispositivo)
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

type stringFlex string

func (s *stringFlex) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*s = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*s = stringFlex(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(b, &num); err == nil {
		*s = stringFlex(num.String())
		return nil
	}
	return fmt.Errorf("stringFlex: valor invalido %s", string(b))
}

type itemPosicao struct {
	Id            int        `json:"id"`
	IdSetor       stringFlex `json:"id_setor"`
	IdDispositivo stringFlex `json:"id_dispositivo"`
	Label         stringFlex `json:"label"`
	Icone         stringFlex `json:"icone"`
	PosX          float64    `json:"pos_x"`
	PosY          float64    `json:"pos_y"`
}

func salvar(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		MapaAmbienteId int           `json:"mapa_ambiente_id"`
		IdCliente      string        `json:"id_cliente"`
		IdClienteAlt   string        `json:"idCliente"`
		NomeCliente    string        `json:"nomeCliente"`
		Itens          []itemPosicao `json:"itens"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if req.MapaAmbienteId <= 0 {
		resposta.Erro(w, http.StatusBadRequest, errors.New("mapa_ambiente_id obrigatorio"))
		return
	}

	idCliente := req.IdCliente
	if idCliente == "" {
		idCliente = req.IdClienteAlt
	}
	if idCliente == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("idCliente obrigatorio"))
		return
	}

	mapa, err := xano.GetMapaAmbiente(req.MapaAmbienteId)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	// Remove setores que saíram do mapa
	existentes, err := xano.ListarSetoresPorMapa(req.MapaAmbienteId)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	enviados := make(map[int]bool)
	for _, item := range req.Itens {
		if item.Id > 0 {
			enviados[item.Id] = true
		}
	}
	for _, s := range existentes {
		if !enviados[s.Id] {
			if err := xano.ExcluirMapaSetor(s.Id); err != nil {
				resposta.Erro(w, http.StatusBadRequest, err)
				return
			}
		}
	}

	nomeCliente := mapa.NomeCliente
	if req.NomeCliente != "" {
		nomeCliente = req.NomeCliente
	}

	for _, item := range req.Itens {
		if string(item.IdSetor) == "" {
			continue
		}

		dispNome, setorNome, err := buscarNomesSetor(string(item.IdDispositivo), string(item.IdSetor))
		if err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}

		label := string(item.Label)
		if label == "" {
			label = setorNome
		}

		tipoSetor, _ := auxiliar.BuscarTipoSetor(string(item.IdSetor))

		payload := tipos.MapaSetor{
			MapaAmbienteId:  req.MapaAmbienteId,
			IdSetor:         string(item.IdSetor),
			IdDispositivo:   string(item.IdDispositivo),
			IdCliente:       idCliente,
			SetorNome:       setorNome,
			DispositivoNome: dispNome,
			ClienteNome:     nomeCliente,
			PosX:            item.PosX,
			PoxY:            item.PosY,
			Label:           label,
			Icone:           auxiliar.ResolverIconeSalvo(string(item.Icone), tipoSetor),
		}

		if item.Id > 0 {
			if _, err := xano.AtualizarMapaSetor(item.Id, payload); err != nil {
				resposta.Erro(w, http.StatusBadRequest, err)
				return
			}
			continue
		}

		if _, err := xano.CriarMapaSetor(payload); err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
	}

	resposta.JsonOK(w)
}

func parseMapaId(r *http.Request) (int, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}

	var req struct {
		MapaAmbienteId  int    `json:"mapa_ambiente_id"`
		IdMapa          int    `json:"id_mapa"`
		Id              int    `json:"id"`
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

type dispositivoItem struct {
	IdDispositivo string `json:"idDispositivo"`
	Nome          string `json:"nome"`
	Conta         string `json:"conta"`
}

func listarDispositivos(idCliente string) ([]dispositivoItem, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT dispositivo.ID_Dispositivo, dispositivo.Nome, dispositivo.Conta
		FROM dispositivo
		WHERE dispositivo.ID_Cliente = ?
	`, idCliente)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	lista := make([]dispositivoItem, 0)
	for tab.Next() {
		var item dispositivoItem
		var conta sql.NullString
		if err := tab.Scan(&item.IdDispositivo, &item.Nome, &conta); err != nil {
			return nil, err
		}
		item.Conta = conta.String
		lista = append(lista, item)
	}
	return lista, nil
}

type setorItem struct {
	IdSetor   string `json:"idSetor"`
	Numero    string `json:"numero"`
	Nome      string `json:"nome"`
	TipoSetor string `json:"tipoSetor"`
}

func listarSetores(idDispositivo string) ([]setorItem, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT setorAlarme.ID_Setor, setorAlarme.Numero, setorAlarme.Nome, setorAlarme.TipoSetor
		FROM setorAlarme
		WHERE setorAlarme.ID_Dispositivo = ?
	`, idDispositivo)
	if err != nil {
		if auxiliar.ErroColunaTipoSetor(err) {
			return listarSetoresSemTipo(idDispositivo, db)
		}
		return nil, err
	}
	defer tab.Close()

	lista := make([]setorItem, 0)
	for tab.Next() {
		var item setorItem
		var numero, nome, tipoSetor sql.NullString
		if err := tab.Scan(&item.IdSetor, &numero, &nome, &tipoSetor); err != nil {
			return nil, err
		}
		item.Numero = numero.String
		item.Nome = nome.String
		item.TipoSetor = tipoSetor.String
		lista = append(lista, item)
	}
	return lista, nil
}

func listarSetoresSemTipo(idDispositivo string, db *sql.DB) ([]setorItem, error) {
	tab, err := db.Query(`
		SELECT setorAlarme.ID_Setor, setorAlarme.Numero, setorAlarme.Nome
		FROM setorAlarme
		WHERE setorAlarme.ID_Dispositivo = ?
	`, idDispositivo)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	lista := make([]setorItem, 0)
	for tab.Next() {
		var item setorItem
		var numero, nome sql.NullString
		if err := tab.Scan(&item.IdSetor, &numero, &nome); err != nil {
			return nil, err
		}
		item.Numero = numero.String
		item.Nome = nome.String
		lista = append(lista, item)
	}
	return lista, nil
}

func buscarNomesSetor(idDispositivo, idSetor string) (dispNome, setorNome string, err error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return "", "", err
	}
	defer db.Close()

	var nomeDisp sql.NullString
	_ = db.QueryRow(`
		SELECT dispositivo.Nome FROM dispositivo WHERE dispositivo.ID_Dispositivo = ?
	`, idDispositivo).Scan(&nomeDisp)

	var nomeSet sql.NullString
	_ = db.QueryRow(`
		SELECT setorAlarme.Nome FROM setorAlarme WHERE setorAlarme.ID_Setor = ?
	`, idSetor).Scan(&nomeSet)

	return nomeDisp.String, nomeSet.String, nil
}
