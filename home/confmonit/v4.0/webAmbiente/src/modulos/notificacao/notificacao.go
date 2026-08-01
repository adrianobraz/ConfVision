package notificacao

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"webAmbiente/src/auxiliar"
	"webAmbiente/src/resposta"
	"webAmbiente/src/seguranca"
	"webAmbiente/src/tipos"
)

var Rotas = []tipos.Rota{
	{Uri: "/notificacoes/minhas", Metodo: http.MethodPost, Controle: minhas, Seguro: true},
	{Uri: "/notificacoes/contagem", Metodo: http.MethodPost, Controle: contagem, Seguro: true},
	{Uri: "/notificacoes/marcar-visto", Metodo: http.MethodPost, Controle: marcarVisto, Seguro: true},
	{Uri: "/notificacoes/marcar-lido", Metodo: http.MethodPost, Controle: marcarLido, Seguro: true},
}

type itemNotif struct {
	IDNotificacao string `json:"idNotificacao"`
	Titulo        string `json:"titulo"`
	Corpo         string `json:"corpo"`
	Tipo          string `json:"tipo"`
	Softwares     string `json:"softwares"`
	Publico       string `json:"publico"`
	IDFranqueado  string `json:"idFranqueado"`
	IDCliente     string `json:"idCliente"`
	Ativo         string `json:"ativo"`
	ValidoAte     string `json:"validoAte"`
	CriadoPor     string `json:"criadoPor"`
	DataCadastro  string `json:"dataCadastro"`
	VistoEm       string `json:"vistoEm,omitempty"`
	LidoEm        string `json:"lidoEm,omitempty"`
}

func ctxOperador(r *http.Request) (software, idUsuario, userTipo, userMaster, idFranqueado, idCliente string, err error) {
	op, err := seguranca.LerOperador(r)
	if err != nil {
		return
	}
	software = "webambiente"
	idUsuario = strings.TrimSpace(op["idOperador"])
	userTipo = strings.ToUpper(strings.TrimSpace(op["userTipo"]))
	userMaster = strings.ToUpper(strings.TrimSpace(op["userMaster"]))
	if userMaster != "S" {
		userMaster = "N"
	}
	idFranqueado = strings.TrimSpace(op["idVinculo"])
	if userTipo != "FRA" {
		// CEN/REP: idVinculo nao e franqueado; avisos "franqueado" so batem se filtrados pelo tipo
		idFranqueado = ""
	}
	idCliente = ""
	if idUsuario == "" {
		err = errors.New("sessao invalida")
	}
	return
}

func minhas(w http.ResponseWriter, r *http.Request) {
	software, idUsuario, userTipo, userMaster, idFranqueado, idCliente, err := ctxOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	lista, err := listarMinhas(software, idUsuario, userTipo, userMaster, idFranqueado, idCliente)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	if len(lista) == 0 {
		resposta.JSON(w, http.StatusOK, map[string]string{"status": "Vazio"})
		return
	}
	resposta.JSON(w, http.StatusOK, map[string]interface{}{"status": "OK", "dados": lista})
}

func contagem(w http.ResponseWriter, r *http.Request) {
	software, idUsuario, userTipo, userMaster, idFranqueado, idCliente, err := ctxOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	total, err := contarNaoVistas(software, idUsuario, userTipo, userMaster, idFranqueado, idCliente)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "OK",
		"dados":  map[string]int{"total": total},
	})
}

func marcarVisto(w http.ResponseWriter, r *http.Request) {
	marcar(w, r, true, false)
}

func marcarLido(w http.ResponseWriter, r *http.Request) {
	marcar(w, r, true, true)
}

func marcar(w http.ResponseWriter, r *http.Request, visto, lido bool) {
	software, idUsuario, userTipo, _, idFranqueado, idCliente, err := ctxOperador(r)
	if err != nil {
		resposta.Erro(w, http.StatusUnauthorized, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	var req struct {
		IDNotificacao string `json:"idNotificacao"`
	}
	_ = json.Unmarshal(body, &req)
	id := strings.TrimSpace(req.IDNotificacao)
	if id == "" {
		resposta.Erro(w, http.StatusBadRequest, errors.New("idNotificacao obrigatorio"))
		return
	}
	if err := upsertLeitura(id, software, idUsuario, userTipo, idFranqueado, idCliente, visto, lido); err != nil {
		resposta.Erro(w, http.StatusBadRequest, err)
		return
	}
	resposta.JSON(w, http.StatusOK, map[string]string{"status": "OK"})
}

func listarMinhas(software, idUsuario, userTipo, userMaster, idFranqueado, idCliente string) ([]itemNotif, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
			n.ID_Notificacao, n.Titulo, n.Corpo, n.Tipo, n.Softwares, n.Publico,
			IFNULL(n.ID_Franqueado,''), IFNULL(n.ID_Cliente,''), n.Ativo,
			IFNULL(DATE_FORMAT(n.Valido_Ate, '%Y-%m-%d %H:%i:%s'),''),
			IFNULL(n.Criado_Por,''),
			DATE_FORMAT(n.Data_Cadastro, '%Y-%m-%d %H:%i:%s'),
			IFNULL(DATE_FORMAT(l.Visto_Em, '%Y-%m-%d %H:%i:%s'),''),
			IFNULL(DATE_FORMAT(l.Lido_Em, '%Y-%m-%d %H:%i:%s'),'')
		FROM cm_notificacao n
		LEFT JOIN cm_notificacao_leitura l
			ON l.ID_Notificacao = n.ID_Notificacao
			AND l.Software = ?
			AND l.ID_Usuario = ?
		WHERE n.Ativo = 'S'
		  AND (n.Valido_Ate IS NULL OR n.Valido_Ate > NOW())
		  AND (
			n.Softwares = 'todos'
			OR FIND_IN_SET(?, REPLACE(n.Softwares, ' ', '')) > 0
		  )
		  AND (
			n.Publico = 'todos'
			OR n.Publico = 'franqueados'
			OR (n.Publico = 'master' AND ? = 'S')
			OR (n.Publico = 'franqueado' AND n.ID_Franqueado = ? AND ? = 'FRA')
		  )
		ORDER BY n.Data_Cadastro DESC
		LIMIT 100
	`, software, idUsuario, software, userMaster, idFranqueado, userTipo)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	out := []itemNotif{}
	for tab.Next() {
		var it itemNotif
		if err := tab.Scan(
			&it.IDNotificacao, &it.Titulo, &it.Corpo, &it.Tipo, &it.Softwares, &it.Publico,
			&it.IDFranqueado, &it.IDCliente, &it.Ativo, &it.ValidoAte, &it.CriadoPor, &it.DataCadastro,
			&it.VistoEm, &it.LidoEm,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func contarNaoVistas(software, idUsuario, userTipo, userMaster, idFranqueado, idCliente string) (int, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	var total int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM cm_notificacao n
		LEFT JOIN cm_notificacao_leitura l
			ON l.ID_Notificacao = n.ID_Notificacao
			AND l.Software = ?
			AND l.ID_Usuario = ?
		WHERE n.Ativo = 'S'
		  AND (n.Valido_Ate IS NULL OR n.Valido_Ate > NOW())
		  AND (
			n.Softwares = 'todos'
			OR FIND_IN_SET(?, REPLACE(n.Softwares, ' ', '')) > 0
		  )
		  AND (
			n.Publico = 'todos'
			OR n.Publico = 'franqueados'
			OR (n.Publico = 'master' AND ? = 'S')
			OR (n.Publico = 'franqueado' AND n.ID_Franqueado = ? AND ? = 'FRA')
		  )
		  AND l.Visto_Em IS NULL
	`, software, idUsuario, software, userMaster, idFranqueado, userTipo).Scan(&total)
	return total, err
}

func upsertLeitura(idNotif, software, idUsuario, userTipo, idFranqueado, idCliente string, marcarVisto, marcarLido bool) error {
	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	agora := time.Now().Format("2006-01-02 15:04:05")
	var vistoEm, lidoEm interface{}
	if marcarVisto {
		vistoEm = agora
	}
	if marcarLido {
		lidoEm = agora
		if vistoEm == nil {
			vistoEm = agora
		}
	}

	res, err := db.Exec(`
		UPDATE cm_notificacao_leitura
		SET
			Visto_Em = COALESCE(Visto_Em, ?),
			Lido_Em = CASE WHEN ? IS NOT NULL THEN COALESCE(Lido_Em, ?) ELSE Lido_Em END,
			User_Tipo = COALESCE(NULLIF(?, ''), User_Tipo),
			ID_Franqueado = COALESCE(NULLIF(?, ''), ID_Franqueado),
			ID_Cliente = COALESCE(NULLIF(?, ''), ID_Cliente)
		WHERE ID_Notificacao = ? AND Software = ? AND ID_Usuario = ?
	`, vistoEm, lidoEm, lidoEm, userTipo, idFranqueado, idCliente, idNotif, software, idUsuario)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff > 0 {
		return nil
	}

	idLeitura := fmt.Sprintf("%020d", time.Now().UnixNano()%1e18)
	_, err = db.Exec(`
		INSERT INTO cm_notificacao_leitura (
			ID_Leitura, ID_Notificacao, Software, ID_Usuario,
			User_Tipo, ID_Franqueado, ID_Cliente, Visto_Em, Lido_Em
		) VALUES (?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)
	`, idLeitura, idNotif, software, idUsuario, nullEmpty(userTipo), idFranqueado, idCliente, vistoEm, lidoEm)
	return err
}

func nullEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
