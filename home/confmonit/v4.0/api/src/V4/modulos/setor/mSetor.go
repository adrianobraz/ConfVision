package setorV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
)

type Setor struct {
	ID_Setor       string `json:"idSetor"`
	ID_Dispositivo string `json:"idDispositivo"`
	Nome           string `json:"nome"`
	Descricao      string `json:"descricao"`
	Tipo           string `json:"tipo"`
	Camera         string `json:"camera"`
	Particao       string `json:"particao"`
	Numero         string `json:"numero"`
	Ativo          string `json:"ativo"`
}

type SSetor struct {
	ID_Setor       sql.NullString
	ID_Dispositivo sql.NullString
	Nome           sql.NullString
	Descricao      sql.NullString
	Tipo           sql.NullString
	Camera         sql.NullString
	Particao       sql.NullString
	Numero         sql.NullString
}

func (s *Setor) Insere() error {
	if s.ID_Setor == "" {
		s.ID_Setor = auxiliar.GeradorDeId()
	}

	if s.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	if s.Nome == "" {
		return errors.New("um nome de setor deve ser informado")
	}

	if s.Numero == "" {
		return errors.New("um numero de setor deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO setorAlarme(
			setorAlarme.ID_Setor, 
			setorAlarme.ID_Dispositivo, 
			setorAlarme.Nome, 
			setorAlarme.Descricao, 
			setorAlarme.Tipo, 
			setorAlarme.Camera, 
			setorAlarme.Particao, 
			setorAlarme.Numero
		) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? ) 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		s.ID_Setor,
		s.ID_Dispositivo,
		s.Nome,
		s.Descricao,
		s.Tipo,
		s.Camera,
		s.Particao,
		s.Numero,
	); err != nil {
		return err
	}

	return nil
}

func (s *Setor) GetDadosById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(` 
		WHERE setorAlarme.ID_Setor = '%s'
	`, s.ID_Setor)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, s); err != nil {
			return err
		}
		return nil
	}
	return errors.New("setor não encontrado na base de dados")
}

func (s *Setor) AlteraById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE setorAlarme SET 
			setorAlarme.Nome = ?,
			setorAlarme.Descricao = ?,
			setorAlarme.Tipo = ?,
			setorAlarme.Camera = ?,
			setorAlarme.Particao = ?,
			setorAlarme.Numero = ?
		WHERE setorAlarme.ID_Setor = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		s.Nome,
		s.Descricao,
		s.Tipo,
		s.Camera,
		s.Particao,
		s.Numero,
		s.ID_Setor,
	); err != nil {
		return err
	}

	return nil
}

func (s *Setor) DeletaById() error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM setorAlarme WHERE setorAlarme.ID_Setor = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(s.ID_Setor); err != nil {
		return err
	}

	return nil
}

func (s *Setor) DeletaAllByDispositivo() error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM setorAlarme WHERE setorAlarme.ID_Dispositivo = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(s.ID_Dispositivo); err != nil {
		return err
	}

	return nil
}

// Manipula campo camera ======================================================
func (s *Setor) GetCameraAtivaById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT setorAlarme.Camera 
		FROM setorAlarme 
		WHERE setorAlarme.ID_Setor = ?
	`, s.ID_Setor)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := tab.Scan(&s.Camera); err != nil {
			return err
		}
		return err
	}
	return nil
}

func (s *Setor) SetCameraAtivaById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}

	if s.Camera == "" {
		return errors.New("um status de camera deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE setorAlarme 
		SET setorAlarme.Camera = ?
		WHERE setorAlarme.ID_Setor = ?   
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		s.Camera,
		s.ID_Setor,
	); err != nil {
		return err
	}
	return nil
}

func (s *Setor) InverteCameraAtivaById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}

	if err := s.GetCameraAtivaById(); err != nil {
		return err
	}

	if err := auxiliar.InverteEstado(&s.Camera); err != nil {
		return err
	}

	if err := s.SetCameraAtivaById(); err != nil {
		return err
	}

	return nil
}

// Manipula sertor ativo ======================================================
func (s *Setor) GetSetorAtivaById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = s.ID_Setor

	if err := lb.GetBloqueado(); err != nil {
		return err
	}

	s.Ativo = lb.Ativo

	return nil
}

func (s *Setor) SetSetorAtivaById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}

	if s.Ativo == "" {
		return errors.New("um estado de ativo deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = s.ID_Setor
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = s.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (s *Setor) InverteSetorAtivaById() error {
	if s.ID_Setor == "" {
		return errors.New("um id de setor deve ser informado")
	}
	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = s.ID_Setor
	lb.Descricao = "ALTERADO VIA WEB"
	if err := lb.InverteBloqueado(); err != nil {
		return err
	}

	s.Ativo = lb.Ativo
	return nil
}

func (s *Setor) ListaByIdDispositivo(lista *[]Setor) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE setorAlarme.ID_Dispositivo = '%s'
	`, s.ID_Dispositivo)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Setor
		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}
	return nil
}

// funcoes internas ===========================================================

func sSetorToSetor(ss SSetor) (s Setor) {

	s.ID_Setor = ss.ID_Setor.String
	s.ID_Dispositivo = ss.ID_Dispositivo.String
	s.Nome = ss.Nome.String
	s.Descricao = ss.Descricao.String
	s.Tipo = ss.Tipo.String
	s.Camera = ss.Camera.String
	s.Particao = ss.Particao.String
	s.Numero = ss.Numero.String
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			setorAlarme.ID_Setor, 
			setorAlarme.ID_Dispositivo, 
			setorAlarme.Nome, 
			setorAlarme.Descricao, 
			setorAlarme.Tipo, 
			setorAlarme.Camera, 
			setorAlarme.Particao, 
			setorAlarme.Numero, 

			listaBloqueio.ID_Alvo
			
		FROM setorAlarme 
		
		LEFT JOIN listaBloqueio 
		ON  setorAlarme.ID_Setor = listaBloqueio.ID_Alvo

		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, s *Setor) error {
	var (
		tmp       SSetor
		bloqSetor sql.NullString
	)

	if err := tab.Scan(
		&tmp.ID_Setor,
		&tmp.ID_Dispositivo,
		&tmp.Nome,
		&tmp.Descricao,
		&tmp.Tipo,
		&tmp.Camera,
		&tmp.Particao,
		&tmp.Numero,

		&bloqSetor,
	); err != nil {
		return err
	}

	*s = sSetorToSetor(tmp)

	if bloqSetor.Valid {
		s.Ativo = "N"
	} else {
		s.Ativo = "S"
	}
	return nil
}
