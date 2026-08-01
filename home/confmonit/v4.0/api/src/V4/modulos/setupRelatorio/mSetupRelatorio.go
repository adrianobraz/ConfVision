package setupRelatorioV4

import (
	connV4 "api/src/V4/conexao"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type SetupRelatorio struct {
	ID_SetupRelatorio string `json:"idSetupRelatorio"`
	ID_Cliente        string `json:"idCliente"`
	Alarme            string `json:"alarme"`
	Arme              string `json:"arme"`
	Desarme           string `json:"desarme"`
	Emergencia        string `json:"emergencia"`
	Falhas            string `json:"falhas"`
	Geral             string `json:"geral"`
	Medico            string `json:"medico"`
	Panico            string `json:"panico"`
	Restaure          string `json:"restaure"`
	Setup             string `json:"setup"`
	Teste             string `json:"teste"`
}

type SSetupRelatorio struct {
	ID_SetupRelatorio sql.NullString
	ID_Cliente        sql.NullString
	Alarme            sql.NullString
	Arme              sql.NullString
	Desarme           sql.NullString
	Emergencia        sql.NullString
	Falhas            sql.NullString
	Geral             sql.NullString
	Medico            sql.NullString
	Panico            sql.NullString
	Restaure          sql.NullString
	Setup             sql.NullString
	Teste             sql.NullString
}

func (sr *SetupRelatorio) Insere() error {

	if sr.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	if sr.Alarme == "" {
		sr.Alarme = "S"
	}

	if sr.Arme == "" {
		sr.Arme = "S"
	}

	if sr.Desarme == "" {
		sr.Desarme = "S"
	}

	if sr.Emergencia == "" {
		sr.Emergencia = "S"
	}

	if sr.Falhas == "" {
		sr.Falhas = "S"
	}

	if sr.Geral == "" {
		sr.Geral = "S"
	}

	if sr.Medico == "" {
		sr.Medico = "S"
	}

	if sr.Panico == "" {
		sr.Panico = "S"
	}

	if sr.Restaure == "" {
		sr.Restaure = "S"
	}

	if sr.Setup == "" {
		sr.Setup = "S"
	}

	if sr.Teste == "" {
		sr.Teste = "S"
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO setupRelatorio(
			ID_SetupRelatorio, 
			ID_Cliente, 
			Alarme, 
			Arme, 
			Desarme, 
			Emergencia, 
			Falhas, 
			Geral, 
			Medico, 
			Panico, 
			Restaure, 
			Setup, 
			Teste
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		auxiliar.GeradorDeId(),
		sr.ID_Cliente,
		sr.Alarme,
		sr.Arme,
		sr.Desarme,
		sr.Emergencia,
		sr.Falhas,
		sr.Geral,
		sr.Medico,
		sr.Panico,
		sr.Restaure,
		sr.Setup,
		sr.Teste,
	); err != nil {
		return err
	}

	return nil
}

func (sr *SetupRelatorio) GetDadosById() error {
	if sr.ID_SetupRelatorio == "" {
		return errors.New("um id de setup relatório deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE setupRelatorio.ID_SetupRelatorio = '%s'
	`, sr.ID_SetupRelatorio)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, sr); err != nil {
			return err
		}

		return nil
	}

	return errors.New("setup não encontrado na base de dados")
}

func (sr *SetupRelatorio) GetDadosByIdCliente() error {

	if sr.ID_Cliente == "" {
		return errors.New("um id de cliente deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(
		`WHERE setupRelatorio.ID_Cliente = '%s'
	`, sr.ID_Cliente)
	fmt.Println(getSelect(filtro))
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return nil
	}
	defer tab.Close()

	if tab.Next() {

		if err := processaItem(tab, sr); err != nil {
			return err
		}
		fmt.Println(sr)
		return nil
	}

	return errors.New("cliente não encontrado na base de dados")
}

func (sr *SetupRelatorio) AlteraById() error {
	if sr.ID_SetupRelatorio == "" {
		return errors.New("um id de setup relátorio deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE setupRelatorio SET 
			emailEvento.Alarme = ?,
			emailEvento.Arme = ?,
			emailEvento.Desarme = ?,
			emailEvento.Emergencia = ?, 
			emailEvento.Falhas = ?,
			emailEvento.Geral = ?,
			emailEvento.Medico = ?, 
			emailEvento.Panico = ?, 
			emailEvento.Restaure = ?,
			emailEvento.Setup = ?, 
			emailEvento.Teste = ?
		WHERE setupRelatorio.ID_SetupRelatorio = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		sr.Alarme,
		sr.Arme,
		sr.Desarme,
		sr.Emergencia,
		sr.Falhas,
		sr.Geral,
		sr.Medico,
		sr.Panico,
		sr.Restaure,
		sr.Setup,
		sr.Teste,
		sr.ID_SetupRelatorio,
	); err != nil {
		return err
	}
	return nil
}

func (sr *SetupRelatorio) DeletaById() error {
	return nil
}

func (sr *SetupRelatorio) DeletaByIdCliente() error {
	return nil
}

//

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			setupRelatorio.ID_SetupRelatorio, 
			setupRelatorio.ID_Cliente, 
			setupRelatorio.Alarme, 
			setupRelatorio.Arme, 
			setupRelatorio.Desarme, 
			setupRelatorio.Emergencia, 
			setupRelatorio.Falhas, 
			setupRelatorio.Geral, 
			setupRelatorio.Medico, 
			setupRelatorio.Panico, 
			setupRelatorio.Restaure, 
			setupRelatorio.Setup, 
			setupRelatorio.Teste
			 
		FROM setupRelatorio 
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, sr *SetupRelatorio) error {

	var (
		tmp SSetupRelatorio
	)

	if err := tab.Scan(
		&tmp.ID_SetupRelatorio,
		&tmp.ID_Cliente,
		&tmp.Alarme,
		&tmp.Arme,
		&tmp.Desarme,
		&tmp.Emergencia,
		&tmp.Falhas,
		&tmp.Geral,
		&tmp.Medico,
		&tmp.Panico,
		&tmp.Restaure,
		&tmp.Setup,
		&tmp.Teste,
	); err != nil {
		return nil
	}

	*sr = SSetupRelatorioToRelatorio(tmp)

	return nil
}

func SSetupRelatorioToRelatorio(ssr SSetupRelatorio) (sr SetupRelatorio) {
	sr.ID_SetupRelatorio = ssr.ID_SetupRelatorio.String
	sr.ID_Cliente = ssr.ID_Cliente.String
	sr.Alarme = ssr.Alarme.String
	sr.Arme = ssr.Arme.String
	sr.Desarme = ssr.Desarme.String
	sr.Emergencia = ssr.Emergencia.String
	sr.Falhas = ssr.Falhas.String
	sr.Geral = ssr.Geral.String
	sr.Medico = ssr.Medico.String
	sr.Panico = ssr.Panico.String
	sr.Restaure = ssr.Restaure.String
	sr.Setup = ssr.Setup.String
	sr.Teste = ssr.Teste.String

	return
}

func GerarSetup() error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
	SELECT 
	cliente.ID_Cliente,
	cliente.Nome 
	FROM cliente 
	`)
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var id sql.NullString
		var nome sql.NullString

		if err := tab.Scan(&id, &nome); err != nil {
			return err
		}
		fmt.Println(nome.String)
		var sr SetupRelatorio

		sr.ID_Cliente = id.String
		if err := sr.Insere(); err != nil {
			return err
		}
	}
	return nil
}
