package notificacao

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"errors"
	"strings"
	"time"
)

type Notificacao struct {
	ID_Notificacao string `json:"idNotificacao"`
	Titulo         string `json:"titulo"`
	Corpo          string `json:"corpo"`
	Tipo           string `json:"tipo"`
	Softwares      string `json:"softwares"`
	Publico        string `json:"publico"`
	ID_Franqueado  string `json:"idFranqueado"`
	ID_Cliente     string `json:"idCliente"`
	Ativo          string `json:"ativo"`
	Valido_Ate     string `json:"validoAte"`
	Criado_Por     string `json:"criadoPor"`
	Data_Cadastro  string `json:"dataCadastro"`
	Visto_Em       string `json:"vistoEm,omitempty"`
	Lido_Em        string `json:"lidoEm,omitempty"`
}

type ContextoUsuario struct {
	Software      string `json:"software"`
	ID_Usuario    string `json:"idUsuario"`
	User_Tipo     string `json:"userTipo"`
	User_Master   string `json:"userMaster"`
	ID_Franqueado string `json:"idFranqueado"`
	ID_Cliente    string `json:"idCliente"`
}

type LeituraReq struct {
	ID_Notificacao string `json:"idNotificacao"`
	Software       string `json:"software"`
	ID_Usuario     string `json:"idUsuario"`
	User_Tipo      string `json:"userTipo"`
	ID_Franqueado  string `json:"idFranqueado"`
	ID_Cliente     string `json:"idCliente"`
}

type DesativarReq struct {
	ID_Notificacao string `json:"idNotificacao"`
}

func (n *Notificacao) Criar() error {
	n.Titulo = strings.TrimSpace(n.Titulo)
	n.Corpo = strings.TrimSpace(n.Corpo)
	if n.Titulo == "" {
		return errors.New("titulo obrigatorio")
	}
	if n.Corpo == "" {
		return errors.New("corpo obrigatorio")
	}

	n.Tipo = strings.ToLower(strings.TrimSpace(n.Tipo))
	if n.Tipo == "" {
		n.Tipo = "info"
	}
	if n.Tipo != "info" && n.Tipo != "aviso" && n.Tipo != "urgente" {
		return errors.New("tipo invalido")
	}

	n.Softwares = normalizarSoftwares(n.Softwares)
	n.Publico = strings.ToLower(strings.TrimSpace(n.Publico))
	if n.Publico == "" {
		n.Publico = "todos"
	}

	validos := map[string]bool{
		"todos": true, "master": true, "franqueados": true, "franqueado": true,
	}
	if !validos[n.Publico] {
		return errors.New("publico invalido (todos|master|franqueados|franqueado)")
	}

	n.ID_Franqueado = strings.TrimSpace(n.ID_Franqueado)
	n.ID_Cliente = ""

	if n.Publico == "franqueado" && n.ID_Franqueado == "" {
		return errors.New("idFranqueado obrigatorio para publico=franqueado")
	}
	if n.Publico != "franqueado" {
		n.ID_Franqueado = ""
	}

	if n.Ativo == "" {
		n.Ativo = "S"
	}
	if n.ID_Notificacao == "" {
		n.ID_Notificacao = idCurto()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	var validoAte interface{}
	va := strings.TrimSpace(n.Valido_Ate)
	if va == "" {
		validoAte = nil
	} else {
		validoAte = va
	}

	stm, err := db.Prepare(`
		INSERT INTO cm_notificacao (
			ID_Notificacao, Titulo, Corpo, Tipo, Softwares, Publico,
			ID_Franqueado, ID_Cliente, Ativo, Valido_Ate, Criado_Por
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	_, err = stm.Exec(
		n.ID_Notificacao,
		n.Titulo,
		n.Corpo,
		n.Tipo,
		n.Softwares,
		n.Publico,
		nullIfEmpty(n.ID_Franqueado),
		nullIfEmpty(n.ID_Cliente),
		n.Ativo,
		validoAte,
		nullIfEmpty(strings.TrimSpace(n.Criado_Por)),
	)
	return err
}

func (n *Notificacao) ListarAdmin(lista *[]Notificacao, ativo string) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	ativo = strings.ToUpper(strings.TrimSpace(ativo))
	query := `
		SELECT
			ID_Notificacao, Titulo, Corpo, Tipo, Softwares, Publico,
			IFNULL(ID_Franqueado,''), IFNULL(ID_Cliente,''), Ativo,
			IFNULL(DATE_FORMAT(Valido_Ate, '%Y-%m-%d %H:%i:%s'),''),
			IFNULL(Criado_Por,''),
			DATE_FORMAT(Data_Cadastro, '%Y-%m-%d %H:%i:%s')
		FROM cm_notificacao
	`
	args := []interface{}{}
	if ativo == "S" || ativo == "N" {
		query += ` WHERE Ativo = ?`
		args = append(args, ativo)
	}
	query += ` ORDER BY Data_Cadastro DESC LIMIT 500`

	tab, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Notificacao
		if err := tab.Scan(
			&item.ID_Notificacao,
			&item.Titulo,
			&item.Corpo,
			&item.Tipo,
			&item.Softwares,
			&item.Publico,
			&item.ID_Franqueado,
			&item.ID_Cliente,
			&item.Ativo,
			&item.Valido_Ate,
			&item.Criado_Por,
			&item.Data_Cadastro,
		); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}
	return nil
}

func (n *Notificacao) Desativar() error {
	n.ID_Notificacao = strings.TrimSpace(n.ID_Notificacao)
	if n.ID_Notificacao == "" {
		return errors.New("idNotificacao obrigatorio")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`UPDATE cm_notificacao SET Ativo = 'N' WHERE ID_Notificacao = ?`)
	if err != nil {
		return err
	}
	defer stm.Close()

	res, err := stm.Exec(n.ID_Notificacao)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return errors.New("notificacao nao encontrada")
	}
	return nil
}

func Minhas(ctx ContextoUsuario, lista *[]Notificacao) error {
	ctx = normalizarContexto(ctx)
	if ctx.Software == "" || ctx.ID_Usuario == "" {
		return errors.New("software e idUsuario obrigatorios")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Publico so para usuarios do franqueado (FRA): normal ou master — nunca CLI.
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
		  AND ? = 'FRA'
		  AND (
			n.Publico = 'todos'
			OR n.Publico = 'franqueados'
			OR (n.Publico = 'master' AND ? = 'S')
			OR (n.Publico = 'franqueado' AND n.ID_Franqueado = ?)
		  )
		ORDER BY n.Data_Cadastro DESC
		LIMIT 100
	`,
		ctx.Software,
		ctx.ID_Usuario,
		ctx.Software,
		ctx.User_Tipo,
		ctx.User_Master,
		ctx.ID_Franqueado,
	)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Notificacao
		if err := tab.Scan(
			&item.ID_Notificacao,
			&item.Titulo,
			&item.Corpo,
			&item.Tipo,
			&item.Softwares,
			&item.Publico,
			&item.ID_Franqueado,
			&item.ID_Cliente,
			&item.Ativo,
			&item.Valido_Ate,
			&item.Criado_Por,
			&item.Data_Cadastro,
			&item.Visto_Em,
			&item.Lido_Em,
		); err != nil {
			return err
		}
		*lista = append(*lista, item)
	}
	return nil
}

func Contagem(ctx ContextoUsuario) (int, error) {
	ctx = normalizarContexto(ctx)
	if ctx.Software == "" || ctx.ID_Usuario == "" {
		return 0, errors.New("software e idUsuario obrigatorios")
	}

	db, err := connV4.Conectar()
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
		  AND ? = 'FRA'
		  AND (
			n.Publico = 'todos'
			OR n.Publico = 'franqueados'
			OR (n.Publico = 'master' AND ? = 'S')
			OR (n.Publico = 'franqueado' AND n.ID_Franqueado = ?)
		  )
		  AND l.Visto_Em IS NULL
	`,
		ctx.Software,
		ctx.ID_Usuario,
		ctx.Software,
		ctx.User_Tipo,
		ctx.User_Master,
		ctx.ID_Franqueado,
	).Scan(&total)
	return total, err
}

func MarcarVisto(req LeituraReq) error {
	return upsertLeitura(req, true, false)
}

func MarcarLido(req LeituraReq) error {
	return upsertLeitura(req, true, true)
}

func upsertLeitura(req LeituraReq, marcarVisto, marcarLido bool) error {
	req.ID_Notificacao = strings.TrimSpace(req.ID_Notificacao)
	req.Software = strings.ToLower(strings.TrimSpace(req.Software))
	req.ID_Usuario = strings.TrimSpace(req.ID_Usuario)
	req.User_Tipo = strings.ToUpper(strings.TrimSpace(req.User_Tipo))
	req.ID_Franqueado = strings.TrimSpace(req.ID_Franqueado)
	req.ID_Cliente = strings.TrimSpace(req.ID_Cliente)

	if req.ID_Notificacao == "" || req.Software == "" || req.ID_Usuario == "" {
		return errors.New("idNotificacao, software e idUsuario obrigatorios")
	}

	db, err := connV4.Conectar()
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

	// tenta update primeiro
	res, err := db.Exec(`
		UPDATE cm_notificacao_leitura
		SET
			Visto_Em = COALESCE(Visto_Em, ?),
			Lido_Em = CASE WHEN ? IS NOT NULL THEN COALESCE(Lido_Em, ?) ELSE Lido_Em END,
			User_Tipo = COALESCE(NULLIF(?, ''), User_Tipo),
			ID_Franqueado = COALESCE(NULLIF(?, ''), ID_Franqueado),
			ID_Cliente = COALESCE(NULLIF(?, ''), ID_Cliente)
		WHERE ID_Notificacao = ? AND Software = ? AND ID_Usuario = ?
	`,
		vistoEm,
		lidoEm, lidoEm,
		req.User_Tipo,
		req.ID_Franqueado,
		req.ID_Cliente,
		req.ID_Notificacao,
		req.Software,
		req.ID_Usuario,
	)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff > 0 {
		return nil
	}

	idLeitura := idCurto()
	_, err = db.Exec(`
		INSERT INTO cm_notificacao_leitura (
			ID_Leitura, ID_Notificacao, Software, ID_Usuario,
			User_Tipo, ID_Franqueado, ID_Cliente, Visto_Em, Lido_Em
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		idLeitura,
		req.ID_Notificacao,
		req.Software,
		req.ID_Usuario,
		nullIfEmpty(req.User_Tipo),
		nullIfEmpty(req.ID_Franqueado),
		nullIfEmpty(req.ID_Cliente),
		vistoEm,
		lidoEm,
	)
	return err
}

func idCurto() string {
	id := auxiliar.GeradorDeId()
	if len(id) > 20 {
		return id[len(id)-20:]
	}
	return id
}

func normalizarSoftwares(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || s == "todos" {
		return "todos"
	}
	partes := strings.Split(s, ",")
	validos := map[string]bool{
		"franqueadopro": true,
		"confvision":    true,
		"webambiente":   true,
	}
	out := []string{}
	seen := map[string]bool{}
	for _, p := range partes {
		p = strings.TrimSpace(p)
		if !validos[p] || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	if len(out) == 0 {
		return "todos"
	}
	if len(out) == 3 {
		return "todos"
	}
	return strings.Join(out, ",")
}

func normalizarContexto(ctx ContextoUsuario) ContextoUsuario {
	ctx.Software = strings.ToLower(strings.TrimSpace(ctx.Software))
	ctx.ID_Usuario = strings.TrimSpace(ctx.ID_Usuario)
	ctx.User_Tipo = strings.ToUpper(strings.TrimSpace(ctx.User_Tipo))
	ctx.User_Master = strings.ToUpper(strings.TrimSpace(ctx.User_Master))
	if ctx.User_Master != "S" {
		ctx.User_Master = "N"
	}
	ctx.ID_Franqueado = strings.TrimSpace(ctx.ID_Franqueado)
	ctx.ID_Cliente = strings.TrimSpace(ctx.ID_Cliente)
	return ctx
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
