package home

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/config"
	"webAmbiente/src/modulos/monitor"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/home",
		Metodo:   http.MethodGet,
		Controle: homePage,
		Seguro:   true,
	},
	{
		Uri:      "/home/clientesComMapa",
		Metodo:   http.MethodPost,
		Controle: clientesComMapa,
		Seguro:   true,
	},
	{
		Uri:      "/home/clientesFranqueado",
		Metodo:   http.MethodPost,
		Controle: clientesFranqueado,
		Seguro:   true,
	},
	{
		Uri:      "/home/franqueados",
		Metodo:   http.MethodPost,
		Controle: franqueados,
		Seguro:   true,
	},
	{
		Uri:      "/home/mapas",
		Metodo:   http.MethodPost,
		Controle: mapasHome,
		Seguro:   true,
	},
	{
		Uri:      "/home/mapasStatus",
		Metodo:   http.MethodPost,
		Controle: mapasStatusHome,
		Seguro:   true,
	},
}

func homePage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/home/home.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	op, _ := seguranca.LerOperador(r)
	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Início", config.TituloSite),
		NavbarTitulo: "Início",
		UserTipo:     op["userTipo"],
		UserMaster:   op["userMaster"],
		UserNome:     op["userNome"],
		NomeVinculo:  op["nomeVinculo"],
		IdVinculo:    op["idVinculo"],
	}
	if err := page.ExecuteTemplate(w, "home.html", d); err != nil {
		fmt.Println(err)
	}
}

type clienteComMapa struct {
	IdCliente   string `json:"idCliente"`
	NomeCliente string `json:"nomeCliente"`
	QtdMapas    int    `json:"qtdMapas"`
}

type franqueadoItem struct {
	IdFranqueado   string `json:"idFranqueado"`
	NomeFranqueado string `json:"nomeFranqueado"`
}

type mapaHomeItem struct {
	Id             int    `json:"id"`
	Descricao      string `json:"descricao"`
	IdCliente      string `json:"idCliente"`
	NomeCliente    string `json:"nomeCliente"`
	IdFranqueado   string `json:"idFranqueado"`
	NomeFranqueado string `json:"nomeFranqueado"`
	ImagemUrl      string `json:"imagem_url"`
}

func mapasHome(w http.ResponseWriter, r *http.Request) {
	op, err := seguranca.LerOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}

	lista, err := listarMapasOperador(op)
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

func listarMapasOperador(op map[string]string) ([]mapaHomeItem, error) {
	tipo := op["userTipo"]
	idVinculo := normIDFranqueado(op["idVinculo"])

	var mapas []tipos.MapaAmbiente
	var err error
	var filtroFranqueados map[string]bool

	switch tipo {
	case "FRA":
		mapas, err = xano.QueryAllMapaAmbiente(idVinculo)
	case "REP":
		fras, e := listarFranqueados("REP", idVinculo, false)
		if e != nil {
			return nil, e
		}
		filtroFranqueados = make(map[string]bool, len(fras))
		for _, f := range fras {
			filtroFranqueados[normIDFranqueado(f.IdFranqueado)] = true
		}
		mapas, err = xano.QueryAllMapaAmbiente("")
	default:
		mapas, err = xano.QueryAllMapaAmbiente("")
	}
	if err != nil {
		return nil, fmt.Errorf("xano mapas: %w", err)
	}

	lista := make([]mapaHomeItem, 0, len(mapas))
	for _, m := range mapas {
		if m.Ativo != nil && !*m.Ativo {
			continue
		}
		idFra := normIDFranqueado(m.IdFranqueado)
		if filtroFranqueados != nil && !filtroFranqueados[idFra] {
			continue
		}
		lista = append(lista, mapaHomeItem{
			Id:             m.Id,
			Descricao:      m.Descricao,
			IdCliente:      normIDFranqueado(m.IdCliente),
			NomeCliente:    m.NomeCliente,
			IdFranqueado:   idFra,
			NomeFranqueado: m.NomeFranqueado,
			ImagemUrl:      m.ImagemUrl,
		})
	}

	if err := preencherNomesClientesMapas(lista); err != nil {
		return nil, err
	}

	sort.SliceStable(lista, func(i, j int) bool {
		ni := strings.ToLower(lista[i].NomeCliente)
		nj := strings.ToLower(lista[j].NomeCliente)
		if ni != nj {
			return ni < nj
		}
		return strings.ToLower(lista[i].Descricao) < strings.ToLower(lista[j].Descricao)
	})

	return lista, nil
}

// ListarMapasOperador expõe listagem de mapas por perfil do operador (CO global).
func ListarMapasOperador(op map[string]string) ([]mapaHomeItem, error) {
	return listarMapasOperador(op)
}

// preencherNomesClientesMapas resolve no MySQL os nomes de clientes ausentes,
// agrupando por franqueado (ID_Cliente pode nao ser unico entre franqueados).
func preencherNomesClientesMapas(lista []mapaHomeItem) error {
	faltam := make(map[string]map[string]struct{})
	for _, it := range lista {
		if it.NomeCliente != "" || it.IdCliente == "" {
			continue
		}
		if faltam[it.IdFranqueado] == nil {
			faltam[it.IdFranqueado] = make(map[string]struct{})
		}
		faltam[it.IdFranqueado][it.IdCliente] = struct{}{}
	}
	if len(faltam) == 0 {
		return nil
	}

	nomesPorFra := make(map[string]map[string]string)
	for fra, set := range faltam {
		ids := make([]string, 0, len(set))
		for id := range set {
			ids = append(ids, id)
		}
		nomes, err := buscarNomesClientesMySQL(fra, ids)
		if err != nil {
			return err
		}
		nomesPorFra[fra] = nomes
	}

	for i := range lista {
		if lista[i].NomeCliente != "" {
			continue
		}
		if nomes := nomesPorFra[lista[i].IdFranqueado]; nomes != nil {
			if n, ok := nomes[lista[i].IdCliente]; ok {
				lista[i].NomeCliente = n
				continue
			}
		}
		lista[i].NomeCliente = lista[i].IdCliente
	}
	return nil
}

func mapasStatusHome(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Ids []int `json:"ids"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req reqBody
	if len(body) > 0 {
		_ = json.Unmarshal(body, &req)
	}

	if len(req.Ids) == 0 {
		resposta.JsonDados(w, http.StatusOK, map[string]string{})
		return
	}

	status, err := monitor.StatusMapasBatch(req.Ids)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	out := make(map[string]string, len(status))
	for id, st := range status {
		out[strconv.Itoa(id)] = st
	}

	resposta.JsonDados(w, http.StatusOK, out)
}

func franqueados(w http.ResponseWriter, r *http.Request) {
	op, err := seguranca.LerOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}

	tipo := op["userTipo"]
	if tipo != "CEN" && tipo != "REP" {
		resposta.Erro(w, http.StatusForbidden, errors.New("listagem de franqueados apenas para central ou representante"))
		return
	}

	lista, err := listarFranqueados(tipo, op["idVinculo"], !seguranca.EhMaster(r))
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

func listarFranqueados(userTipo, idVinculo string, somenteComMapa bool) ([]franqueadoItem, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var tab *sql.Rows
	if userTipo == "REP" {
		tab, err = db.Query(`
			SELECT franqueado.ID_Franqueado, franqueado.NomeFantasia, franqueado.RazaoSocial
			FROM franqueado
			WHERE franqueado.ID_Representante = ?
			AND franqueado.DataCancelamento IS NULL
			ORDER BY COALESCE(NULLIF(franqueado.NomeFantasia, ''), franqueado.RazaoSocial)
		`, idVinculo)
	} else {
		tab, err = db.Query(`
			SELECT franqueado.ID_Franqueado, franqueado.NomeFantasia, franqueado.RazaoSocial
			FROM franqueado
			WHERE franqueado.DataCancelamento IS NULL
			ORDER BY COALESCE(NULLIF(franqueado.NomeFantasia, ''), franqueado.RazaoSocial)
		`)
	}
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	lista := make([]franqueadoItem, 0)
	for tab.Next() {
		var id, fantasia, razao sql.NullString
		if err := tab.Scan(&id, &fantasia, &razao); err != nil {
			return nil, err
		}
		nome := razao.String
		if fantasia.String != "" {
			nome = fantasia.String
		}
		lista = append(lista, franqueadoItem{
			IdFranqueado:   id.String,
			NomeFranqueado: nome,
		})
	}

	if !somenteComMapa {
		return lista, nil
	}

	comMapa, err := franqueadosComMapa(userTipo, idVinculo)
	if err != nil {
		return nil, err
	}

	filtrada := make([]franqueadoItem, 0)
	for _, f := range lista {
		if comMapa[normIDFranqueado(f.IdFranqueado)] > 0 {
			filtrada = append(filtrada, f)
		}
	}
	return filtrada, nil
}

func franqueadosComMapa(userTipo, idVinculo string) (map[string]int, error) {
	out := make(map[string]int)

	idsRep := make(map[string]bool)
	if userTipo == "REP" {
		todos, err := listarFranqueados("REP", idVinculo, false)
		if err != nil {
			return nil, err
		}
		for _, f := range todos {
			idsRep[normIDFranqueado(f.IdFranqueado)] = true
		}
	}

	mapas, err := xano.QueryAllMapaAmbiente("")
	if err != nil {
		return nil, fmt.Errorf("xano mapas: %w", err)
	}

	for _, m := range mapas {
		idFra := normIDFranqueado(m.IdFranqueado)
		if idFra == "" {
			continue
		}
		if userTipo == "REP" && !idsRep[idFra] {
			continue
		}
		out[idFra]++
	}
	return out, nil
}

func clientesFranqueado(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdFranqueado string `json:"idFranqueado"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req reqBody
	if len(body) > 0 {
		_ = json.Unmarshal(body, &req)
	}

	op, err := seguranca.LerOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	if (op["userTipo"] == "CEN" || op["userTipo"] == "REP") && req.IdFranqueado == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("selecione um franqueado"))
		return
	}

	idFranqueado, err := seguranca.IdFranqueadoOperador(r, req.IdFranqueado)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}

	var lista []clienteComMapa
	if seguranca.EhMaster(r) {
		lista, err = listarClientesFranqueado(idFranqueado)
	} else {
		lista, err = listarClientesComMapa(idFranqueado)
	}
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

func clientesComMapa(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		IdFranqueado string `json:"idFranqueado"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	var req reqBody
	if len(body) > 0 {
		_ = json.Unmarshal(body, &req)
	}

	idFranqueado, err := seguranca.IdFranqueadoOperador(r, req.IdFranqueado)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}

	lista, err := listarClientesComMapa(idFranqueado)
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

func listarClientesFranqueado(idFranqueado string) ([]clienteComMapa, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT cliente.ID_Cliente, cliente.Nome, cliente.Nick
		FROM cliente
		WHERE cliente.ID_Franqueado = ?
		ORDER BY COALESCE(NULLIF(cliente.Nick, ''), cliente.Nome)
	`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	qtdPorCliente, err := contarMapasPorCliente(idFranqueado)
	if err != nil {
		return nil, fmt.Errorf("xano mapas: %w", err)
	}

	lista := make([]clienteComMapa, 0)
	for tab.Next() {
		var id, nome, nick sql.NullString
		if err := tab.Scan(&id, &nome, &nick); err != nil {
			return nil, err
		}
		n := nome.String
		if nick.String != "" {
			n = nick.String
		}
		idCli := normIDFranqueado(id.String)
		lista = append(lista, clienteComMapa{
			IdCliente:   idCli,
			NomeCliente: n,
			QtdMapas:    qtdPorCliente[idCli],
		})
	}
	return lista, nil
}

func contarMapasPorCliente(idFranqueado string) (map[string]int, error) {
	mapas, err := xano.QueryAllMapaAmbiente(idFranqueado)
	if err != nil {
		return nil, err
	}
	out := make(map[string]int)
	for _, m := range mapas {
		idCli := normIDFranqueado(m.IdCliente)
		if idCli != "" {
			out[idCli]++
		}
	}
	return out, nil
}

func listarClientesComMapa(idFranqueado string) ([]clienteComMapa, error) {
	mapas, err := xano.QueryAllMapaAmbiente(idFranqueado)
	if err != nil {
		return nil, fmt.Errorf("xano mapas: %w", err)
	}

	porCliente := make(map[string]*clienteComMapa)
	ordem := make([]string, 0)

	for _, m := range mapas {
		idCli := normIDFranqueado(m.IdCliente)
		if idCli == "" {
			continue
		}
		if _, ok := porCliente[idCli]; !ok {
			porCliente[idCli] = &clienteComMapa{
				IdCliente:   idCli,
				NomeCliente: m.NomeCliente,
				QtdMapas:    0,
			}
			ordem = append(ordem, idCli)
		}
		porCliente[idCli].QtdMapas++
		if porCliente[idCli].NomeCliente == "" && m.NomeCliente != "" {
			porCliente[idCli].NomeCliente = m.NomeCliente
		}
	}

	lista := make([]clienteComMapa, 0, len(ordem))
	faltamNome := make([]string, 0)

	for _, id := range ordem {
		c := *porCliente[id]
		if c.NomeCliente == "" {
			faltamNome = append(faltamNome, id)
		}
		lista = append(lista, c)
	}

	if len(faltamNome) > 0 {
		nomes, err := buscarNomesClientesMySQL(idFranqueado, faltamNome)
		if err != nil {
			return nil, err
		}
		for i := range lista {
			if lista[i].NomeCliente == "" {
				if n, ok := nomes[lista[i].IdCliente]; ok {
					lista[i].NomeCliente = n
				} else {
					lista[i].NomeCliente = lista[i].IdCliente
				}
			}
		}
	}

	return lista, nil
}

func buscarNomesClientesMySQL(idFranqueado string, ids []string) (map[string]string, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = strings.TrimSuffix(placeholders, ",")

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, idFranqueado)
	for _, id := range ids {
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT cliente.ID_Cliente, cliente.Nome, cliente.Nick
		FROM cliente
		WHERE cliente.ID_Franqueado = ?
		AND cliente.ID_Cliente IN (%s)
	`, placeholders)

	tab, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	nomes := make(map[string]string)
	for tab.Next() {
		var id, nome, nick sql.NullString
		if err := tab.Scan(&id, &nome, &nick); err != nil {
			return nil, err
		}
		n := nome.String
		if nick.String != "" {
			n = nick.String
		}
		nomes[id.String] = n
	}
	return nomes, nil
}

func normIDFranqueado(s string) string {
	return strings.TrimSpace(s)
}
