package centroOperacional

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
	"webAmbiente/src/auxiliar"
	"webAmbiente/src/xano"
)

// ProcessoFila item da fila do Centro Operacional (um processo aberto).
type ProcessoFila struct {
	IdProcesso       string `json:"idProcesso"`
	IdCliente        string `json:"idCliente"`
	IdFranqueado     string `json:"idFranqueado"`
	NomeCliente      string `json:"nomeCliente"`
	IdDispositivo    string `json:"idDispositivo"`
	NomeDispositivo  string `json:"nomeDispositivo"`
	Nivel            string `json:"nivel"`
	DataHora         string `json:"dataHora"`
	DataCriacao      string `json:"dataCriacao"`
	Grupo            string `json:"grupo"`
	DescricaoGrupo   string `json:"descricaoGrupo"`
	Codigo           string `json:"codigo"`
	Quantidade       int    `json:"quantidade"`
	Particao         string `json:"particao"`
	ZonaUser         string `json:"zonaUser"`
	NomeSetor        string `json:"nomeSetor"`
	IdSetor          string `json:"idSetor"`
	CameraOn         string `json:"cameraOn"`
	StatusAtend      string `json:"statusAtendimento"`
	IdAtendente      string `json:"idAtendente"`
	NomeOperador     string `json:"nomeOperador"`
	SomSirene        bool   `json:"somSirene"`
}

// EventoGrupo eventos agrupados por código+zona (como webTerminal).
type EventoGrupo struct {
	Quantidade      int    `json:"quantidade"`
	Codigo          string `json:"codigo"`
	CodigoFull      string `json:"codigoFull"`
	Descricao       string `json:"descricao"`
	Grupo           string `json:"grupo"`
	Particao        string `json:"particao"`
	ZonaUser        string `json:"zonaUser"`
	DescZona        string `json:"descZona"`
	Hora            string `json:"hora"`
	IdDispositivo   string `json:"idDispositivo"`
	IdSetor         string `json:"idSetor"`
	CameraOn        string `json:"cameraOn"`
}

// ContadoresMapa totais do mapa ativo.
type ContadoresMapa struct {
	NomeLoja            string `json:"nomeLoja"`
	Online              bool   `json:"online"`
	KeepAlive           int    `json:"keepAlive"`
	TotalZonas          int    `json:"totalZonas"`
	TotalAlarmes        int    `json:"totalAlarmes"`
	TotalSemComunicacao int    `json:"totalSemComunicacao"`
	Armado              string `json:"armado"`
}

func listarProcessosCliente(idCliente, idFranqueado string, limite int) ([]ProcessoFila, error) {
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return nil, fmt.Errorf("idCliente obrigatorio")
	}
	if limite <= 0 || limite > 100 {
		limite = 50
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT
			processo.ID_Processo,
			processo.Nivel,
			processo.DataCriacao,
			processo.ID_Atendente,
			usuarios.Nick,
			cliente.ID_Cliente,
			cliente.Nome,
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			cliente.ID_Franqueado,
			(SELECT evento.Codigo FROM evento
			 WHERE evento.ID_Processo = processo.ID_Processo
			 ORDER BY evento.DataEntrada DESC LIMIT 1),
			(SELECT evento.Particao FROM evento
			 WHERE evento.ID_Processo = processo.ID_Processo
			 ORDER BY evento.DataEntrada DESC LIMIT 1),
			(SELECT evento.ZonaUser FROM evento
			 WHERE evento.ID_Processo = processo.ID_Processo
			 ORDER BY evento.DataEntrada DESC LIMIT 1),
			(SELECT COUNT(*) FROM evento WHERE evento.ID_Processo = processo.ID_Processo),
			(SELECT MAX(evento.DataEntrada) FROM evento WHERE evento.ID_Processo = processo.ID_Processo),
			setorAlarme.ID_Setor,
			setorAlarme.Nome,
			setorAlarme.Camera
		FROM processo
		INNER JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
		INNER JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
		LEFT JOIN usuarios ON processo.ID_Atendente = usuarios.ID_Usuario
		LEFT JOIN setorAlarme
			ON setorAlarme.ID_Dispositivo = dispositivo.ID_Dispositivo
			AND setorAlarme.Numero = (
				SELECT e2.ZonaUser FROM evento e2
				WHERE e2.ID_Processo = processo.ID_Processo
				ORDER BY e2.DataEntrada DESC LIMIT 1
			)
			AND setorAlarme.Particao = (
				SELECT e3.Particao FROM evento e3
				WHERE e3.ID_Processo = processo.ID_Processo
				ORDER BY e3.DataEntrada DESC LIMIT 1
			)
		WHERE cliente.ID_Cliente = ?
			AND processo.DataAtenFim IS NULL
			AND processo.Nivel > 0
		ORDER BY processo.Nivel DESC, processo.DataCriacao ASC
		LIMIT %d
	`, limite)

	tab, err := db.Query(query, idCliente)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	return scanProcessosFila(tab, db)
}

func listarProcessosOperador(op map[string]string, limite int) ([]ProcessoFila, error) {
	if limite <= 0 || limite > 200 {
		limite = 100
	}

	tipo := strings.TrimSpace(op["userTipo"])
	idVinculo := strings.TrimSpace(op["idVinculo"])

	filtro := ""
	args := make([]interface{}, 0)

	switch tipo {
	case "FRA":
		filtro = "AND cliente.ID_Franqueado = ?"
		args = append(args, idVinculo)
	case "REP":
		ids, err := franqueadosIDsRepresentante(idVinculo)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []ProcessoFila{}, nil
		}
		ph := strings.Repeat("?,", len(ids))
		ph = strings.TrimSuffix(ph, ",")
		filtro = fmt.Sprintf("AND cliente.ID_Franqueado IN (%s)", ph)
		for _, id := range ids {
			args = append(args, id)
		}
	case "CEN":
		// todos os franqueados
	default:
		if idVinculo != "" {
			filtro = "AND cliente.ID_Franqueado = ?"
			args = append(args, idVinculo)
		}
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT
			processo.ID_Processo,
			processo.Nivel,
			processo.DataCriacao,
			processo.ID_Atendente,
			usuarios.Nick,
			cliente.ID_Cliente,
			cliente.Nome,
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			cliente.ID_Franqueado,
			(SELECT evento.Codigo FROM evento
			 WHERE evento.ID_Processo = processo.ID_Processo
			 ORDER BY evento.DataEntrada DESC LIMIT 1),
			(SELECT evento.Particao FROM evento
			 WHERE evento.ID_Processo = processo.ID_Processo
			 ORDER BY evento.DataEntrada DESC LIMIT 1),
			(SELECT evento.ZonaUser FROM evento
			 WHERE evento.ID_Processo = processo.ID_Processo
			 ORDER BY evento.DataEntrada DESC LIMIT 1),
			(SELECT COUNT(*) FROM evento WHERE evento.ID_Processo = processo.ID_Processo),
			(SELECT MAX(evento.DataEntrada) FROM evento WHERE evento.ID_Processo = processo.ID_Processo),
			setorAlarme.ID_Setor,
			setorAlarme.Nome,
			setorAlarme.Camera
		FROM processo
		INNER JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
		INNER JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
		LEFT JOIN usuarios ON processo.ID_Atendente = usuarios.ID_Usuario
		LEFT JOIN setorAlarme
			ON setorAlarme.ID_Dispositivo = dispositivo.ID_Dispositivo
			AND setorAlarme.Numero = (
				SELECT e2.ZonaUser FROM evento e2
				WHERE e2.ID_Processo = processo.ID_Processo
				ORDER BY e2.DataEntrada DESC LIMIT 1
			)
			AND setorAlarme.Particao = (
				SELECT e3.Particao FROM evento e3
				WHERE e3.ID_Processo = processo.ID_Processo
				ORDER BY e3.DataEntrada DESC LIMIT 1
			)
		WHERE processo.DataAtenFim IS NULL
			AND processo.Nivel > 0
			%s
		ORDER BY processo.Nivel DESC, processo.DataCriacao ASC
		LIMIT %d
	`, filtro, limite)

	tab, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	return scanProcessosFila(tab, db)
}

func franqueadosIDsRepresentante(idRepresentante string) ([]string, error) {
	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT franqueado.ID_Franqueado
		FROM franqueado
		WHERE franqueado.ID_Representante = ?
		AND franqueado.DataCancelamento IS NULL
	`, idRepresentante)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	ids := make([]string, 0)
	for tab.Next() {
		var id sql.NullString
		if err := tab.Scan(&id); err != nil {
			return nil, err
		}
		if strings.TrimSpace(id.String) != "" {
			ids = append(ids, id.String)
		}
	}
	return ids, nil
}

func scanProcessosFila(tab *sql.Rows, db *sql.DB) ([]ProcessoFila, error) {
	out := make([]ProcessoFila, 0)
	for tab.Next() {
		var (
			item        ProcessoFila
			nivel       sql.NullString
			dataCriacao sql.NullTime
			idAtend     sql.NullString
			nickOp      sql.NullString
			idFranq     sql.NullString
			codigo      sql.NullString
			particao    sql.NullString
			zonaUser    sql.NullString
			qtd         sql.NullInt64
			dataUlt     sql.NullTime
			idSetor     sql.NullString
			nomeSetor   sql.NullString
			camera      sql.NullString
		)
		if err := tab.Scan(
			&item.IdProcesso,
			&nivel,
			&dataCriacao,
			&idAtend,
			&nickOp,
			&item.IdCliente,
			&item.NomeCliente,
			&item.IdDispositivo,
			&item.NomeDispositivo,
			&idFranq,
			&codigo,
			&particao,
			&zonaUser,
			&qtd,
			&dataUlt,
			&idSetor,
			&nomeSetor,
			&camera,
		); err != nil {
			return nil, err
		}

		item.Nivel = strings.TrimSpace(nivel.String)
		item.Codigo = strings.TrimSpace(codigo.String)
		item.Particao = strings.TrimSpace(particao.String)
		item.ZonaUser = strings.TrimSpace(zonaUser.String)
		item.IdSetor = idSetor.String
		item.NomeSetor = nomeSetor.String
		item.CameraOn = strings.TrimSpace(camera.String)
		if item.CameraOn == "" {
			item.CameraOn = "N"
		}
		item.IdAtendente = strings.TrimSpace(idAtend.String)
		item.NomeOperador = strings.TrimSpace(nickOp.String)
		item.IdFranqueado = strings.TrimSpace(idFranq.String)
		if item.IdAtendente != "" && item.IdAtendente != "0" {
			item.StatusAtend = "ATENDIDO"
		} else {
			item.StatusAtend = "ABERTO"
		}
		if qtd.Valid {
			item.Quantidade = int(qtd.Int64)
		}
		if dataCriacao.Valid {
			item.DataCriacao = dataCriacao.Time.Format("02/01/2006 15:04:05")
		}
		if dataUlt.Valid {
			item.DataHora = dataUlt.Time.Format("02/01/2006 15:04:05")
		}

		var grupo, descricao string
		_ = auxiliar.GetCtiGrupoAndDescricaoDB(db, idFranq.String, item.Codigo, &grupo, &descricao)
		item.Grupo = grupo
		item.DescricaoGrupo = descricao
		item.SomSirene = grupoDisparaSirene(grupo)

		out = append(out, item)
	}
	return out, nil
}

func listarEventosAgrupados(idProcesso, idFranqueado string) ([]EventoGrupo, error) {
	idProcesso = strings.TrimSpace(idProcesso)
	if idProcesso == "" {
		return nil, fmt.Errorf("idProcesso obrigatorio")
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT
			COUNT(evento.ID_Evento),
			evento.Codigo,
			evento.Particao,
			evento.ZonaUser,
			MAX(evento.DataEntrada),
			processo.ID_Dispositivo,
			usuariosAlarme.Nome,
			setorAlarme.ID_Setor,
			setorAlarme.Nome,
			setorAlarme.Camera
		FROM evento
		INNER JOIN processo ON evento.ID_Processo = processo.ID_Processo
		LEFT JOIN usuariosAlarme
			ON processo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo
			AND evento.ZonaUser = usuariosAlarme.codigo
		LEFT JOIN setorAlarme
			ON processo.ID_Dispositivo = setorAlarme.ID_Dispositivo
			AND evento.ZonaUser = setorAlarme.Numero
			AND evento.Particao = setorAlarme.Particao
		WHERE evento.ID_Processo = ?
		GROUP BY evento.Codigo, evento.ZonaUser
		ORDER BY MAX(evento.DataEntrada) ASC
	`, idProcesso)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	out := make([]EventoGrupo, 0)
	for tab.Next() {
		var (
			item        EventoGrupo
			codigoFull  sql.NullString
			particao    sql.NullString
			zonaUser    sql.NullString
			dataEntrada sql.NullTime
			idDisp      sql.NullString
			userNome    sql.NullString
			idSetor     sql.NullString
			setorNome   sql.NullString
			camera      sql.NullString
		)
		if err := tab.Scan(
			&item.Quantidade,
			&codigoFull,
			&particao,
			&zonaUser,
			&dataEntrada,
			&idDisp,
			&userNome,
			&idSetor,
			&setorNome,
			&camera,
		); err != nil {
			return nil, err
		}

		item.CodigoFull = strings.TrimSpace(codigoFull.String)
		item.Codigo = strings.TrimPrefix(item.CodigoFull, "E")
		item.Codigo = strings.TrimPrefix(item.Codigo, "R")
		item.Particao = strings.TrimSpace(particao.String)
		item.ZonaUser = strings.TrimSpace(zonaUser.String)
		item.IdDispositivo = idDisp.String
		item.IdSetor = idSetor.String
		item.CameraOn = strings.TrimSpace(camera.String)
		if item.CameraOn == "" {
			item.CameraOn = "N"
		}

		if strings.TrimSpace(setorNome.String) != "" {
			item.DescZona = setorNome.String
		} else if strings.TrimSpace(userNome.String) != "" {
			item.DescZona = userNome.String
		} else {
			item.DescZona = "Zona " + item.ZonaUser
		}

		if dataEntrada.Valid {
			item.Hora = dataEntrada.Time.Format("02/01/2006 15:04:05")
		}

		var grupo, descricao string
		_ = auxiliar.GetCtiGrupoAndDescricaoDB(db, idFranqueado, item.CodigoFull, &grupo, &descricao)
		item.Grupo = grupo
		item.Descricao = descricao

		out = append(out, item)
	}
	return out, nil
}

func grupoDisparaSirene(grupo string) bool {
	g := strings.ToUpper(strings.TrimSpace(grupo))
	g = strings.NewReplacer("Ã", "A", "É", "E", "Ê", "E", "Í", "I", "Ó", "O").Replace(g)
	chaves := []string{"ALARME", "EMERGENCIA", "FALHA", "MEDICO", "PANICO"}
	for _, c := range chaves {
		if strings.Contains(g, c) {
			return true
		}
	}
	return false
}

func finalizarProcesso(idProcesso, idCliente, descricao, idOperador, nomeOperador string) error {
	idProcesso = strings.TrimSpace(idProcesso)
	if idProcesso == "" {
		return fmt.Errorf("idProcesso obrigatorio")
	}
	descricao = strings.TrimSpace(descricao)
	if descricao == "" {
		return fmt.Errorf("descricao obrigatoria")
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tag := fmt.Sprintf("[%s - %s] ", time.Now().Format("02/01/2006 - 15:04:05"), strings.ToUpper(nomeOperador))
	descFinal := strings.ToUpper(tag + descricao)

	stm, err := db.Prepare(`
		UPDATE processo
		SET processo.Descricao = ?,
			processo.ID_Atendente = ?,
			processo.DataAtenFim = ?
		WHERE processo.ID_Processo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(descFinal, idOperador, time.Now().Format("2006-01-02 15:04:05"), idProcesso); err != nil {
		return err
	}

	return gravarTarifacao(db, idCliente, nomeOperador)
}

func gravarTarifacao(db *sql.DB, idVinculo, atendente string) error {
	stm, err := db.Prepare(`
		INSERT INTO tarifacao(ID_Tarifacao, ID_Vinculo, TipoOperacao, DadoOperacao, Aux)
		VALUES (?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	_, err = stm.Exec(auxiliar.GeradorDeId(), idVinculo, "ATENDIMENTO", atendente, "")
	return err
}

func contadoresDoMapa(mapaId int) (ContadoresMapa, error) {
	var out ContadoresMapa

	setores, err := xano.ListarSetoresPorMapa(mapaId)
	if err != nil {
		return out, err
	}
	out.TotalZonas = len(setores)

	dispSet := make(map[string]struct{})
	idSetores := make([]string, 0, len(setores))
	for _, s := range setores {
		if s.IdSetor != "" {
			idSetores = append(idSetores, s.IdSetor)
		}
		if s.IdDispositivo != "" {
			dispSet[s.IdDispositivo] = struct{}{}
		}
		if out.NomeLoja == "" && strings.TrimSpace(s.DispositivoNome) != "" {
			out.NomeLoja = s.DispositivoNome
		}
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return out, err
	}
	defer db.Close()

	if len(idSetores) > 0 {
		ph := strings.Repeat("?,", len(idSetores))
		ph = strings.TrimSuffix(ph, ",")
		args := make([]interface{}, len(idSetores))
		for i, id := range idSetores {
			args[i] = id
		}
		query := fmt.Sprintf(`
			SELECT COUNT(DISTINCT setorAlarme.ID_Setor)
			FROM setorAlarme
			INNER JOIN processo ON processo.ID_Dispositivo = setorAlarme.ID_Dispositivo
				AND processo.DataAtenFim IS NULL AND processo.Nivel > 0
			INNER JOIN evento ON evento.ID_Processo = processo.ID_Processo
				AND evento.ZonaUser = setorAlarme.Numero
			WHERE setorAlarme.ID_Setor IN (%s)
		`, ph)
		var qtd sql.NullInt64
		_ = db.QueryRow(query, args...).Scan(&qtd)
		if qtd.Valid {
			out.TotalAlarmes = int(qtd.Int64)
		}
	}

	for idDisp := range dispSet {
		var keepAlive sql.NullString
		var dataUlt sql.NullTime
		var nome, armado sql.NullString
		_ = db.QueryRow(`
			SELECT dispositivo.Nome, dispositivo.KeepAlive, dispositivo.DataUltimoEvento, dispositivo.Armado
			FROM dispositivo WHERE dispositivo.ID_Dispositivo = ?
		`, idDisp).Scan(&nome, &keepAlive, &dataUlt, &armado)

		if out.NomeLoja == "" && nome.Valid {
			out.NomeLoja = nome.String
		}
		out.Armado = strings.TrimSpace(armado.String)

		keepMin, _ := strconv.Atoi(strings.TrimSpace(keepAlive.String))
		out.KeepAlive = keepMin
		online := true
		if keepMin > 0 {
			if !dataUlt.Valid {
				online = false
			} else if time.Since(dataUlt.Time) > 10*time.Minute {
				online = false
			}
		}
		if !online {
			out.TotalSemComunicacao++
		}
		if len(dispSet) == 1 {
			out.Online = online
		} else if !online {
			out.Online = false
		}
	}

	return out, nil
}
