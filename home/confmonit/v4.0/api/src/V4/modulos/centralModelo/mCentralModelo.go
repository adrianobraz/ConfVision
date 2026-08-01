package centralModeloV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type CentralModelo struct {
	ID_Modelo     string `json:"idModelo"`
	ID_Fabricante string `json:"idFabricante"`
	Nome          string `json:"nome"`
	QtdParticoes  string `json:"qtdParticoes"`
	QtdZona       string `json:"qtdZona"`
	QtdPgm        string `json:"qtdPgm"`
	Eletrificador string `json:"eletrificador"`
	DataCadastro  string `json:"dataCadastro"`
	Ativo         string `json:"ativo"`
}

type SCentralModelo struct {
	ID_Modelo     sql.NullString
	ID_Fabricante sql.NullString
	Nome          sql.NullString
	QtdParticoes  sql.NullString
	QtdZona       sql.NullString
	QtdPgm        sql.NullString
	Eletrificador sql.NullString
	DataCadastro  sql.NullTime
}

func (cm *CentralModelo) Insere() error {
	if cm.ID_Modelo == "" {
		cm.ID_Modelo = auxiliar.GeradorDeId()
	}

	if cm.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	if cm.Nome == "" {
		return errors.New("um nome de central deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO modeloCentral (
				modeloCentral.ID_Modelo, 
				modeloCentral.ID_Fabricante, 
				modeloCentral.Nome, 
				modeloCentral.QtdParticoes, 
				modeloCentral.QtdZona, 
				modeloCentral.QtdPgm, 
				modeloCentral.Eletrificador
			) VALUES (?,?,?,?,?,?,?)
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		cm.ID_Modelo,
		cm.ID_Fabricante,
		cm.Nome,
		cm.QtdParticoes,
		cm.QtdZona,
		cm.QtdPgm,
		cm.Eletrificador,
	); err != nil {
		return err
	}

	return nil
}

func (cm *CentralModelo) GetDadosById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`WHERE modeloCentral.ID_Modelo = %s`, cm.ID_Modelo)
	tab, err := db.Query(cm.getSelect(filtro))

	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := cm.processaItem(tab); err != nil {
			return err
		}
		return nil

	}
	return msgNaoEncontrado()
}

func (cm *CentralModelo) AlterarById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE modeloCentral 
		SET 
			modeloCentral.Nome = ?,
			modeloCentral.QtdParticoes = ?,
			modeloCentral.QtdZona = ?,
			modeloCentral.QtdPgm = ?,
			modeloCentral.Eletrificador = ? 
		WHERE modeloCentral.ID_Modelo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		cm.Nome,
		cm.QtdParticoes,
		cm.QtdZona,
		cm.QtdPgm,
		cm.Eletrificador,
		cm.ID_Modelo,
	); err != nil {
		return err
	}
	return nil
}

func (cm *CentralModelo) DeletaById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM modeloCentral WHERE modeloCentral.ID_Modelo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(cm.ID_Modelo); err != nil {
		return err
	}

	return nil
}

func (cm *CentralModelo) DeletaAllByIdFabricante() error {
	if cm.ID_Fabricante == "" {
		return msgIdModelo()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM modeloCentral WHERE modeloCentral.ID_Fabricante = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(cm.ID_Fabricante); err != nil {
		return err
	}

	return nil
}

func (cm *CentralModelo) ListarByIdFabricante(lista *[]CentralModelo) error {
	if cm.ID_Fabricante == "" {
		return errors.New("um id de fabricante deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf("WHERE modeloCentral.ID_Fabricante = '%s'", cm.ID_Fabricante)
	tab, err := db.Query(cm.getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item CentralModelo

		if err := item.processaItem(tab); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

func (cm *CentralModelo) GetAtivoById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = cm.ID_Modelo

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	cm.Ativo = lb.Ativo

	return nil
}

func (cm *CentralModelo) SetAtivoById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	if cm.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = cm.ID_Modelo
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = cm.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (cm *CentralModelo) InvereteAtivoById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = cm.ID_Modelo
	lb.Descricao = "ALTERADO VIA WEB"
	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	cm.Ativo = lb.Ativo

	return nil
}

func (cm *CentralModelo) GetEletrificadorById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT modeloCentral.Eletrificador FROM modeloCentral WHERE modeloCentral.ID_Modelo = ?
	`, cm.ID_Modelo)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&cm.Eletrificador); err != nil {
			return err
		}

		return nil
	}
	return msgNaoEncontrado()
}

func (cm *CentralModelo) SetEletrificadorById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	if cm.Eletrificador == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE modeloCentral 
		SET modeloCentral.Eletrificador = ? 
		WHERE modeloCentral.ID_Modelo = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(cm.Eletrificador, cm.ID_Modelo); err != nil {
		return err
	}
	return nil
}

func (cm *CentralModelo) InvereteEletrificadorById() error {
	if cm.ID_Modelo == "" {
		return msgIdModelo()
	}

	if err := cm.GetEletrificadorById(); err != nil {
		return err
	}

	auxiliar.InverteEstado(&cm.Eletrificador)

	if err := cm.SetEletrificadorById(); err != nil {
		return err
	}

	return nil
}

// Funcoes internas ===========================================================

func (cm *CentralModelo) sCentralModeloToCentralModelo(scm SCentralModelo) {
	cm.ID_Modelo = scm.ID_Modelo.String
	cm.ID_Fabricante = scm.ID_Fabricante.String
	cm.Nome = scm.Nome.String
	cm.QtdParticoes = scm.QtdParticoes.String
	cm.QtdZona = scm.QtdZona.String
	cm.QtdPgm = scm.QtdPgm.String
	cm.Eletrificador = scm.Eletrificador.String
	cm.DataCadastro = scm.DataCadastro.Time.Format("02/01/2006 15:04:05")
}

func (cm *CentralModelo) getSelect(filtro string) string {

	return fmt.Sprintf(`
		SELECT
			modeloCentral.ID_Modelo, 
			modeloCentral.ID_Fabricante, 
			modeloCentral.Nome, 
			modeloCentral.QtdParticoes, 
			modeloCentral.QtdZona, 
			modeloCentral.QtdPgm, 
			modeloCentral.Eletrificador, 
			modeloCentral.DataCadastro,

			listaBloqueio.ID_Alvo

		FROM modeloCentral
		
		LEFT JOIN listaBloqueio 
		ON modeloCentral.ID_Modelo = listaBloqueio.ID_Alvo
		%s
	`, filtro)
}

func (cm *CentralModelo) processaItem(tab *sql.Rows) error {
	var (
		tmp               SCentralModelo
		bloqCentralModelo sql.NullString
	)
	if err := tab.Scan(
		&tmp.ID_Modelo,
		&tmp.ID_Fabricante,
		&tmp.Nome,
		&tmp.QtdParticoes,
		&tmp.QtdZona,
		&tmp.QtdPgm,
		&tmp.Eletrificador,
		&tmp.DataCadastro,

		&bloqCentralModelo,
	); err != nil {
		return err
	}

	cm.sCentralModeloToCentralModelo(tmp)

	if bloqCentralModelo.Valid {
		cm.Ativo = "N"
	} else {
		cm.Ativo = "S"
	}

	return nil
}

func msgNaoEncontrado() error {
	return errors.New("não encontrado na base de dados")
}

func msgIdModelo() error {
	return errors.New("um id de modelo deve ser informado")
}
