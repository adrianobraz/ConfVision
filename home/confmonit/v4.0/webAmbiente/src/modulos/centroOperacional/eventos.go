package centroOperacional

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"webAmbiente/src/auxiliar"
)

// EventoFila item da fila de eventos do Centro Operacional.
type EventoFila struct {
	IdEvento        string `json:"idEvento"`
	IdProcesso      string `json:"idProcesso"`
	IdDispositivo   string `json:"idDispositivo"`
	NomeDispositivo string `json:"nomeDispositivo"`
	IdSetor         string `json:"idSetor"`
	NomeSetor       string `json:"nomeSetor"`
	Codigo          string `json:"codigo"`
	Titulo          string `json:"titulo"`
	Particao        string `json:"particao"`
	ZonaUser        string `json:"zonaUser"`
	Local           string `json:"local"`
	CodigoLabel     string `json:"codigoLabel"`
	DataHora        string `json:"dataHora"`
	Prioridade      string `json:"prioridade"`
	Categoria       string `json:"categoria"`
	Nivel           string `json:"nivel"`
}

func listarEventosCliente(idCliente string, limite int) ([]EventoFila, error) {
	idCliente = strings.TrimSpace(idCliente)
	if idCliente == "" {
		return nil, fmt.Errorf("idCliente obrigatorio")
	}
	if limite <= 0 || limite > 200 {
		limite = 100
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT
			evento.ID_Evento,
			evento.ID_Processo,
			evento.Codigo,
			evento.Particao,
			evento.ZonaUser,
			evento.DataEntrada,
			processo.Nivel,
			processo.ID_Dispositivo,
			dispositivo.Nome,
			setorAlarme.ID_Setor,
			setorAlarme.Nome
		FROM evento
		INNER JOIN processo ON evento.ID_Processo = processo.ID_Processo
		INNER JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
		INNER JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
		LEFT JOIN setorAlarme
			ON setorAlarme.ID_Dispositivo = dispositivo.ID_Dispositivo
			AND setorAlarme.Numero = evento.ZonaUser
			AND setorAlarme.Particao = evento.Particao
		WHERE cliente.ID_Cliente = ?
			AND processo.DataAtenFim IS NULL
			AND processo.Nivel > 0
		ORDER BY evento.DataEntrada DESC, evento.ID_Evento DESC
		LIMIT %d
	`, limite)

	tab, err := db.Query(query, idCliente)
	if err != nil {
		return nil, err
	}
	defer tab.Close()

	out := make([]EventoFila, 0)
	for tab.Next() {
		var (
			item          EventoFila
			dataEntrada   sql.NullTime
			nivel         sql.NullString
			idDisp        sql.NullString
			dispNome      sql.NullString
			idSetor       sql.NullString
			nomeSetor     sql.NullString
			codigo        sql.NullString
			particao      sql.NullString
			zonaUser      sql.NullString
			idEvento      sql.NullString
			idProcesso    sql.NullString
		)
		if err := tab.Scan(
			&idEvento,
			&idProcesso,
			&codigo,
			&particao,
			&zonaUser,
			&dataEntrada,
			&nivel,
			&idDisp,
			&dispNome,
			&idSetor,
			&nomeSetor,
		); err != nil {
			return nil, err
		}

		item.IdEvento = idEvento.String
		item.IdProcesso = idProcesso.String
		item.Codigo = strings.TrimSpace(codigo.String)
		item.Particao = strings.TrimSpace(particao.String)
		item.ZonaUser = strings.TrimSpace(zonaUser.String)
		item.Nivel = strings.TrimSpace(nivel.String)
		item.IdDispositivo = idDisp.String
		item.NomeDispositivo = dispNome.String
		item.IdSetor = idSetor.String
		item.NomeSetor = nomeSetor.String

		if dataEntrada.Valid {
			item.DataHora = dataEntrada.Time.Format("02/01/2006 15:04:05")
		}

		item.Titulo = tituloEvento(item.Codigo)
		item.Prioridade = prioridadePorNivel(item.Nivel)
		item.Categoria = categoriaPorCodigo(item.Codigo, item.Nivel)
		item.CodigoLabel = fmt.Sprintf("Z-%s / P%s", padZona(item.ZonaUser), padParticao(item.Particao))
		item.Local = montarLocal(item.NomeDispositivo, item.NomeSetor, item.ZonaUser)

		out = append(out, item)
	}
	return out, nil
}

func tituloEvento(codigo string) string {
	c := strings.TrimSpace(codigo)
	if c == "" {
		return "Evento"
	}
	return "Evento " + c
}

func prioridadePorNivel(nivel string) string {
	n, _ := strconv.Atoi(strings.TrimSpace(nivel))
	if n >= 3 {
		return "ALTA"
	}
	if n >= 2 {
		return "MEDIA"
	}
	return "INFO"
}

func categoriaPorCodigo(codigo, nivel string) string {
	c := strings.ToUpper(strings.TrimSpace(codigo))
	switch {
	case strings.HasPrefix(c, "E") || strings.Contains(c, "FALHA") || strings.Contains(c, "OFF"):
		return "falhas"
	case strings.HasPrefix(c, "R") || strings.HasPrefix(c, "O") || strings.Contains(c, "ARM"):
		return "controle"
	default:
		n, _ := strconv.Atoi(strings.TrimSpace(nivel))
		if n > 0 {
			return "alarmes"
		}
		return "todos"
	}
}

func padZona(z string) string {
	z = strings.TrimSpace(z)
	if z == "" {
		return "000"
	}
	if len(z) >= 3 {
		return z
	}
	return strings.Repeat("0", 3-len(z)) + z
}

func padParticao(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return "00"
	}
	if len(p) >= 2 {
		return p
	}
	return strings.Repeat("0", 2-len(p)) + p
}

func montarLocal(disp, setor, zona string) string {
	partes := make([]string, 0, 3)
	if strings.TrimSpace(disp) != "" {
		partes = append(partes, strings.TrimSpace(disp))
	}
	if strings.TrimSpace(setor) != "" {
		partes = append(partes, strings.TrimSpace(setor))
	} else if strings.TrimSpace(zona) != "" {
		partes = append(partes, "Zona "+strings.TrimSpace(zona))
	}
	return strings.Join(partes, " — ")
}
func statusDispositivo(idDispositivo string) (armado, senha, particao string, err error) {
	idDispositivo = strings.TrimSpace(idDispositivo)
	if idDispositivo == "" {
		return "", "", "", fmt.Errorf("idDispositivo obrigatorio")
	}

	db, err := auxiliar.Conectar()
	if err != nil {
		return "", "", "", err
	}
	defer db.Close()

	var arm sql.NullString
	var sen sql.NullString
	var part sql.NullString
	err = db.QueryRow(`
		SELECT dispositivo.Armado, dispositivo.Senha, dispositivo.Particao
		FROM dispositivo
		WHERE dispositivo.ID_Dispositivo = ?
	`, idDispositivo).Scan(&arm, &sen, &part)
	if err == sql.ErrNoRows {
		return "N", "", "01", nil
	}
	if err != nil {
		return "", "", "", err
	}
	return strings.TrimSpace(arm.String), strings.TrimSpace(sen.String), strings.TrimSpace(part.String), nil
}
