package interno

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"terminal/assets/modulos/device"
	"terminal/src/push"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/interno/push/processo",
		Metodo: http.MethodPost,
		Funcao: pushProcesso,
	},
	{
		Uri:    "/interno/push/teste",
		Metodo: http.MethodPost,
		Funcao: pushTeste,
	},
}

func pushProcesso(w http.ResponseWriter, r *http.Request) {
	if !authorizeInterno(r) {
		aux.RespostaErro(w, http.StatusUnauthorized, errors.New("não autorizado"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IdProcesso    string `json:"idProcesso"`
		IdFranqueado  string `json:"idFranqueado"`
		NomeCliente   string `json:"nomeCliente"`
		NomeFranqueado string `json:"nomeFranqueado"`
		Nivel         string `json:"nivel"`
		Codigo        string `json:"codigo"`
		Descricao     string `json:"descricao"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.IdFranqueado) == "" {
		aux.RespostaErro(w, http.StatusBadRequest, errors.New("idFranqueado obrigatório"))
		return
	}

	_ = device.EnsureTable()

	tokens, err := tokensParaFranqueado(req.IdFranqueado)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	title := "Novo alarme"
	if req.NomeCliente != "" {
		title = req.NomeCliente
	}
	bodyTxt := "Processo na fila do Terminal Móvel"
	if req.Codigo != "" || req.Descricao != "" {
		bodyTxt = strings.TrimSpace(req.Codigo + " " + req.Descricao)
	} else if req.NomeFranqueado != "" {
		bodyTxt = req.NomeFranqueado
	}

	go push.SendToTokens(tokens, push.Message{
		Title: title,
		Body:  bodyTxt,
		Data: map[string]string{
			"tipo":         "novo_processo",
			"idProcesso":   req.IdProcesso,
			"idFranqueado": req.IdFranqueado,
			"nivel":        req.Nivel,
		},
	})

	aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{
		"enviados": len(tokens),
	})
}

func pushTeste(w http.ResponseWriter, r *http.Request) {
	if !authorizeInterno(r) {
		aux.RespostaErro(w, http.StatusUnauthorized, errors.New("não autorizado"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		IdOperador string `json:"idOperador"`
		Token      string `json:"token"`
	}
	_ = json.Unmarshal(body, &req)

	var tokens []string
	if strings.TrimSpace(req.Token) != "" {
		tokens = []string{req.Token}
	} else if strings.TrimSpace(req.IdOperador) != "" {
		db, err := aux.Conectar()
		if err != nil {
			aux.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
		defer db.Close()
		rows, err := db.Query(`
			SELECT fcm_token FROM operador_fcm_token
			WHERE id_usuario=? AND ativo='S'
		`, req.IdOperador)
		if err != nil {
			aux.RespostaErro(w, http.StatusBadRequest, err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var t string
			if err := rows.Scan(&t); err == nil {
				tokens = append(tokens, t)
			}
		}
	}

	go push.SendToTokens(tokens, push.Message{
		Title: "Terminal Móvel",
		Body:  "Notificação de teste",
		Data:  map[string]string{"tipo": "teste"},
	})
	aux.RespostaJsonDados(w, http.StatusOK, map[string]interface{}{"enviados": len(tokens)})
}

func authorizeInterno(r *http.Request) bool {
	secret := strings.TrimSpace(os.Getenv("TERMINAL_PUSH_SECRET"))
	if secret == "" {
		return true
	}
	got := strings.TrimSpace(r.Header.Get("X-Terminal-Push-Secret"))
	if got == "" {
		got = strings.TrimSpace(r.URL.Query().Get("secret"))
	}
	return got == secret
}

func tokensParaFranqueado(idFranqueado string) ([]string, error) {
	db, err := aux.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Operadores do franqueado + todos da CENTRAL (user_tipo CEN ou vínculo em tabela central)
	rows, err := db.Query(`
		SELECT DISTINCT t.fcm_token
		FROM operador_fcm_token t
		INNER JOIN usuarios u ON u.ID_Usuario = t.id_usuario
		WHERE t.ativo = 'S'
		AND u.UsuarioTeminal = 'S'
		AND (
			t.id_vinculo = ?
			OR t.user_tipo = 'CEN'
			OR EXISTS (
				SELECT 1 FROM central c WHERE c.ID_Central = t.id_vinculo
			)
		)
	`, idFranqueado)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var tok string
		if err := rows.Scan(&tok); err != nil {
			continue
		}
		out = append(out, tok)
	}
	return out, nil
}
