package device

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
	"time"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/device/register",
		Metodo: http.MethodPost,
		Funcao: register,
	},
	{
		Uri:    "/device/unregister",
		Metodo: http.MethodPost,
		Funcao: unregister,
	},
	{
		Uri:    "/device/diag",
		Metodo: http.MethodPost,
		Funcao: diag,
	},
}

func EnsureTable() error {
	db, err := aux.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS operador_fcm_token (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			id_usuario VARCHAR(64) NOT NULL,
			id_vinculo VARCHAR(64) NOT NULL DEFAULT '',
			user_tipo VARCHAR(8) NOT NULL DEFAULT '',
			fcm_token VARCHAR(512) NOT NULL,
			platform VARCHAR(16) NOT NULL DEFAULT 'android',
			ativo CHAR(1) NOT NULL DEFAULT 'S',
			updated_at DATETIME NOT NULL,
			UNIQUE KEY uk_fcm_token (fcm_token),
			KEY idx_usuario (id_usuario),
			KEY idx_vinculo (id_vinculo)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS app_diag_log (
			id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
			id_usuario VARCHAR(64) NOT NULL DEFAULT '',
			id_vinculo VARCHAR(64) NOT NULL DEFAULT '',
			nivel VARCHAR(16) NOT NULL DEFAULT 'info',
			categoria VARCHAR(64) NOT NULL DEFAULT '',
			mensagem TEXT NOT NULL,
			detalhe TEXT NULL,
			app_version VARCHAR(32) NOT NULL DEFAULT '',
			platform VARCHAR(16) NOT NULL DEFAULT '',
			firebase_project VARCHAR(128) NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL,
			KEY idx_created (created_at),
			KEY idx_usuario (id_usuario),
			KEY idx_categoria (categoria)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`)
	return err
}

func register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IdOperador string `json:"idOperador"`
		IdVinculo  string `json:"idVinculo"`
		UserTipo   string `json:"userTipo"`
		Token      string `json:"token"`
		Platform   string `json:"platform"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	req.IdOperador = strings.TrimSpace(req.IdOperador)
	if req.Token == "" || req.IdOperador == "" {
		aux.RespostaErro(w, http.StatusBadRequest, errors.New("token e idOperador obrigatórios"))
		return
	}
	if req.Platform == "" {
		req.Platform = "android"
	}
	if req.UserTipo == "" {
		req.UserTipo = "FRA"
	}

	if err := EnsureTable(); err != nil {
		aux.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}

	db, err := aux.Conectar()
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	defer db.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec(`
		INSERT INTO operador_fcm_token
			(id_usuario, id_vinculo, user_tipo, fcm_token, platform, ativo, updated_at)
		VALUES (?, ?, ?, ?, ?, 'S', ?)
		ON DUPLICATE KEY UPDATE
			id_usuario = VALUES(id_usuario),
			id_vinculo = VALUES(id_vinculo),
			user_tipo = VALUES(user_tipo),
			platform = VALUES(platform),
			ativo = 'S',
			updated_at = VALUES(updated_at)
	`, req.IdOperador, req.IdVinculo, req.UserTipo, req.Token, req.Platform, now)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	log.Printf("[device/register] OK id=%s platform=%s tokenLen=%d", req.IdOperador, req.Platform, len(req.Token))
	aux.RespostaJsonOK(w)
}

func unregister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		Token      string `json:"token"`
		IdOperador string `json:"idOperador"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	db, err := aux.Conectar()
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	defer db.Close()

	if strings.TrimSpace(req.Token) != "" {
		_, _ = db.Exec(`UPDATE operador_fcm_token SET ativo='N', updated_at=NOW() WHERE fcm_token=?`, req.Token)
	} else if strings.TrimSpace(req.IdOperador) != "" {
		_, _ = db.Exec(`UPDATE operador_fcm_token SET ativo='N', updated_at=NOW() WHERE id_usuario=?`, req.IdOperador)
	}

	aux.RespostaJsonOK(w)
}

func diag(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		IdOperador      string `json:"idOperador"`
		IdVinculo       string `json:"idVinculo"`
		Nivel           string `json:"nivel"`
		Categoria       string `json:"categoria"`
		Mensagem        string `json:"mensagem"`
		Detalhe         string `json:"detalhe"`
		AppVersion      string `json:"appVersion"`
		Platform        string `json:"platform"`
		FirebaseProject string `json:"firebaseProject"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	if strings.TrimSpace(req.Mensagem) == "" {
		aux.RespostaErro(w, http.StatusBadRequest, errors.New("mensagem obrigatória"))
		return
	}
	if req.Nivel == "" {
		req.Nivel = "info"
	}
	if req.Categoria == "" {
		req.Categoria = "app"
	}

	if err := EnsureTable(); err != nil {
		aux.RespostaErro(w, http.StatusInternalServerError, err)
		return
	}

	db, err := aux.Conectar()
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}
	defer db.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec(`
		INSERT INTO app_diag_log
			(id_usuario, id_vinculo, nivel, categoria, mensagem, detalhe, app_version, platform, firebase_project, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, req.IdOperador, req.IdVinculo, req.Nivel, req.Categoria, req.Mensagem, req.Detalhe,
		req.AppVersion, req.Platform, req.FirebaseProject, now)
	if err != nil {
		aux.RespostaErro(w, http.StatusBadRequest, err)
		return
	}

	log.Printf("[device/diag] %s/%s user=%s | %s | %s",
		req.Nivel, req.Categoria, req.IdOperador, req.Mensagem, truncate(req.Detalhe, 200))
	aux.RespostaJsonOK(w)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
