/*
*******************************************************
Modulo S006 -> Verefica a existencia de dispositivos
que estajam ativos e OFF-LINE e envia um email para
informar o franqueado do ocorrido
*******************************************************
*/
package S006

import (
	"database/sql"
	"fmt"
	"robot/src/auxiliar"
	"time"
)

type tDisp struct {
	fraId       string
	fraNome     string
	fraEmail    string
	dispId      string
	dispNome    string
	dispDataEvt string
	dispCodigo  string
	Autorizado  bool
}

func Start(horario int) {
	var erro error
	fmt.Println("Subindo modulo S006 -> Controle dipositivos sem comunicacao + 24H")

	for {
		// processa somente a 3 horas da tarde
		if time.Now().Hour() == horario {
			fmt.Println("Modulo S006 -> Processando disp sem comunicação")

			if erro = buscaDispositivos(); erro != nil {
				fmt.Println("Erro modulo S006 ->", erro.Error())
				break
			}
		}

		time.Sleep(time.Hour)
	}

	fmt.Println("Reiniciando modulo S006 -> Controle dipositivos sem comunicacao + 24H")

	go Start(horario)
}

// buscaDispositivos busca todos dispositivo com DataUltimo evento menor que 24 horas
func buscaDispositivos() error {
	// Abre um canal de conexao
	db, erro := auxiliar.Conectar()
	if erro != nil {
		return erro
	}
	defer db.Close()

	// dataStart := time.Now().Add(-48 * time.Hour).Format("2006-01-02 15:04:05")
	data := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")

	tab, erro := db.Query(`

		SELECT
			franqueado.ID_Franqueado,
			franqueado.RazaoSocial,
			usuarios.Email1,
			
			dispositivo.ID_Dispositivo,
			dispositivo.Nome,
			dispositivo.DataUltimoEvento,
			dispositivo.CodgioUltimoEvento,

			blocFra.Email,
			blocRep.Email

		FROM dispositivo 

		LEFT JOIN cliente
		ON dispositivo.ID_Cliente = cliente.ID_Cliente
					
		LEFT JOIN franqueado
		ON cliente.ID_Franqueado = franqueado.ID_Franqueado

		LEFT JOIN usuarios
		ON franqueado.ID_UsuarioMaster = usuarios.ID_Vinculo

		LEFT JOIN listaEnvio blocFra
		ON franqueado.ID_Franqueado = blocFra.ID_Alvo

		LEFT JOIN listaEnvio blocRep
		ON franqueado.ID_Representante = blocRep.ID_Alvo

		WHERE dispositivo.ID_Dispositivo NOT IN (
			SELECT listaBloqueio.ID_Alvo
			FROM listaBloqueio
		)
		
		AND dispositivo.DataUltimoEvento < ?
	`, data)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	for tab.Next() {
		var disp tDisp
		var idFra, nomeFra, emailFra sql.NullString
		var dispDataUltimoEvento sql.NullTime
		var enviaFra, enviaRep sql.NullString
		if erro := tab.Scan(
			&idFra,
			&nomeFra,
			&emailFra,
			&disp.dispId,
			&disp.dispNome,
			&dispDataUltimoEvento,
			&disp.dispCodigo,
			&enviaFra,
			&enviaRep,
		); erro != nil {
			return erro
		}

		disp.fraId = idFra.String
		disp.fraNome = nomeFra.String
		disp.fraEmail = emailFra.String
		disp.dispDataEvt = dispDataUltimoEvento.Time.Format("02/01/2006 15:04:05")

		// Verifica se o envio do franqueado esta liberado
		if enviaFra.Valid {
			if enviaFra.String == "S" {

				// Caso envio franquedado esteja liberado ele verifica o envio do representante
				if enviaRep.Valid {
					if enviaRep.String == "S" {
						disp.Autorizado = true
					} else {
						disp.Autorizado = false
					}

				} else {
					disp.Autorizado = false
				}

			} else {
				disp.Autorizado = false
			}
		} else {
			disp.Autorizado = false
		}

		// envia o email para tabela envio
		if erro := enviarEmailFranqueado(db, disp); erro != nil {
			return erro
		}

	}

	return nil
}

// enviarEmailFranqueado envia um email informativo para o franqueado
func enviarEmailFranqueado(db *sql.DB, disp tDisp) error {
	if disp.Autorizado {

		// Cria a menssagem
		msg := fmt.Sprintf(`
			<h3>Aviso de dispositivo off-line a mais de 24 horas</h3>

			<p>
				Dispositivo <strong>%s</strong> enviou ultimo evento em 
				<strong>%s</strong> com <br> código <strong>%s</strong>, 
				sendo assim ele esta a mais de 24h<br> sem comunicar
			</p>
		`, disp.dispNome, disp.dispDataEvt, disp.dispCodigo)

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
			disp.fraId,
			"FRA",
			disp.fraEmail,
			msg,
		); err != nil {
			return err
		}
	}

	return nil
}
