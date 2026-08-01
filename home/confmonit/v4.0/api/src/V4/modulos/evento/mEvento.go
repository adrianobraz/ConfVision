package eventoV4

import (
	paginacaoV4 "api/src/V4/paginacaoV4"
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Evento struct {
	ID_Evento   string `json:"idEvento"`
	ID_Processo string `json:"idProcesso"`
	Codigo      string `json:"codigo,omitempty"`
	Particao    string `json:"particao,omitempty"`
	ZonaUser    string `json:"zonaUser,omitempty"`
	Nivel       string `json:"nivel"`
	Img         string `json:"img"`
	DataEntrada string `json:"dataEntrada"`
	// propriedades auxiliares
	Qtd       string `json:"qtd,omitempty"`
	Conta     string `json:"conta,omitempty"`
	Grupo     string `json:"grupo,omitempty"`
	Descricao string `json:"descricao,omitempty"`
	Filtro    string `json:"filtro,omitempty"`
}

type SEvento struct {
	ID_Evento   sql.NullString
	ID_Processo sql.NullString
	Codigo      sql.NullString
	Particao    sql.NullString
	ZonaUser    sql.NullString
	Nivel       sql.NullString
	Img         sql.NullString
	DataEntrada sql.NullTime
}

type EvtFiltro struct {
	DataEntrada string `json:"dataEntrada"`
	Codigo      string `json:"codigo,omitempty"`

	DispId            string `json:"dispId"`
	DispNome          string `json:"dispNome"`
	DispConta         string `json:"dispConta"`
	DispDataUltimoEvt string `dispDataUltimoEvt`

	CliId   string `json:"cliId"`
	CliNome string `json:"cliNome"`

	FraId string `json:"FraId"`

	CtiGrupo     string `json:"ctiGrupo"`
	CtiDescricao string `json:"ctiDescricao"`

	ZonaUser          string `json:"zonaUser"`
	ZonaUserDescricao string `json:"zonaUserDescricao"`

	// auxiliares
	DataInicio   string `json:"dataInicio,omitempty"`
	DataFim      string `json:"dataFim,omitempty"`
	Grupos       string `json:"grupos"`
	IdFranqueado string `json:"idFranqueado"`
	Grupo        string `json:"grupo"`
	Limit        int    `json:"limit"`
	Offset       int    `json:"offset"`
	Termo        string `json:"termo"`
}

func (e *Evento) Insere() error {
	if e.ID_Evento == "" {
		e.ID_Evento = auxiliar.GeradorDeId()
	}

	if e.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO evento(
			evento.ID_Evento, 
			evento.ID_Processo, 
			evento.Codigo, 
			evento.Particao, 
			evento.ZonaUser, 
			evento.Nivel, 
			evento.Img, 
			evento.DataEntrada
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		e.ID_Evento,
		e.ID_Processo,
		e.Codigo,
		e.Particao,
		e.ZonaUser,
		e.Nivel,
		e.Img,
		e.DataEntrada,
	); err != nil {
		return err
	}
	return nil
}

func (e *Evento) ListarByIdProcesso(lista *[]Evento) error {

	if e.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(` WHERE evento.ID_Processo = '%s'`, e.ID_Processo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Evento

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (e *Evento) ListarAgrupadoByIdProcesso(lista *[]Evento) error {
	if e.ID_Processo == "" {
		return errors.New("um id de processo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(` 
		WHERE evento.ID_Processo = '%s'
		GROUP BY evento.Codigo	
	`, e.ID_Processo)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Evento

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (e *Evento) ListarByDispProcOpen(idDisp string, lista *[]Evento) error {
	if idDisp == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(` 
		LEFT JOIN processo
		ON processo.ID_Processo = evento.ID_Processo

		WHERE processo.ID_Dispositivo = '%s'
		AND processo.DataAtenFim = NULL
		
	`, idDisp)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Evento

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// Funções internas ===========================================================
func sEventoToEvento(evt *Evento, sEvt SEvento) {
	evt.ID_Evento = sEvt.ID_Evento.String
	evt.ID_Processo = sEvt.ID_Processo.String
	evt.Codigo = sEvt.Codigo.String
	evt.Particao = sEvt.Particao.String
	evt.ZonaUser = sEvt.ZonaUser.String
	evt.Nivel = sEvt.Nivel.String
	evt.Img = sEvt.Img.String

	if sEvt.DataEntrada.Valid {
		evt.DataEntrada = sEvt.DataEntrada.Time.Format("02/01/2006 15:04:05")
	} else {
		evt.DataEntrada = ""
	}
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			evento.ID_Evento, 
			evento.ID_Processo, 
			evento.Codigo, 
			evento.Particao, 
			evento.ZonaUser, 
			evento.Nivel, 
			evento.Img, 
			evento.DataEntrada 
		FROM evento
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, item *Evento) error {

	var tmp SEvento
	if err := tab.Scan(
		&tmp.ID_Evento,
		&tmp.ID_Processo,
		&tmp.Codigo,
		&tmp.Particao,
		&tmp.ZonaUser,
		&tmp.Nivel,
		&tmp.Img,
		&tmp.DataEntrada,
	); err != nil {
		return err
	}

	sEventoToEvento(item, tmp)
	return nil
}

// funcoes para evento por filtro =============================================

func (ef *EvtFiltro) ListarByDispStartEnd(lista *[]EvtFiltro) (int, error) {
	// Modalidade "todos os clientes": lista por franqueado (nao por dispositivo unico).
	if strings.EqualFold(strings.TrimSpace(ef.DispId), "TODOS") {
		if strings.TrimSpace(ef.IdFranqueado) == "" {
			return 0, errors.New("um id de franqueado deve ser informado")
		}
		return ef.ListarByIdFranqueadoStartEndGrupo(lista)
	}

	if ef.DispId == "" {
		return 0, errors.New("um id de dispositivo deve ser informado")
	}

	if ef.DataInicio == "" {
		return 0, errors.New("uma data inicial deve ser informada")
	}

	if ef.DataFim == "" {
		return 0, errors.New("uma data final deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	fromWhere := `
		FROM evento 
		LEFT JOIN processo ON evento.ID_Processo = processo.ID_Processo
		LEFT JOIN dispositivo ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo
		LEFT JOIN cliente ON dispositivo.ID_Cliente = cliente.ID_Cliente
		LEFT JOIN franqueado ON cliente.ID_Franqueado = franqueado.ID_Franqueado
		LEFT JOIN contactId AS ctiPadrao
			ON evento.Codigo = ctiPadrao.Codigo AND ctiPadrao.ID_Vinculo = 'CENTRAL'
		LEFT JOIN contactId AS ctiPersonalizado
			ON evento.Codigo = ctiPersonalizado.Codigo AND ctiPersonalizado.ID_Vinculo = cliente.ID_Franqueado
		LEFT JOIN usuariosAlarme 
			ON dispositivo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo
			AND evento.ZonaUser = usuariosAlarme.Codigo
		LEFT JOIN setorAlarme
			ON dispositivo.ID_Dispositivo = setorAlarme.ID_Dispositivo
			AND evento.ZonaUser = setorAlarme.Numero
		WHERE evento.ID_Processo IN(
			SELECT processo.ID_Processo
			FROM processo
			WHERE processo.ID_Dispositivo = ?
			AND processo.DataCriacao >= ?
			AND processo.DataCriacao <= ?		
		)`

	args := []interface{}{ef.DispId, ef.DataInicio, ef.DataFim}

	total := 0
	if ef.Limit > 0 {
		sqlCount := `SELECT COUNT(*) ` + fromWhere
		var qtd sql.NullInt64
		if err := db.QueryRow(sqlCount, args...).Scan(&qtd); err != nil {
			return 0, err
		}
		total = int(qtd.Int64)
	}

	txtSql := `SELECT 
			evento.DataEntrada,
			evento.Codigo,
			evento.ZonaUser,
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			dispositivo.Conta,
			dispositivo.DataUltimoEvento,
			cliente.ID_Cliente,
			cliente.Nome,
			franqueado.RazaoSocial,
			ctiPadrao.Descricao,
			ctiPadrao.Grupo,
			ctiPersonalizado.Descricao,
			ctiPersonalizado.Grupo,
			usuariosAlarme.Nome,
			usuariosAlarme.Codigo,
			setorAlarme.Nome,
			setorAlarme.Numero
		` + fromWhere + `
		ORDER BY evento.DataEntrada DESC` + paginacaoV4.Clausula(ef.Limit, ef.Offset)

	tab, err := db.Query(txtSql, args...)
	if err != nil {
		return 0, err
	}

	defer tab.Close()

	for tab.Next() {
		var (
			dataEntrada sql.NullTime
			codigo      sql.NullString
			zonaUser    sql.NullString

			idDispositivo     sql.NullString
			dispNome          sql.NullString
			dispConta         sql.NullString
			dispDataUltimoEvt sql.NullTime

			idCliente sql.NullString
			cliNome   sql.NullString

			idFranqueado sql.NullString

			padDescricao sql.NullString
			padGrupo     sql.NullString

			perDescricao sql.NullString
			perGrupo     sql.NullString

			usuNome   sql.NullString
			usuCodigo sql.NullString

			setNome   sql.NullString
			setNumero sql.NullString
		)

		if err := tab.Scan(
			&dataEntrada,
			&codigo,
			&zonaUser,

			&idDispositivo,
			&dispNome,
			&dispConta,
			&dispDataUltimoEvt,

			&idCliente,
			&cliNome,

			&idFranqueado,

			&padDescricao,
			&padGrupo,

			&perDescricao,
			&perGrupo,

			&usuNome,
			&usuCodigo,

			&setNome,
			&setNumero,
		); err != nil {
			return 0, err
		}
		var item EvtFiltro
		item.DataEntrada = dataEntrada.Time.Format("02/01/2006 15:04:05")
		item.Codigo = codigo.String

		item.DispId = idDispositivo.String
		item.DispNome = dispNome.String
		item.DispConta = dispConta.String
		item.DispDataUltimoEvt = dispDataUltimoEvt.Time.Format("02/01/2006 15:04:05")

		item.CliId = idCliente.String
		item.CliNome = cliNome.String

		item.FraId = idFranqueado.String

		if perDescricao.Valid {
			item.CtiDescricao = perDescricao.String
			item.CtiGrupo = perGrupo.String
		} else {
			item.CtiDescricao = padDescricao.String
			item.CtiGrupo = padGrupo.String
		}

		//println(perGrupo.String, padGrupo.String, usuNome.String, setNome.String)

		if perGrupo.String == "ARME" || padGrupo.String == "ARME" || perGrupo.String == "DESARME" || padGrupo.String == "DESARME" {
			item.ZonaUser = usuCodigo.String
			item.ZonaUserDescricao = strings.ToUpper(usuNome.String)
			if item.ZonaUser == "" {
				item.ZonaUser = zonaUser.String
				if zonaUser.String == "000" {
					item.ZonaUserDescricao = "MASTER/APP"
				}
			}
		} else {
			item.ZonaUser = setNumero.String
			item.ZonaUserDescricao = strings.ToUpper(setNome.String)
			if item.ZonaUser == "" {
				item.ZonaUser = zonaUser.String
				if zonaUser.String == "000" {
					item.ZonaUserDescricao = "CENTRAL/TECLADO"
				}
			}
		}

		*lista = append(*lista, item)
	}
	if ef.Limit <= 0 {
		total = len(*lista)
	}
	return total, nil
}

func (ef *EvtFiltro) ListarByTodosStartEnd(lista *[]EvtFiltro) error {
	if ef.DispId == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if ef.DataInicio == "" {
		return errors.New("uma data inicial deve ser informada")
	}

	if ef.DataFim == "" {
		return errors.New("uma data final deve ser informada")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT 
			evento.DataEntrada,
			evento.Codigo,
			
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			dispositivo.Conta,
			dispositivo.DataUltimoEvento,

			cliente.ID_Cliente,
			cliente.Nome,

			franqueado.RazaoSocial,

			ctiPadrao.Descricao,
			ctiPadrao.Grupo,

			ctiPersonalizado.Descricao,
			ctiPersonalizado.Grupo,

			usuariosAlarme.Nome,
			usuariosAlarme.Codigo,

			setorAlarme.Nome,
			setorAlarme.Numero

			
		FROM evento 

		LEFT JOIN processo
		ON evento.ID_Processo = processo.ID_Processo

		LEFT JOIN dispositivo
		ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN contactId AS ctiPadrao
		ON evento.Codigo = ctiPadrao.Codigo
		AND ctiPadrao.ID_Vinculo = "CENTRAL"
		
		LEFT JOIN contactId AS ctiPersonalizado
		ON evento.Codigo = ctiPadrao.Codigo
		AND ctiPadrao.ID_Vinculo = cliente.ID_Franqueado
		
		LEFT JOIN usuariosAlarme 
		ON dispositivo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo
		AND evento.ZonaUser = usuariosAlarme.Codigo

		LEFT JOIN setorAlarme
		ON dispositivo.ID_Dispositivo = setorAlarme.ID_Dispositivo
		AND evento.ZonaUser = setorAlarme.Numero
		
		WHERE evento.ID_Processo IN(
			SELECT processo.ID_Processo
			FROM processo
			WHERE processo.ID_Dispositivo = ?
			AND processo.DataCriacao >= ?
			AND processo.DataCriacao <= ?		
		)
		
		ORDER BY evento.DataEntrada DESC
	`, ef.DispId, ef.DataInicio, ef.DataFim)
	if err != nil {
		return err
	}

	defer tab.Close()

	for tab.Next() {
		var (
			dataEntrada sql.NullTime
			codigo      sql.NullString

			idDispositivo     sql.NullString
			dispNome          sql.NullString
			dispConta         sql.NullString
			dispDataUltimoEvt sql.NullTime

			idCliente sql.NullString
			cliNome   sql.NullString

			idFranqueado sql.NullString

			padDescricao sql.NullString
			padGrupo     sql.NullString

			perDescricao sql.NullString
			perGrupo     sql.NullString

			usuNome   sql.NullString
			usuCodigo sql.NullString

			setNome   sql.NullString
			setNumero sql.NullString
		)

		if err := tab.Scan(
			&dataEntrada,
			&codigo,

			&idDispositivo,
			&dispNome,
			&dispConta,
			&dispDataUltimoEvt,

			&idCliente,
			&cliNome,

			&idFranqueado,

			&padDescricao,
			&padGrupo,

			&perDescricao,
			&perGrupo,

			&usuNome,
			&usuCodigo,

			&setNome,
			&setNumero,
		); err != nil {
			return err
		}
		var item EvtFiltro
		item.DataEntrada = dataEntrada.Time.Format("02/01/2006 15:04:05")
		item.Codigo = codigo.String

		item.DispId = idDispositivo.String
		item.DispNome = dispNome.String
		item.DispConta = dispConta.String
		item.DispDataUltimoEvt = dispDataUltimoEvt.Time.Format("02/01/2006 15:04:05")

		item.CliId = idCliente.String
		item.CliNome = cliNome.String

		item.FraId = idFranqueado.String

		if perDescricao.Valid {
			item.CtiDescricao = perDescricao.String
			item.CtiGrupo = perGrupo.String
		} else {
			item.CtiDescricao = padDescricao.String
			item.CtiGrupo = padGrupo.String
		}

		//println(perGrupo.String, padGrupo.String, usuNome.String, setNome.String)

		if perGrupo.String == "ARME" || padGrupo.String == "ARME" || perGrupo.String == "DESARME" || padGrupo.String == "DESARME" {
			item.ZonaUser = usuCodigo.String
			item.ZonaUserDescricao = strings.ToUpper(usuNome.String)
		} else {
			item.ZonaUser = setNumero.String
			item.ZonaUserDescricao = strings.ToUpper(setNome.String)
		}

		if item.ZonaUser == "" {
			item.ZonaUser = "NÃO CADASTRADO"
		}

		if item.ZonaUserDescricao == "" {
			item.ZonaUserDescricao = "NÃO CADASTRADO"
		}
		*lista = append(*lista, item)
	}
	return nil
}

func (ef *EvtFiltro) ListarByDispStartEndGrupo(lista *[]EvtFiltro) error {
	if ef.DispId == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if ef.DataInicio == "" {
		return errors.New("uma data inicial deve ser informada")
	}

	if ef.DataFim == "" {
		return errors.New("uma data final deve ser informada")
	}
	if ef.Grupos == "" {
		return errors.New("pelomenos um grupo de eventos deve ser infomado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	txtSql := fmt.Sprintf(`
			SELECT 
				evento.DataEntrada,
				evento.Codigo,
				
				dispositivo.ID_Dispositivo,
				dispositivo.Nome,
				dispositivo.Conta,

				cliente.ID_Cliente,
				cliente.Nome,

				franqueado.RazaoSocial,

				ctiPadrao.Descricao,
				ctiPadrao.Grupo,

				ctiPersonalizado.Descricao,
				ctiPersonalizado.Grupo,

				usuariosAlarme.Nome,
				usuariosAlarme.Codigo,

				setorAlarme.Nome,
				setorAlarme.Numero

				
			FROM evento 

			LEFT JOIN processo
			ON evento.ID_Processo = processo.ID_Processo

			LEFT JOIN dispositivo
			ON processo.ID_Dispositivo = dispositivo.ID_Dispositivo

			LEFT JOIN cliente
			ON dispositivo.ID_Cliente = cliente.ID_Cliente

			LEFT JOIN franqueado
			ON cliente.ID_Franqueado = franqueado.ID_Franqueado

			LEFT JOIN contactId AS ctiPadrao
			ON evento.Codigo = ctiPadrao.Codigo
			AND ctiPadrao.ID_Vinculo = "CENTRAL"
			
			LEFT JOIN contactId AS ctiPersonalizado
			ON evento.Codigo = ctiPadrao.Codigo
			AND ctiPadrao.ID_Vinculo = cliente.ID_Franqueado
			
			LEFT JOIN usuariosAlarme 
			ON dispositivo.ID_Dispositivo = usuariosAlarme.ID_Dispositivo
			AND evento.ZonaUser = usuariosAlarme.Codigo

			LEFT JOIN setorAlarme
			ON dispositivo.ID_Dispositivo = setorAlarme.ID_Dispositivo
			AND evento.ZonaUser = setorAlarme.Numero
			
			WHERE evento.ID_Processo IN(
				SELECT processo.ID_Processo
				FROM processo
				WHERE processo.ID_Dispositivo = %s
				AND processo.DataCriacao >= '%s'
				AND processo.DataCriacao <= '%s'		
			)
			AND evento.Codigo IN (
				SELECT contactId.Codigo 
				FROM contactId 
				WHERE contactId.Grupo IN(%s)
			)
	`, ef.DispId, ef.DataInicio, ef.DataFim, ef.Grupos)

	tab, err := db.Query(txtSql)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var (
			dataEntrada sql.NullTime
			codigo      sql.NullString

			idDispositivo sql.NullString
			dispNome      sql.NullString
			dispConta     sql.NullString

			idCliente sql.NullString
			cliNome   sql.NullString

			idFranqueado sql.NullString

			padDescricao sql.NullString
			padGrupo     sql.NullString

			perDescricao sql.NullString
			perGrupo     sql.NullString

			usuNome   sql.NullString
			usuCodigo sql.NullString

			setNome   sql.NullString
			setNumero sql.NullString
		)

		if err := tab.Scan(
			&dataEntrada,
			&codigo,

			&idDispositivo,
			&dispNome,
			&dispConta,

			&idCliente,
			&cliNome,

			&idFranqueado,

			&padDescricao,
			&padGrupo,

			&perDescricao,
			&perGrupo,

			&usuNome,
			&usuCodigo,

			&setNome,
			&setNumero,
		); err != nil {
			return err
		}
		var item EvtFiltro
		item.DataEntrada = dataEntrada.Time.Format("02/01/2006 15:04:05")
		item.Codigo = codigo.String

		item.DispId = idDispositivo.String
		item.DispNome = dispNome.String
		item.DispConta = dispConta.String

		item.CliId = idCliente.String
		item.CliNome = cliNome.String

		item.FraId = idFranqueado.String

		if perDescricao.Valid {
			item.CtiDescricao = perDescricao.String
			item.CtiGrupo = perGrupo.String
		} else {
			item.CtiDescricao = padDescricao.String
			item.CtiGrupo = padGrupo.String
		}

		if perGrupo.String == "ARME" || perGrupo.String == "DESARME" {
			item.ZonaUser = usuCodigo.String
			item.ZonaUserDescricao = usuNome.String
		} else {
			item.ZonaUser = setNumero.String
			item.ZonaUserDescricao = setNome.String
		}

		if item.ZonaUser == "" {
			item.ZonaUser = "NÃO CADASTRADO"
		}

		if item.ZonaUserDescricao == "" {
			item.ZonaUserDescricao = "NÃO CADASTRADO"
		}
		*lista = append(*lista, item)
	}
	return nil
}
