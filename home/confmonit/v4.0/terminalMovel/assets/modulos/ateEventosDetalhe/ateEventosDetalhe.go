package ateEventosDetalhe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	aux "terminal/src/auxiliar"
	"terminal/src/setup"
	"terminal/src/tipos"
	"time"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/ateEventosDetalhe/buscarEventos",
		Metodo: http.MethodPost,
		Funcao: buscarEventos,
	},
	{
		Uri:    "/ateEventosDetalhe/buscarUltimos25",
		Metodo: http.MethodPost,
		Funcao: buscarUltimos25,
	},
	{
		Uri:    "/ateEventosDetalhe/colocaSetorManutecao",
		Metodo: http.MethodPost,
		Funcao: colocaSetorManutecao,
	},
	{
		Uri:    "/ateEventosDetalhe/buscarConfVision",
		Metodo: http.MethodPost,
		Funcao: buscarConfVision,
	},
	{
		Uri:    "/ateEventosDetalhe/buscarConfVisionClips",
		Metodo: http.MethodPost,
		Funcao: buscarConfVisionClips,
	},
	{
		Uri:    "/ateEventosDetalhe/confVisionConfig",
		Metodo: http.MethodGet,
		Funcao: confVisionConfig,
	},
	{
		Uri:    "/ateEventosDetalhe/buscarConfVisionPorSetor",
		Metodo: http.MethodPost,
		Funcao: buscarConfVisionPorSetor,
	},
	{
		Uri:    "/ateEventosDetalhe/buscarConfVisionUltimos25Setor",
		Metodo: http.MethodPost,
		Funcao: buscarConfVisionUltimos25Setor,
	},
	{
		Uri:    "/ateEventosDetalhe/buscarConfVisionUltimos25Dispositivo",
		Metodo: http.MethodPost,
		Funcao: buscarConfVisionUltimos25Dispositivo,
	},
}

func buscarEventos(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso string `json:"idProcesso"`
		Quantidade string `json:"quantidade"`
		Tipo       string `json:"tipo"`
		Codigo     string `json:"codigo"`
		CodigoFull string `json:"codigoFull"`
		Descricao  string `json:"descricao"`
		Particao   string `json:"particao"`
		ZonaUser   string `json:"zonaUser"`
		Hora       string `json:"hora"`
		DescZonUse string `json:"descZonUse"`
		IdDisp     string `json:"idDispositivo"`
		Conta      string `json:"conta"`
		CodBenuvem    string `json:"codBenuvem"`
		CameraOn      string `json:"cameraOn"`
		IdSetor       string `json:"idSetor"`
		SetorManu     string `json:"setorManu"`
		ProvedorVideo string `json:"provedorVideo"`
		UsaConfVision string `json:"usaConfVision"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()
	defer db.Close()

	tab, erro := db.Query(`
		SELECT 
			COUNT(evento.ID_Evento) AS qtd, 
			evento.Codigo,
			evento.Particao,
			evento.ZonaUser,
			MAX(evento.DataEntrada) AS dataEntrada,

			processo.ID_Dispositivo,

			cliente.ID_Franqueado,

			usuariosAlarme.Nome,

			setorAlarme.ID_Setor,
			setorAlarme.Nome,
			setorAlarme.Camera,
			listaManutencao.ID_Alvo,

			dispositivo.Conta,

			franqueado.CodBenuvem,
			dispositivo.ProvedorVideo,
			franqueado.UsaConfVision

		FROM evento

		LEFT JOIN processo
		ON evento.ID_Processo = processo.ID_Processo	

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN usuariosAlarme
		ON processo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo
		AND evento.ZonaUser = usuariosAlarme.codigo 

		LEFT JOIN setorAlarme
		ON processo.ID_Dispositivo = setorAlarme.ID_Dispositivo
		AND evento.ZonaUser = setorAlarme.Numero

		LEFT JOIN listaManutencao
		ON setorAlarme.ID_Setor = listaManutencao.ID_Alvo

		WHERE evento.ID_Processo = ?
		GROUP BY evento.Codigo, evento.ZonaUser
		ORDER BY MAX(evento.DataEntrada) ASC
	`, obj.IdProcesso)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto
	for tab.Next() {
		var (
			item          objeto
			idDispositivo sql.NullString
			idFranq       string
			user          sql.NullString
			setorId       sql.NullString
			setor         sql.NullString
			setorManu     sql.NullString

			conta         sql.NullString
			codBenuvem    sql.NullString
			provedorVideo sql.NullString
			usaConfVision sql.NullString
			camera        sql.NullString
			dataEntrada sql.NullTime
		)

		if erro := tab.Scan(
			&item.Quantidade,
			&item.CodigoFull,
			&item.Particao,
			&item.ZonaUser,
			&dataEntrada,
			&idDispositivo,
			&idFranq,

			&user,

			&setorId,
			&setor,
			&camera,
			&setorManu,

			&conta,
			&codBenuvem,
			&provedorVideo,
			&usaConfVision,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.IdDisp = idDispositivo.String
		item.IdProcesso = obj.IdProcesso
		if dataEntrada.Valid {
			item.Hora = dataEntrada.Time.Format("15:04:05")
		}
		item.Conta = conta.String
		item.CodBenuvem = codBenuvem.String
		item.ProvedorVideo = strings.ToLower(strings.TrimSpace(provedorVideo.String))
		item.UsaConfVision = usaConfVision.String
		if string(item.CodigoFull[0]) == "1" {
			item.Tipo = "E"
		} else if string(item.CodigoFull[0]) == "3" {
			item.Tipo = "R"
		}

		item.Codigo = item.CodigoFull[1:]

		var grupo string
		if err := aux.GetCtiGrupoAndDescricao(idFranq, item.CodigoFull, &grupo, &item.Descricao); err != nil {
			aux.RespostaErro(w, http.StatusBadRequest, err)
			return
		}

		if setorManu.Valid {
			item.SetorManu = "N"
		} else {
			item.SetorManu = "S"
		}

		/*
			Descricao  string `json:"descricao"`
			DescZonUse string `json:"descZonUse"`

		*/
		item.CameraOn = "N"
		if grupo == "ARME" || grupo == "DESARME" || grupo == "PANICO" {
			if user.Valid {
				item.DescZonUse = user.String
			} else {
				item.DescZonUse = "NÃO CADASTRADO"
			}

			item.IdSetor = "USER"

		} else {
			if setorId.Valid {
				item.IdSetor = setorId.String
			} else {
				item.IdSetor = "USER"
			}

			if setor.Valid {
				item.DescZonUse = setor.String
			} else {
				item.DescZonUse = "NÃO CADASTRADO"
			}
			if camera.Valid {
				item.CameraOn = camera.String
			}
		}

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}

func buscarUltimos25(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso       string `json:"idProcesso"`
		IdEvento         string `json:"idEvento"`
		Img              string `json:"img"`
		Tipo             string `json:"tipo"`
		Codigo           string `json:"codigo"`
		Descricao        string `json:"descricao"`
		Particao         string `json:"particao"`
		ZonaUser         string `json:"zonaUser"`
		DescZonUse       string `json:"descZonUse"`
		Entrada          string `json:"entrada"`
		IdDisp           string `json:"idDispositivo"`
		DescrAtendimento string `json:"descrAtendimento"`
		CameraOn         string `json:"cameraOn"`
		Conta            string `json:"conta"`
		CodBenuvem       string `json:"codBenuvem"`
		ProvedorVideo    string `json:"provedorVideo"`
		UsaConfVision    string `json:"usaConfVision"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
	SELECT 
		sub.Codigo,
		sub.Particao,
		sub.ZonaUser,
		sub.DataEntrada,
		sub.ID_Dispositivo,
		sub.ID_Processo,
		sub.ID_Evento,
		sub.Img,
		sub.Descricao,
		sub.ID_Franqueado,
		usuariosAlarme.Nome,
		setorAlarme.Nome,
		setorAlarme.Camera,
		dispositivo.Conta,
		franqueado.CodBenuvem,
		dispositivo.ProvedorVideo,
		franqueado.UsaConfVision
	FROM (
		SELECT 
			evento.Codigo,
			evento.Particao,
			evento.ZonaUser,
			evento.DataEntrada,
			evento.ID_Evento,
			evento.Img,
			processo.ID_Dispositivo,
			processo.ID_Processo,
			processo.Descricao,
			cliente.ID_Franqueado
		FROM dispositivo
		STRAIGHT_JOIN processo
			ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
			AND processo.DataAtenFim IS NOT NULL
		STRAIGHT_JOIN cliente
			ON cliente.ID_Cliente = dispositivo.ID_Cliente
		STRAIGHT_JOIN evento
			ON evento.ID_Processo = processo.ID_Processo
		WHERE dispositivo.ID_Dispositivo = ?
		ORDER BY evento.DataEntrada DESC
		LIMIT 25
	) AS sub
	LEFT JOIN dispositivo
		ON dispositivo.ID_Dispositivo = sub.ID_Dispositivo
	LEFT JOIN cliente
		ON cliente.ID_Cliente = dispositivo.ID_Cliente
	LEFT JOIN franqueado
		ON franqueado.ID_Franqueado = cliente.ID_Franqueado
	LEFT JOIN usuariosAlarme
		ON usuariosAlarme.ID_Dispositivo = sub.ID_Dispositivo
		AND usuariosAlarme.Codigo = sub.ZonaUser
	LEFT JOIN setorAlarme
		ON setorAlarme.ID_Dispositivo = sub.ID_Dispositivo
		AND setorAlarme.Numero = sub.ZonaUser
		AND setorAlarme.Particao = sub.Particao
	ORDER BY sub.DataEntrada DESC
	`, obj.IdDisp)

	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var lista []objeto
	for tab.Next() {
		var (
			item                 objeto
			idDispositivo        sql.NullString
			idProcesso           sql.NullString
			idEvento             sql.NullString
			img                  sql.NullString
			descricaoAtendimento sql.NullString
			idFranq              string
			usuarioAlarme        sql.NullString
			setorAlarme          sql.NullString
			cameraOn             sql.NullString
			conta                sql.NullString
			codBenuvem           sql.NullString
			provedorVideo        sql.NullString
			usaConfVision        sql.NullString
			entrada              sql.NullTime
		)

		if erro := tab.Scan(
			&item.Codigo,
			&item.Particao,
			&item.ZonaUser,
			&entrada,

			&idDispositivo,
			&idProcesso,
			&idEvento,
			&img,
			&descricaoAtendimento,

			&idFranq,
			&usuarioAlarme,
			&setorAlarme,
			&cameraOn,
			&conta,
			&codBenuvem,
			&provedorVideo,
			&usaConfVision,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		item.IdDisp = idDispositivo.String
		item.IdProcesso = idProcesso.String
		item.IdEvento = idEvento.String
		item.Img = img.String
		item.DescrAtendimento = descricaoAtendimento.String
		item.Conta = conta.String
		item.CodBenuvem = codBenuvem.String
		item.ProvedorVideo = strings.ToLower(strings.TrimSpace(provedorVideo.String))
		item.UsaConfVision = usaConfVision.String
		item.CameraOn = "N"
		if cameraOn.Valid {
			item.CameraOn = cameraOn.String
		}
		item.Entrada = entrada.Time.Format("02/01/2006 15:04:05")
		if string(item.Codigo[0]) == "1" {
			item.Tipo = "E"
		} else if string(item.Codigo[0]) == "3" {
			item.Tipo = "R"
		}

		//ADICIONADO POR ADRIANO BRAZ
		var grupo string
		if err := aux.GetCtiGrupoAndDescricaoDB(db, idFranq, item.Codigo, &grupo, &item.Descricao); err != nil {
			aux.RespostaErro(w, http.StatusBadRequest, err)
			return
		}

		//		var grupo string
		//		if err := aux.GetCtiGrupoAndDescricao(idFranq, item.Codigo, &grupo, &item.Descricao); err != nil {
		//			aux.RespostaErro(w, http.StatusBadRequest, erro)
		//			return
		//		}
		//fmt.Println(item.ZonaUser, grupo, usuarioAlarme.String, setorAlarme.String)
		if grupo == "ARME" || grupo == "DESARME" || grupo == "PANICO" {
			if usuarioAlarme.Valid {
				item.DescZonUse = usuarioAlarme.String
			} else {
				item.DescZonUse = "NÃO CADASTRADO"
			}
		} else {
			if setorAlarme.Valid {
				item.DescZonUse = setorAlarme.String
			} else {
				item.DescZonUse = "NÃO CADASTRADO"
			}
		}

		lista = append(lista, item)
	}

	if len(lista) > 0 {
		aux.RespostaJsonDados(w, http.StatusOK, lista)
	} else {
		aux.RespostaJsonVazio(w)
	}
}

func colocaSetorManutecao(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdSetor  string `json:"idSetor"`
		Operador string `json:"operador"`
	}
	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	stm, err := db.Prepare(`
		INSERT INTO listaManutencao(
			listaManutencao.ID_Alvo, 
			listaManutencao.DataBloqueio, 
			listaManutencao.DataRetirada, 
			listaManutencao.Descricao
		) VALUES ( ?, ?, ?, ? )
	`)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	defer stm.Close()

	if _, err := stm.Exec(
		obj.IdSetor,
		time.Now().Format("2006-01-02 15:04:05"),
		"2037-12-31 23:59:59",
		fmt.Sprintf("ADICIONADO PELO OPERADOR %s", strings.ToUpper(obj.Operador)),
	); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	aux.RespostaJsonOK(w)
}

func confVisionConfig(w http.ResponseWriter, r *http.Request) {
	type cfg struct {
		HlsBase string `json:"hlsBase"`
	}
	aux.RespostaJsonDados(w, http.StatusOK, cfg{HlsBase: strings.TrimRight(setup.ConfVisionHlsBase, "/")})
}

func buscarConfVision(w http.ResponseWriter, r *http.Request) {
	type entrada struct {
		IdProcesso    string `json:"idProcesso"`
		IdCliente     string `json:"idCliente"`
		IdDispositivo string `json:"idDispositivo"`
		IdFranqueado  string `json:"idFranqueado"`
		DataInicio    string `json:"dataInicio"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj entrada
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if setup.XanoBaseUrl == "" {
		aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{"dados": []interface{}{}})
		return
	}

	params := fmt.Sprintf(
		"/vis_evento_by_contexto?id_processo=%s&id_cliente=%s&id_dispositivo=%s&id_franqueado=%s",
		urlQuery(obj.IdProcesso),
		urlQuery(obj.IdCliente),
		urlQuery(obj.IdDispositivo),
		urlQuery(obj.IdFranqueado),
	)
	if obj.DataInicio != "" {
		params += "&data_inicio=" + urlQuery(obj.DataInicio)
	}

	dados, err := xanoGet(params)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(dados, &parsed); err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, parsed)
}

func buscarConfVisionPorSetor(w http.ResponseWriter, r *http.Request) {
	type entrada struct {
		IdProcesso    string `json:"idProcesso"`
		IdDispositivo string `json:"idDispositivo"`
		Particao      string `json:"particao"`
		ZonaUser      string `json:"zonaUser"`
		IdEvento      string `json:"idEvento"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj entrada
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if setup.XanoBaseUrl == "" {
		aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{"dados": nil, "camera": nil})
		return
	}

	params := fmt.Sprintf(
		"/vis_evento_by_processo_setor?id_processo=%s&id_dispositivo=%s&particao=%s&zonauser=%s&id_evento=%s",
		urlQuery(obj.IdProcesso),
		urlQuery(obj.IdDispositivo),
		urlQuery(obj.Particao),
		urlQuery(obj.ZonaUser),
		urlQuery(obj.IdEvento),
	)

	dados, err := xanoGet(params)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(dados, &parsed); err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, parsed)
}

func buscarConfVisionUltimos25Setor(w http.ResponseWriter, r *http.Request) {
	type entrada struct {
		IdDispositivo string `json:"idDispositivo"`
		Particao      string `json:"particao"`
		ZonaUser      string `json:"zonaUser"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj entrada
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if setup.XanoBaseUrl == "" {
		aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{"dados": []interface{}{}, "camera": nil})
		return
	}

	params := fmt.Sprintf(
		"/vis_evento_ultimos25_setor?id_dispositivo=%s&particao=%s&zonauser=%s",
		urlQuery(obj.IdDispositivo),
		urlQuery(obj.Particao),
		urlQuery(obj.ZonaUser),
	)

	dados, err := xanoGet(params)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(dados, &parsed); err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, parsed)
}

func buscarConfVisionUltimos25Dispositivo(w http.ResponseWriter, r *http.Request) {
	type entrada struct {
		IdDispositivo string `json:"idDispositivo"`
		Offset        int    `json:"offset"`
		Limit         int    `json:"limit"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj entrada
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if setup.XanoBaseUrl == "" {
		aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{
			"dados":  []interface{}{},
			"total":  0,
			"offset": 0,
			"limit":  25,
		})
		return
	}

	limit := obj.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	offset := obj.Offset
	if offset < 0 {
		offset = 0
	}

	params := fmt.Sprintf(
		"/vis_evento_ultimos25_dispositivo?id_dispositivo=%s&offset=%d&limit=%d",
		urlQuery(obj.IdDispositivo),
		offset,
		limit,
	)

	dados, err := xanoGet(params)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(dados, &parsed); err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, parsed)
}

func buscarConfVisionClips(w http.ResponseWriter, r *http.Request) {
	type entrada struct {
		EventoId int `json:"eventoId"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj entrada
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if setup.XanoBaseUrl == "" || obj.EventoId <= 0 {
		aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{"evento": nil, "clips": []interface{}{}})
		return
	}

	path := fmt.Sprintf("/vis_evento/%d/clips?id=%d", obj.EventoId, obj.EventoId)
	dados, err := xanoGet(path)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(dados, &parsed); err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, parsed)
}

func xanoGet(path string) ([]byte, error) {
	reqURL := strings.TrimRight(setup.XanoBaseUrl, "/") + path
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+setup.XanoToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("xano %s: %s", resp.Status, string(body))
	}
	return body, nil
}

func urlQuery(v string) string {
	return url.QueryEscape(v)
}
