package ambiente

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/config"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/storage"
	"webAmbiente/src/tipos"
	"webAmbiente/src/xano"
)

var Rotas = []tipos.Rota{
	{
		Uri:      "/ambiente/page",
		Metodo:   http.MethodGet,
		Controle: cadastroPage,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/ambiente/carregar",
		Metodo:   http.MethodPost,
		Controle: carregar,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/ambiente/salvar",
		Metodo:   http.MethodPost,
		Controle: salvar,
		Seguro:   true,
		Master:   true,
	},
	{
		Uri:      "/ambiente/excluir",
		Metodo:   http.MethodPost,
		Controle: excluir,
		Seguro:   true,
		Master:   true,
	},
}

type salvarReq struct {
	Id           int    `json:"id"`
	Descricao    string `json:"descricao"`
	ImagemUrl    string `json:"imagem_url"`
	IdCliente    string `json:"idCliente"`
	IdFranqueado string `json:"idFranqueado"`
	NomeCliente  string `json:"nomeCliente"`
	Ordem        int    `json:"ordem"`
	Ativo        bool   `json:"ativo"`
}

func cadastroPage(w http.ResponseWriter, r *http.Request) {
	templ := []string{
		"public/templates/ambiente/ambiente.html",
		"public/templates/components/pagina.html",
		"public/templates/components/navbar.html",
		"public/templates/components/botoes.html",
	}

	page, err := template.ParseFiles(templ...)
	if err != nil {
		fmt.Println(err)
	}

	d := tipos.Page{
		Titulo:       fmt.Sprintf("%s - Cadastro de mapa", config.TituloSite),
		NavbarTitulo: "Cadastro de mapa",
	}
	if err := page.ExecuteTemplate(w, "ambiente.html", d); err != nil {
		fmt.Println(err)
	}
}

func carregar(w http.ResponseWriter, r *http.Request) {
	id, err := parseMapaId(r)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	mapa, err := xano.GetMapaAmbiente(id)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	resposta.JsonDados(w, http.StatusOK, mapa)
}

func salvar(w http.ResponseWriter, r *http.Request) {
	req, imagemRaw, err := parseSalvarRequest(r)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if req.Descricao == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("descricao obrigatoria"))
		return
	}
	if req.IdCliente == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("idCliente obrigatorio"))
		return
	}

	idFranqueado, err := seguranca.ResolverIdFranqueado(r, req.IdFranqueado, req.IdCliente)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}

	nomeCliente, nomeFranqueado, err := buscarNomesMySQL(req.IdCliente, idFranqueado)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if req.NomeCliente != "" {
		nomeCliente = req.NomeCliente
	}

	ativo := req.Ativo
	imagemUrl := storage.NormalizeImagemURL(req.ImagemUrl)
	if len(imagemRaw) > 0 {
		imagemUrl = ""
	}

	m := tipos.MapaAmbiente{
		Descricao:      req.Descricao,
		ImagemUrl:      imagemUrl,
		IdCliente:      req.IdCliente,
		IdFranqueado:   idFranqueado,
		NomeCliente:    nomeCliente,
		NomeFranqueado: nomeFranqueado,
		Ordem:          req.Ordem,
		Ativo:          &ativo,
	}

	var out tipos.MapaAmbiente
	if req.Id > 0 {
		out, err = xano.AtualizarMapaAmbiente(req.Id, m)
	} else {
		out, err = xano.CriarMapaAmbiente(m)
	}
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	if len(imagemRaw) > 0 {
		url, err := storage.SubstituirImagemMapa(r.Context(), out.Id, imagemRaw)
		if err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
		m.ImagemUrl = url
		out.ImagemUrl = url
		out, err = xano.AtualizarMapaAmbiente(out.Id, tipos.MapaAmbiente{
			Descricao:      out.Descricao,
			ImagemUrl:      url,
			IdCliente:      out.IdCliente,
			IdFranqueado:   out.IdFranqueado,
			NomeCliente:    out.NomeCliente,
			NomeFranqueado: out.NomeFranqueado,
			Ordem:          out.Ordem,
			Ativo:          out.Ativo,
		})
		if err != nil {
			resposta.Erro(w, http.StatusBadRequest, err)
			return
		}
	} else if !storage.IsURLPlantaMapa(out.Id, imagemUrl) {
		_ = storage.ExcluirPastaMapa(r.Context(), out.Id)
	}

	resposta.JsonDados(w, http.StatusOK, out)
}

func parseSalvarRequest(r *http.Request) (salvarReq, []byte, error) {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		maxUpload := config.ContaboMapaMaxUploadBytes
		if maxUpload <= 0 {
			maxUpload = 10 << 20
		}
		if err := r.ParseMultipartForm(int64(maxUpload) + (1 << 20)); err != nil {
			return salvarReq{}, nil, errors.New("formulario invalido")
		}

		var req salvarReq
		if dados := r.FormValue("dados"); dados != "" {
			if err := json.Unmarshal([]byte(dados), &req); err != nil {
				return salvarReq{}, nil, err
			}
		}

		file, _, err := r.FormFile("arquivo")
		if err != nil {
			return req, nil, nil
		}
		defer file.Close()

		raw, err := io.ReadAll(file)
		if err != nil {
			return salvarReq{}, nil, err
		}
		return req, raw, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return salvarReq{}, nil, err
	}
	var req salvarReq
	if err := json.Unmarshal(body, &req); err != nil {
		return salvarReq{}, nil, err
	}
	return req, nil, nil
}

func excluir(w http.ResponseWriter, r *http.Request) {
	id, err := parseMapaId(r)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}

	_ = storage.ExcluirPastaMapa(r.Context(), id)

	if err := xano.ExcluirMapaAmbiente(id); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JsonOK(w)
}

func parseMapaId(r *http.Request) (int, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}

	var req struct {
		Id              int    `json:"id"`
		MapaAmbienteId  int    `json:"mapa_ambiente_id"`
		MapaAmbienteIdS string `json:"mapa_ambiente_id_str"`
	}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return 0, err
		}
	}

	if req.Id > 0 {
		return req.Id, nil
	}
	if req.MapaAmbienteId > 0 {
		return req.MapaAmbienteId, nil
	}
	if req.MapaAmbienteIdS != "" {
		return strconv.Atoi(req.MapaAmbienteIdS)
	}
	return 0, fmt.Errorf("id do mapa obrigatorio")
}

func buscarNomesMySQL(idCliente, idFranqueado string) (nomeCliente, nomeFranqueado string, err error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return "", "", err
	}
	defer db.Close()

	var nome, nick sql.NullString
	err = db.QueryRow(`
		SELECT cliente.Nome, cliente.Nick
		FROM cliente
		WHERE cliente.ID_Cliente = ? AND cliente.ID_Franqueado = ?
	`, idCliente, idFranqueado).Scan(&nome, &nick)
	if err != nil && err != sql.ErrNoRows {
		return "", "", err
	}
	nomeCliente = nome.String
	if nick.String != "" {
		nomeCliente = nick.String
	}

	var fraNome sql.NullString
	_ = db.QueryRow(`
		SELECT franqueado.RazaoSocial FROM franqueado WHERE franqueado.ID_Franqueado = ?
	`, idFranqueado).Scan(&fraNome)
	nomeFranqueado = fraNome.String

	return nomeCliente, nomeFranqueado, nil
}
