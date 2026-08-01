/*********** MODULO S003 **********
* Sistema nao armado
 */
package S003

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"robot/src/auxiliar"
	"strconv"
	"strings"
	"time"
)

type Grade struct {
	IdFranqueado string
	FraNome      string
	FraEmail     string

	IdCliente string
	CliNome   string
	CliEmail  string

	IdDispositivo string
	IdGrade       string
	Tolerancia    string
	Conta         string
	GradeAlvo     string
	EnviaEmailCli bool
	EnviaEmailFra bool
	EnviaEmailRep bool
}

type Evento struct {
	IdDispositivo string
	IdEvento      string
	Nivel         string
	IdProcesso    string
	Codigo        string
	Conta         string
	Particao      string
	ZonaUser      string
	Img           string
}

type listaProc struct {
	tempoEnvio time.Time
	tempoFinal time.Time
	enviado    bool
}

var listaProcessado = make(map[string]listaProc)

type S003 struct{}

func Start(tempo time.Duration) {

	fmt.Println("Subindo modulo S003 -> Sistema não armado")
	time.Sleep(100 * time.Millisecond)

	for {
		time.Sleep(tempo * time.Second)
		fmt.Printf("Modulo S003 -> Processando sistema não armado\n")

		// Abre um canal de conexao
		db, erro := auxiliar.Conectar()
		if erro != nil {
			fmt.Println("MODULO S003 ERROR: ", erro.Error())
			break
		}
		defer db.Close()

		// Busca uma lista de grades a processar
		lista, erro := buscarGrades(db)
		if erro != nil {
			fmt.Println("MODULO S003 ERROR: ", erro.Error())
			break
		}

		// Remove itens que nao estao mais desarmados.
		// Assim, quando armar e desarmar novamente, um novo ciclo de envio e aberto.
		dispositivosDesarmados := make(map[string]bool)
		for _, i := range lista {
			dispositivosDesarmados[i.IdDispositivo] = true
		}
		for id := range listaProcessado {
			if !dispositivosDesarmados[id] {
				delete(listaProcessado, id)
			}
		}

		// Processa a lista de grades
		for _, i := range lista {
			dLista, ok := listaProcessado[i.IdDispositivo]

			if !ok {
				dLista.tempoEnvio = time.Now()
				dLista.tempoFinal = time.Now().Add(180 * time.Minute)
				dLista.enviado = false
				listaProcessado[i.IdDispositivo] = dLista
			}

			tempoEnvio := dLista.tempoEnvio
			tempoFinal := dLista.tempoFinal

			enviar := time.Now().After(tempoEnvio)

			if enviar && !dLista.enviado {
				// Regra antiga (até 3 envios):
				// dLista.tempoEnvio = time.Now().Add(30 * time.Minute)

				fmt.Println("Processando grade", i.CliNome)

				// Verifica seo o item é não armado
				naoArmado, erro := calcularNaoArmado(i.GradeAlvo, i.Tolerancia)
				if erro != nil {
					fmt.Println("MODULO S003 ERROR: ", erro.Error())
					break
				}

				// Processa caso seja nao armado
				if naoArmado {
					if erro := processaNaoArmado(db, i); erro != nil {
						fmt.Println("MODULO S003 ERROR: ", erro.Error())
						break
					}

					// Nova regra: envia 3M02 somente uma vez por ciclo
					dLista.enviado = true
					listaProcessado[i.IdDispositivo] = dLista
				}
			}

			// Remove o dispositivo da lista
			remover := time.Now().After(tempoFinal)
			if remover {
				delete(listaProcessado, i.IdDispositivo)
			}
		}

		db.Close()

	}

	fmt.Println("Reiniciando modulo S003 -> Sistema não armado no horario")
	go Start(tempo)
}

func processaNaoArmado(db *sql.DB, item Grade) error {
	// Cria um objeto evento pra processar o evento nao armado
	var evt Evento

	// IdDispositivo
	evt.IdDispositivo = item.IdDispositivo

	// Trava de banco: evita duplicidade do 3M02 na mesma janela
	jaEnviado, erro := jaEnviouNaoArmadoRecente(db, evt.IdDispositivo)
	if erro != nil {
		return erro
	}
	if jaEnviado {
		return nil
	}

	//IdEvento
	evt.IdEvento = auxiliar.GeradorDeId()

	// Nivel
	if erro := carregaNivel(
		db, item.IdFranqueado, &evt.Nivel,
	); erro != nil {
		return erro
	}

	// IdProcesuo
	if err := carregarIdProcesso(
		db, evt.IdDispositivo, evt.Nivel, &evt.IdProcesso,
	); err != nil {
		return err
	}

	// Codigo
	evt.Codigo = "3M02"

	// Conta
	evt.Conta = item.Conta

	// Particao
	evt.Particao = "00"

	// ZonaUser
	evt.ZonaUser = "000"

	// Img
	evt.Img = ""

	if erro := gravaEvento(db, evt); erro != nil {
		return erro
	}

	// Envia para o franqueado
	if erro := enviarEmailFranqueado(db, item); erro != nil {
		return erro
	}

	// Envia para o cliente
	if erro := enviarEmailCliente(db, item); erro != nil {
		return erro
	}

	// Envia evento 3M02 para o webhook do Xano
	if erro := notificaAppWebhook(evt, item); erro != nil {
		fmt.Printf("MODULO S003 -> Erro ao enviar 3M02 para webhook: %v\n", erro)
		// não retorna erro para não interromper o fluxo (evento já foi gravado e emails enviados)
	}
	return nil
}

// notificaAppWebhook envia o evento 3M02 (sistema não armado) para o webhook do Xano
func notificaAppWebhook(evt Evento, item Grade) error {
	url := "https://xpcy-oyme-lno7.b2.xano.io/api:G5ijUd6Z/alarm_events_webhook"

	payload := map[string]string{
		"idEvento":      evt.IdEvento,
		"codigo":        evt.Codigo,
		"particao":      evt.Particao,
		"zonaUser":      evt.ZonaUser,
		"nivel":         evt.Nivel,
		"dataEntrada":   time.Now().Format("2006-01-02 15:04:05"),
		"img":           evt.Img,
		"idProcesso":    evt.IdProcesso,
		"idDispositivo": evt.IdDispositivo,
		"idCliente":     item.IdCliente,
		"nomeCliente":   item.CliNome,
		"emailCliente":  item.CliEmail,
		"ctiGrupo":      "GERAL",
		"ctiDescricao":  "Sistema não armado",
		"idFranqueado":  item.IdFranqueado,
		"codigoBenuvem": "",
		"conta":         evt.Conta,
		"carmeraAtiva":  "N",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer 1e2d2eef75ccd7cfea2e06d7ce1c63c4")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook retornou status %d", resp.StatusCode)
	}

	return nil
}

func buscarGrades(db *sql.DB) ([]Grade, error) {

	txtSql := fmt.Sprintf(`
		SELECT	
			grade.ID_Grade,
			grade.%s, 
			grade.Tolerancia,
			
			dispositivo.ID_Dispositivo,
			dispositivo.Conta,
			
			cliente.ID_Cliente,
			cliente.Nome,
			cliente.Email1,
			
			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,
			usuarios.Email1,
			
			blocEmailCli.Email,
			blocEmailFra.Email,
			blocEmailRep.Email

		FROM dispositivo

		LEFT JOIN grade
		ON dispositivo.ID_Dispositivo = grade.ID_Dispositivo
		
		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente

		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN usuarios
		ON franqueado.ID_UsuarioMaster = usuarios.ID_Usuario	

		LEFT JOIN listaEnvio AS blocEmailCli
		ON cliente.ID_Cliente = blocEmailCli.ID_Alvo

		LEFT JOIN listaEnvio AS blocEmailFra
		ON franqueado.ID_Franqueado = blocEmailFra.ID_Alvo

		LEFT JOIN listaEnvio AS blocEmailRep
		ON franqueado.ID_Representante = blocEmailRep.ID_Alvo
		
		WHERE dispositivo.Armado = "N"
		
		AND grade.ID_Grade NOT IN (
			SELECT listaBloqueio.ID_Alvo FROM listaBloqueio
		)	
		`, pegarDia(),
	)

	tab, erro := db.Query(txtSql)
	if erro != nil {
		return nil, erro
	}
	defer tab.Close()

	var lista []Grade
	for tab.Next() {
		var i Grade
		var cliId, cliNome, cliEmail sql.NullString
		var fraId, fraNome, fraEmail sql.NullString
		var blocEmailCli, blocEmailFra, blocEmailRep sql.NullString

		if erro := tab.Scan(
			&i.IdGrade,
			&i.GradeAlvo,
			&i.Tolerancia,

			&i.IdDispositivo,
			&i.Conta,

			&cliId,
			&cliNome,
			&cliEmail,

			&fraId,
			&fraNome,
			&fraEmail,

			&blocEmailCli,
			&blocEmailFra,
			&blocEmailRep,
		); erro != nil {
			return nil, erro
		}
		i.IdCliente = cliId.String
		i.CliNome = cliNome.String
		i.CliEmail = cliEmail.String

		i.IdFranqueado = fraId.String
		i.FraNome = fraNome.String
		i.FraEmail = fraEmail.String

		if blocEmailCli.Valid {
			if blocEmailCli.String == "S" {
				i.EnviaEmailCli = true
			} else {
				i.EnviaEmailCli = false
			}
		} else {
			i.EnviaEmailCli = false
		}

		if blocEmailFra.Valid {
			if blocEmailFra.String == "S" {
				i.EnviaEmailFra = true
			} else {
				i.EnviaEmailFra = false
			}
		} else {
			i.EnviaEmailFra = false
		}

		if blocEmailRep.Valid {
			if blocEmailRep.String == "S" {
				i.EnviaEmailRep = true
			} else {
				i.EnviaEmailRep = false
			}
		} else {
			i.EnviaEmailRep = false
		}

		lista = append(lista, i)

	}
	return lista, nil
}

func pegarDia() string {
	switch int(time.Now().Weekday()) {
	case 0:
		return "DomSpm"
	case 1:
		return "SegSpm"
	case 2:
		return "TerSpm"
	case 3:
		return "QuaSpm"
	case 4:
		return "QuiSpm"
	case 5:
		return "SexSpm"
	default:
		return "SabSpm"
	}
}

func calcularNaoArmado(grade, tol string) (bool, error) {
	// Nao permite grade no fomrmato errado
	valido, _ := regexp.MatchString("^([0-1]?[0-9]|2[0-3]):[0-5][0-9]$", grade)
	if valido {
		sliceHora := strings.Split(grade, ":")

		// Converte os valores da Hora da grade para inteiro
		hora, erro := strconv.Atoi(sliceHora[0])
		if erro != nil {

			return false, erro
		}
		// Converte os valores Minutos da grade para inteiro
		minuto, erro := strconv.Atoi(sliceHora[1])
		if erro != nil {

			return false, erro
		}

		// Converte tolerancia para duração em minutos
		periodo, erro := time.ParseDuration(tol + "m")
		if erro != nil {
			return false, erro
		}

		// Cria um time.Time com os dados vindo da grade
		t := time.Now()

		gradeHora := time.Date(
			t.Year(), t.Month(), t.Day(), hora, minuto,
			0, t.Nanosecond(), t.Location(),
		).Add(periodo)

		agora := time.Now()

		if agora.After(gradeHora) && agora.Before(gradeHora.Add(70*time.Minute)) {
			return true, nil
		}
	}
	return false, nil
}

// carregaNivel carrega o nivel de atendimento do evento para criação
func carregaNivel(db *sql.DB, idFra string, nivel *string) error {
	if idFra == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	tab, erro := db.Query(`
		SELECT contactId.Nivel
		FROM contactId 		
		WHERE contactId.ID_Vinculo = ? 		
		AND contactId.Codigo = "3M02"
	`, idFra)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	if tab.Next() { // Caso exista um contactid personalizado
		var niv sql.NullString
		if erro := tab.Scan(
			&niv,
		); erro != nil {
			return erro
		}

		*nivel = niv.String
		return nil
	} else { // Senao procura na base central
		tab, erro = db.Query(`
			SELECT contactId.Nivel
			FROM contactId 		
			WHERE contactId.ID_Vinculo = "CENTRAL"		
			AND contactId.Codigo = "3M02"
		`)
		if erro != nil {
			return erro
		}
		defer tab.Close()

		if tab.Next() {
			var niv sql.NullString
			if erro := tab.Scan(
				&niv,
			); erro != nil {
				return erro
			}
			*nivel = niv.String
			return nil
		}
	}
	*nivel = "0"

	return nil
}

// evento
// carregarIdProcesso carrega o numero do processo para criação do evento
func carregarIdProcesso(db *sql.DB, idDisp, nivelIn string, idProc *string) error {

	// Obriga o recebimento de id de dispositivo
	if idDisp == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	tab, erro := db.Query(`
		SELECT 
			processo.ID_Processo, 
			processo.Nivel

		FROM processo 
		
		WHERE processo.ID_Dispositivo = ?
		
		AND processo.DataAtenFim IS NULL		
	`, idDisp)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	if tab.Next() { // Caso encontre um processo ele atualiza o nivel

		// Recebe o valor do processo
		var id sql.NullString
		var nivelProc sql.NullString
		if err := tab.Scan(
			&id, // pega o id do processo para retorno
			&nivelProc,
		); err != nil {
			return err
		}
		*idProc = id.String
		nivelEvento, erro := strconv.Atoi(nivelIn)
		if erro != nil {
			fmt.Println("error 1", nivelIn)
			return erro
		}

		nivelProcesso, erro := strconv.Atoi(nivelProc.String)
		if erro != nil {
			return erro
		}

		// Se o nivel do evento for maior que o do processo atualiza o
		// processo e mantem o nivel do evento
		if nivelEvento > nivelProcesso {
			stm, erro := db.Prepare(`
					UPDATE processo 
					SET processo.Nivel = ?
					WHERE processo.ID_Processo = ?
				`)
			if erro != nil {
				return erro
			}
			defer stm.Close()

			if _, erro := stm.Exec(
				nivelIn,
				idProc,
			); erro != nil {
				return erro
			}
		}

	} else { // Senao gera um processo
		// Gera um id para o novo processo
		*idProc = auxiliar.GeradorDeId()

		smt, erro := db.Prepare(`
			INSERT INTO processo(
				processo.ID_Processo, 
				processo.ID_Dispositivo, 
				processo.Nivel
			)values(?,?,?)
		`)
		if erro != nil {
			return erro
		}
		defer smt.Close()

		if _, erro := smt.Exec(
			*idProc,
			idDisp,
			nivelIn,
		); erro != nil {
			return erro
		}
	}

	return nil
}

// enviarEmailFranqueado envia um email informativo para o franqueado
func enviarEmailFranqueado(db *sql.DB, item Grade) error {
	if item.EnviaEmailFra {
		var agora = time.Now().Format("02/01/2006 15:04:05")

		// Cria a menssagem
		msg := fmt.Sprintf(`
		<h1>Informativo de Alarme não Armado</h1> <br>
		<p>Na data de %s, o alarme de %s não foi armado no horário</p>
		`, agora, item.CliNome)

		stm, err := db.Prepare(`
			INSERT INTO email(
				ID_Email, 
				ID_Vinculo, 
				TipoVinculo, 
				Destinatario, 
				Menssagem
			) VALUES (?,?,?,?,?)
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(
			auxiliar.GeradorDeId(),
			item.IdFranqueado,
			"FRA",
			item.FraEmail,
			msg,
		); err != nil {
			return err
		}
	}
	return nil
}

// enviarEmailCliente envia um email informativo para o cliente
func enviarEmailCliente(db *sql.DB, item Grade) error {
	if item.EnviaEmailCli {
		var agora = time.Now().Format("02/01/2006 15:04:05")

		// Cria a mensagem
		msg := fmt.Sprintf(`
		<h1>Informativo de Alarme não Armado</h1>
		<br>
		<p>Sr.(a) %s, na data de %s, o seu alarme não foi armado no horário</p>
	`, item.CliNome, agora)

		stm, err := db.Prepare(`
			INSERT INTO email(
				ID_Email, 
				ID_Vinculo, 
				TipoVinculo, 
				Destinatario, 
				Menssagem
			) VALUES (?,?,?,?,?)
		`)
		if err != nil {
			return err
		}
		defer stm.Close()

		if _, err := stm.Exec(
			auxiliar.GeradorDeId(),
			item.IdCliente,
			"CLI",
			item.CliEmail,
			msg,
		); err != nil {
			return err
		}
	}

	return nil
}

// Grava o evento recebido e ja tratado no banco
func gravaEvento(db *sql.DB, evt Evento) error {

	if evt.IdEvento == "" {
		evt.IdEvento = auxiliar.GeradorDeId()
	}

	if evt.IdProcesso == "" {
		return errors.New("um id de processo deve ser informado")
	}

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
		evt.IdEvento,
		evt.IdProcesso,
		evt.Codigo,
		evt.Particao,
		evt.ZonaUser,
		evt.Nivel,
		evt.Img,
		time.Now().Format("2006-01-02 15:04:05"),
	); err != nil {
		return err
	}
	return nil
}

// jaEnviouNaoArmadoRecente verifica se ja existe 3M02 recente para o dispositivo
func jaEnviouNaoArmadoRecente(db *sql.DB, idDispositivo string) (bool, error) {
	if idDispositivo == "" {
		return false, errors.New("um id de dispositivo deve ser informado")
	}

	tab, err := db.Query(`
		SELECT evento.ID_Evento
		FROM evento
		INNER JOIN processo
		ON processo.ID_Processo = evento.ID_Processo
		WHERE processo.ID_Dispositivo = ?
		AND evento.Codigo = "3M02"
		AND evento.DataEntrada >= DATE_SUB(NOW(), INTERVAL 180 MINUTE)
		LIMIT 1
	`, idDispositivo)
	if err != nil {
		return false, err
	}
	defer tab.Close()

	return tab.Next(), nil
}
