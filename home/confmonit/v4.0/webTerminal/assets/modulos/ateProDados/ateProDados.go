package ateProDados

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	aux "terminal/src/auxiliar"
	"terminal/src/setup"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/ateProDados/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
	{
		Uri:    "/ateProDados/aguardar",
		Metodo: http.MethodPost,
		Funcao: aguardar,
	},
	{
		Uri:    "/ateProDados/finalizar",
		Metodo: http.MethodPost,
		Funcao: finalizar,
	},
	{
		Uri:    "/ateProDados/eventoProcesso",
		Metodo: http.MethodPost,
		Funcao: eventoProcesso,
	},
	{
		Uri:    "/ateProDados/sugerirDescricao",
		Metodo: http.MethodPost,
		Funcao: sugerirDescricao,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso      string `json:"idProcesso"`
		Descricao       string `json:"descricao"`
		IdFranqueado    string `json:"idFranqueado"`
		FranqNome       string `json:"franqNome"`
		FranqCodBenuvem string `json:"franqCodBenuvem"`
		IdCliente       string `json:"idCliente"`
		CliNome         string `json:"cliNome"`
		IdDispositivo   string `json:"idDispositivo"`
		DispNome        string `json:"dispNome"`
		DispConta       string `json:"dispConta"`
		DispArmado      string `json:"dispArmado"`
		DispParticao    string `json:"dispParticao"`
		DispMsgAtend    string `json:"dispMsgAtend"`
		DispSenha       string `json:"dispSenha"`
		UsaConfVision   string `json:"usaConfVision"`
		ProvedorVideo   string `json:"provedorVideo"`
		DataCriacao     string `json:"dataCriacao"`
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
			processo.ID_Processo,
			processo.Descricao,

			franqueado.ID_Franqueado,
			franqueado.NomeFantasia,
			franqueado.RazaoSocial,
			franqueado.CodBenuvem,

			cliente.ID_Cliente,
			cliente.Nome,

			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			dispositivo.Conta,
			dispositivo.Armado,
			dispositivo.Particao,
			dispositivo.MsgAtendente,
			dispositivo.Senha,
			franqueado.UsaConfVision,
			dispositivo.ProvedorVideo,
			processo.DataCriacao

		FROM processo

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		WHERE processo.ID_Processo = ?
	`, obj.IdProcesso)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	for tab.Next() {
		var (
			idProcesso       sql.NullString
			descricao        sql.NullString
			idFranqueado     sql.NullString
			franqNomeFantasia sql.NullString
			franqRazaoSocial  sql.NullString
			franqCodBenuvem  sql.NullString
			idCliente        sql.NullString
			cliNome          sql.NullString
			idDispositivo    sql.NullString
			dispNome         sql.NullString
			dispConta        sql.NullString
			dispArmado       sql.NullString
			dispParticao     sql.NullString
			dispMsgAtend     sql.NullString
			dispSenha        sql.NullString
			usaConfVision    sql.NullString
			provedorVideo    sql.NullString
			dataCriacao      sql.NullTime
		)

		if erro := tab.Scan(
			&idProcesso,
			&descricao,

			&idFranqueado,
			&franqNomeFantasia,
			&franqRazaoSocial,
			&franqCodBenuvem,

			&idCliente,
			&cliNome,

			&idDispositivo,
			&dispNome,
			&dispConta,
			&dispArmado,
			&dispParticao,
			&dispMsgAtend,
			&dispSenha,
			&usaConfVision,
			&provedorVideo,
			&dataCriacao,
		); erro != nil {
			fmt.Println(erro)
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		obj.IdProcesso = idProcesso.String
		obj.Descricao = descricao.String

		obj.IdFranqueado = idFranqueado.String
		// Empresa: exibir nome curto (Nome Fantasia quando existir, senão Razão Social)
		fantasia := strings.TrimSpace(franqNomeFantasia.String)
		razao := strings.TrimSpace(franqRazaoSocial.String)
		if fantasia != "" {
			obj.FranqNome = fantasia
		} else {
			obj.FranqNome = razao
		}
		obj.FranqCodBenuvem = franqCodBenuvem.String

		obj.IdCliente = idCliente.String
		obj.CliNome = cliNome.String

		obj.IdDispositivo = idDispositivo.String
		obj.DispNome = dispNome.String
		obj.DispConta = dispConta.String
		obj.DispArmado = dispArmado.String
		obj.DispParticao = dispParticao.String
		obj.DispMsgAtend = dispMsgAtend.String
		obj.DispSenha = dispSenha.String
		obj.UsaConfVision = usaConfVision.String
		obj.ProvedorVideo = strings.ToLower(strings.TrimSpace(provedorVideo.String))
		if dataCriacao.Valid {
			obj.DataCriacao = dataCriacao.Time.Format("2006-01-02T15:04:05Z07:00")
		}

	}

	aux.RespostaJsonDados(w, http.StatusOK, obj)
}

func aguardar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso string `json:"idProcesso"`
		Descricao  string `json:"descricao"`
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

	stm, erro := db.Prepare(`
		UPDATE processo
		SET 
			processo.Descricao = ?,
			processo.ID_Atendente = "0"
		WHERE processo.ID_Processo = ?
	`)

	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		strings.ToUpper(obj.Descricao),
		obj.IdProcesso,
	); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, obj)
}

func finalizar(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdProcesso   string `json:"idProcesso"`
		IdCliente    string `json:"idCliente"`
		Descricao    string `json:"descricao"`
		IdOperador   string `json:"idOperador"`
		NomeOperador string `json:"nomeOperador"`
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

	stm, erro := db.Prepare(`
		UPDATE processo
		SET 
			processo.Descricao = ?,
			processo.ID_Atendente = ?,
			processo.DataAtenFim = ?

		WHERE processo.ID_Processo = ?
	`)

	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer stm.Close()

	if _, erro := stm.Exec(
		strings.ToUpper(obj.Descricao),
		obj.IdOperador,
		time.Now().Format("2006-01-02 15:04:05"),
		obj.IdProcesso,
	); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if err := gravaAtendimento(db, obj.IdCliente, obj.NomeOperador); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, obj)
}

func eventoProcesso(w http.ResponseWriter, r *http.Request) {

	type objSetor struct {
		IdEvento string `json:"idEvento"`
		Img      string `json:"img"`
	}

	type objeto struct {
		IdProcesso string     `json:"idProcesso"`
		Setor      []objSetor `json:"setor"`
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
			evento.ID_Evento,
			evento.Img 
		FROM evento 
		WHERE evento.ID_Processo = ?
	`, obj.IdProcesso)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	for tab.Next() {
		var item objSetor

		if err := tab.Scan(&item.IdEvento, &item.Img); err != nil {
			aux.RespostaErro(w, http.StatusBadRequest, err)
			return
		}

		obj.Setor = append(obj.Setor, item)

	}

	aux.RespostaJsonDados(w, http.StatusOK, obj)
}

func sugerirDescricao(w http.ResponseWriter, r *http.Request) {
	type req struct {
		IdProcesso string `json:"idProcesso"`
		Rascunho   string `json:"rascunho"`
	}
	type resp struct {
		TextoSugerido string `json:"textoSugerido"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	var obj req
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	if setup.OpenAIApiKey == "" {
		aux.RespostaErro(w, http.StatusServiceUnavailable, fmt.Errorf("OPENAI_API_KEY não configurada"))
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	var cliNome, dispNome, idFranqueado string
	row := db.QueryRow(`
		SELECT cliente.Nome, dispositivo.Nome, cliente.ID_Franqueado
		FROM processo
		LEFT JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
		LEFT JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
		WHERE processo.ID_Processo = ?
	`, obj.IdProcesso)
	if err := row.Scan(&cliNome, &dispNome, &idFranqueado); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	tab, erro := db.Query(`
		SELECT evento.Codigo, evento.Particao, evento.ZonaUser, evento.DataEntrada
		FROM evento
		WHERE evento.ID_Processo = ?
		ORDER BY evento.DataEntrada ASC
	`, obj.IdProcesso)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	var linhasEventos []string
	for tab.Next() {
		var codigo, particao, zonaUser string
		var dataEntrada sql.NullTime
		if err := tab.Scan(&codigo, &particao, &zonaUser, &dataEntrada); err != nil {
			continue
		}
		var grupo, descricao string
		if err := aux.GetCtiGrupoAndDescricao(idFranqueado, codigo, &grupo, &descricao); err != nil {
			descricao = codigo
		}
		dataStr := ""
		if dataEntrada.Valid {
			dataStr = dataEntrada.Time.Format("02/01/2006 15:04:05")
		}
		linhasEventos = append(linhasEventos, fmt.Sprintf("- %s | Partição %s Zona %s | %s | %s", dataStr, particao, zonaUser, descricao, grupo))
	}

	// Monta o texto de tipo de evento a partir dos eventos do processo
	tipoEvento := strings.Join(linhasEventos, "; ")
	if strings.TrimSpace(tipoEvento) == "" {
		tipoEvento = fmt.Sprintf("EVENTOS DO CLIENTE %s / DISPOSITIVO %s NÃO INFORMADOS", cliNome, dispNome)
	}

	descricaoOperador := strings.TrimSpace(obj.Rascunho)
	if descricaoOperador == "" {
		descricaoOperador = "SEM COMPLEMENTO DO OPERADOR"
	}

	// Prompt: registro de atendimento com evento e obs
	prompt := "" +
		"Reescreva como registro de atendimento de central de alarme.\n\n" +
		"Melhore a observação do operador usando o evento como contexto.\n" +
		"Não invente informações.\n" +
		"Não incluir zona, partição, data, hora ou códigos técnicos.\n" +
		"Máx 96 caracteres. Até 2 frases curtas.\n\n" +
		"Evento: " + tipoEvento + "\nObs: " + descricaoOperador + "\n\n" +
		"Responda apenas com o registro final, sem explicações."

	// Chamada HTTP para /v1/responses (modelo gpt-5.4)
	model := setup.OpenAIModel
	if strings.TrimSpace(model) == "" {
		model = "gpt-5.4"
	}
	// Payload mínimo como no curl: só model e input
	payload := map[string]interface{}{
		"model": model,
		"input": prompt,
	}
	bodyReq, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/responses", bytes.NewReader(bodyReq))
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+setup.OpenAIApiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadGateway, err)
		return
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		aux.RespostaErro(w, httpResp.StatusCode, fmt.Errorf("OpenAI API: %s", string(respBody)))
		return
	}

	// Parse: output[] -> type "message" -> content[] -> type "output_text" -> text
	var openAIResp struct {
		Output []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("resposta OpenAI inválida: %v", err))
		return
	}
	var texto string
	for _, out := range openAIResp.Output {
		if out.Type != "message" {
			continue
		}
		for _, c := range out.Content {
			if c.Type == "output_text" && strings.TrimSpace(c.Text) != "" {
				texto = c.Text
				break
			}
		}
		if texto != "" {
			break
		}
	}
	texto = strings.TrimSpace(texto)
	if texto == "" {
		// Inclui o corpo da resposta para debug (truncado se muito grande)
		bodyDebug := string(respBody)
		if len(bodyDebug) > 800 {
			bodyDebug = bodyDebug[:800] + "..."
		}
		aux.RespostaErro(w, http.StatusBadRequest, fmt.Errorf("resposta OpenAI vazia: %s", bodyDebug))
		return
	}

	aux.RespostaJsonDados(w, http.StatusOK, resp{TextoSugerido: texto})
}

func gravaAtendimento(db *sql.DB, idVinculo, atendente string) error {
	stm, err := db.Prepare(`
		INSERT INTO tarifacao(
			ID_Tarifacao,
			ID_Vinculo,
			TipoOperacao,
			DadoOperacao,
			Aux
		) VALUES (?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		aux.GeradorDeId(),
		idVinculo,
		"ATENDIMENTO",
		atendente,
		"",
	); err != nil {
		return err
	}
	return nil
}
