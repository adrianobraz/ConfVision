package gradeV4

import (
	connV4 "api/src/V4/conexao"
	listaBoqueioV4 "api/src/V4/modulos/listaBoqueio"
	"api/src/auxiliar"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Grade struct {
	ID_Grade       string `json:"idGrade"`
	ID_Dispositivo string `json:"idDispositivo"`
	Nome           string `json:"nome"`
	SegEam         string `json:"segEam"`
	SegSam         string `json:"segSam"`
	SegEpm         string `json:"segEpm"`
	SegSpm         string `json:"segSpm"`
	TerEam         string `json:"terEam"`
	TerSam         string `json:"terSam"`
	TerEpm         string `json:"terEpm"`
	TerSpm         string `json:"terSpm"`
	QuaEam         string `json:"quaEam"`
	QuaSam         string `json:"quaSam"`
	QuaEpm         string `json:"quaEpm"`
	QuaSpm         string `json:"quaSpm"`
	QuiEam         string `json:"quiEam"`
	QuiSam         string `json:"quiSam"`
	QuiEpm         string `json:"quiEpm"`
	QuiSpm         string `json:"quiSpm"`
	SexEam         string `json:"sexEam"`
	SexSam         string `json:"sexSam"`
	SexEpm         string `json:"sexEpm"`
	SexSpm         string `json:"sexSpm"`
	SabEam         string `json:"sabEam"`
	SabSam         string `json:"sabSam"`
	SabEpm         string `json:"sabEpm"`
	SabSpm         string `json:"sabSpm"`
	DomEam         string `json:"domEam"`
	DomSam         string `json:"domSam"`
	DomEpm         string `json:"domEpm"`
	DomSpm         string `json:"domSpm"`
	Ativo          string `json:"ativo"`
	Tolerancia     string `json:"tolerancia"`
	DataCadastro   string `json:"dataCadastro"`
}

type SGrade struct {
	ID_Grade       sql.NullString
	ID_Dispositivo sql.NullString
	Nome           sql.NullString
	SegEam         sql.NullString
	SegSam         sql.NullString
	SegEpm         sql.NullString
	SegSpm         sql.NullString
	TerEam         sql.NullString
	TerSam         sql.NullString
	TerEpm         sql.NullString
	TerSpm         sql.NullString
	QuaEam         sql.NullString
	QuaSam         sql.NullString
	QuaEpm         sql.NullString
	QuaSpm         sql.NullString
	QuiEam         sql.NullString
	QuiSam         sql.NullString
	QuiEpm         sql.NullString
	QuiSpm         sql.NullString
	SexEam         sql.NullString
	SexSam         sql.NullString
	SexEpm         sql.NullString
	SexSpm         sql.NullString
	SabEam         sql.NullString
	SabSam         sql.NullString
	SabEpm         sql.NullString
	SabSpm         sql.NullString
	DomEam         sql.NullString
	DomSam         sql.NullString
	DomEpm         sql.NullString
	DomSpm         sql.NullString
	Tolerancia     sql.NullString
	DataCadastro   sql.NullTime
}

func (g *Grade) Insere() error {
	if g.ID_Grade == "" {
		g.ID_Grade = auxiliar.GeradorDeId()
	}

	if g.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		INSERT INTO grade(
			grade.ID_Grade, 
			grade.ID_Dispositivo, 
			grade.Nome, 
			grade.SegEam, 
			grade.SegSam, 
			grade.SegEpm, 
			grade.SegSpm, 
			grade.TerEam, 
			grade.TerSam, 
			grade.TerEpm, 
			grade.TerSpm, 
			grade.QuaEam, 
			grade.QuaSam, 
			grade.QuaEpm, 
			grade.QuaSpm, 
			grade.QuiEam, 
			grade.QuiSam, 
			grade.QuiEpm, 
			grade.QuiSpm, 
			grade.SexEam, 
			grade.SexSam, 
			grade.SexEpm, 
			grade.SexSpm, 
			grade.SabEam, 
			grade.SabSam, 
			grade.SabEpm, 
			grade.SabSpm, 
			grade.DomEam, 
			grade.DomSam, 
			grade.DomEpm, 
			grade.DomSpm, 
			grade.Tolerancia
		) VALUES ( ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,? )	
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		g.ID_Grade,
		g.ID_Dispositivo,
		g.Nome,
		g.SegEam,
		g.SegSam,
		g.SegEpm,
		g.SegSpm,
		g.TerEam,
		g.TerSam,
		g.TerEpm,
		g.TerSpm,
		g.QuaEam,
		g.QuaSam,
		g.QuaEpm,
		g.QuaSpm,
		g.QuiEam,
		g.QuiSam,
		g.QuiEpm,
		g.QuiSpm,
		g.SexEam,
		g.SexSam,
		g.SexEpm,
		g.SexSpm,
		g.SabEam,
		g.SabSam,
		g.SabEpm,
		g.SabSpm,
		g.DomEam,
		g.DomSam,
		g.DomEpm,
		g.DomSpm,
		g.Tolerancia,
	); err != nil {
		return err
	}

	return nil
}

func (g *Grade) GetDadosById() error {
	if g.ID_Grade == "" {
		return errors.New("um id de grade deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE grade.ID_Grade = '%s'
	`, g.ID_Grade)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		if err := processaItem(tab, g); err != nil {
			return err
		}
		return nil
	}

	return errors.New("grade não encontrado na base de dados")
}

func (g *Grade) AlteraById() error {
	if g.ID_Grade == "" {
		return errors.New("um id de grade deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		UPDATE grade SET 
			grade.Nome = ?,
			grade.SegEam = ?,
			grade.SegSam = ?,
			grade.SegEpm = ?,
			grade.SegSpm = ?,
			grade.TerEam = ?,
			grade.TerSam = ?,
			grade.TerEpm = ?,
			grade.TerSpm = ?,
			grade.QuaEam = ?,
			grade.QuaSam = ?,
			grade.QuaEpm = ?,
			grade.QuaSpm = ?,
			grade.QuiEam = ?,
			grade.QuiSam = ?,
			grade.QuiEpm = ?,
			grade.QuiSpm = ?,
			grade.SexEam = ?,
			grade.SexSam = ?,
			grade.SexEpm = ?,
			grade.SexSpm = ?,
			grade.SabEam = ?,
			grade.SabSam = ?,
			grade.SabEpm = ?,
			grade.SabSpm = ?,
			grade.DomEam = ?,
			grade.DomSam = ?,
			grade.DomEpm = ?,
			grade.DomSpm = ?,
			grade.Tolerancia = ?

		WHERE grade.ID_Grade = ?
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(
		g.Nome,
		g.SegEam,
		g.SegSam,
		g.SegEpm,
		g.SegSpm,
		g.TerEam,
		g.TerSam,
		g.TerEpm,
		g.TerSpm,
		g.QuaEam,
		g.QuaSam,
		g.QuaEpm,
		g.QuaSpm,
		g.QuiEam,
		g.QuiSam,
		g.QuiEpm,
		g.QuiSpm,
		g.SexEam,
		g.SexSam,
		g.SexEpm,
		g.SexSpm,
		g.SabEam,
		g.SabSam,
		g.SabEpm,
		g.SabSpm,
		g.DomEam,
		g.DomSam,
		g.DomEpm,
		g.DomSpm,
		g.Tolerancia,
		g.ID_Grade,
	); err != nil {
		return err
	}

	return nil
}

func (g *Grade) DeletaById() error {
	if g.ID_Grade == "" {
		return errors.New("um id de grade deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM grade WHERE grade.ID_Grade = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()

	if _, err := stm.Exec(g.ID_Grade); err != nil {
		return err
	}

	return nil
}

func (g *Grade) DeletaAllByIdDispositivo() error {
	if g.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	stm, err := db.Prepare(`
		DELETE FROM grade WHERE grade.ID_Dispositivo = ? 
	`)
	if err != nil {
		return err
	}
	defer stm.Close()
	
	if _, err := stm.Exec(g.ID_Dispositivo); err != nil {
		return err
	}
	return nil
}

func (g *Grade) ListarByIdDispositivo(lista *[]Grade) error {
	if g.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	filtro := fmt.Sprintf(`
		WHERE grade.ID_Dispositivo = '%s'
	`, g.ID_Dispositivo)
	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return err
	}
	defer tab.Close()

	for tab.Next() {
		var item Grade

		if err := processaItem(tab, &item); err != nil {
			return err
		}

		*lista = append(*lista, item)
	}

	return nil
}

// Manipula campo Ativo =======================================================
func (g *Grade) GetAtivo() error {

	if g.ID_Grade == "" {
		return errors.New("um id de grade deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT listaBloqueio.ID_Alvo
		FROM listaBloqueio 		
		WHERE listaBloqueio.ID_Alvo = ?
	`, g.ID_Grade)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		g.Ativo = "N"
	} else {
		g.Ativo = "S"
	}
	return nil
}

func (g *Grade) SetAtivo() error {
	if g.ID_Grade == "" {
		return errors.New("um id de grade deve ser informado")
	}

	var lb listaBoqueioV4.ListaBloqueio
	lb.ID_Alvo = g.ID_Grade
	lb.DataRetirada = ""
	lb.Descricao = "ALTERADO VIA WEB"
	lb.Ativo = g.Ativo

	if err := lb.SetBloqueado(); err != nil {
		return err
	}

	return nil
}

func (g *Grade) InverteAtivo() error {
	if g.ID_Grade == "" {
		return errors.New("um id de grade deve ser informado")
	}
	if err := g.GetAtivo(); err != nil {
		return err
	}

	if err := getIdDispByIDGrade(g.ID_Grade, &g.ID_Dispositivo); err != nil {
		return err
	}
	
	if err := g.DesabiltaAllAtivoByIdDispositivo(); err != nil {
		fmt.Println(err)
		return err
	}

	if err := auxiliar.InverteEstado(&g.Ativo); err != nil {
		return err
	}

	if err := g.SetAtivo(); err != nil {
		return err
	}

	return nil
}

func (g *Grade) DesabiltaAllAtivoByIdDispositivo() error {
	if g.ID_Dispositivo == "" {
		return errors.New("um id de dispositivo deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	// Busco todas as grades do dispositivo
	tab, err := db.Query(`
		SELECT grade.ID_Grade 
		FROM grade 
		WHERE grade.ID_Dispositivo = ?
	`, g.ID_Dispositivo)
	if err != nil {
		return err
	}
	defer tab.Close()

	// Desabilita todas as grade
	for tab.Next() {
		var id sql.NullString

		if err := tab.Scan(&id); err != nil {
			return err
		}

		var lb listaBoqueioV4.ListaBloqueio
		lb.ID_Alvo = id.String
		lb.DataRetirada = ""
		lb.Descricao = "ALTERADO VIA WEB"
		lb.Ativo = "N"

		if err := lb.SetBloqueado(); err != nil {
			return err
		}
	}

	return nil
}

// Verifica se existe uma grade e se o evento esta fora do horario caso exista
func (g *Grade) DesarmeForaHorarioByIdCliente(codigo *string) error {

	var txt string
	switch int(time.Now().Weekday()) {
	case 0:
		txt = `
			SELECT 
				DomEam, 
				DomSam, 
				DomEpm, 
				DomSpm, 
				Tolerancia
			FROM grade 
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	case 1:
		txt = `
			SELECT 
				SegEam, 
				SegSam, 
				SegEpm, 
				SegSpm, 
				Tolerancia
			FROM grade
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	case 2:
		txt = `
			SELECT 
				TerEam, 
				TerSam, 
				TerEpm, 
				TerSpm, 
				Tolerancia
			FROM grade 
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	case 3:
		txt = `
			SELECT 
				QuaEam, 
				QuaSam, 
				QuaEpm, 
				QuaSpm , 
				Tolerancia
			FROM grade 
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	case 4:
		txt = `
			SELECT 
				QuiEam, 
				QuiSam, 
				QuiEpm, 
				QuiSpm, 
				Tolerancia
			FROM grade 
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	case 5:
		txt = `
			SELECT 
				SexEam, 
				SexSam, 
				SexEpm, 
				SexSpm, 
				Tolerancia
			FROM grade 
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	case 6:
		txt = `
			SELECT 
				SabEam, 
				SabSam, 
				SabEpm, 
				SabSpm, 
				Tolerancia
			FROM grade 
			WHERE ID_Cliente = ? AND Ativo = 'S'`
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, erro := db.Query(txt, g.ID_Grade)
	if erro != nil {
		return erro
	}
	defer tab.Close()

	// Caso temha uma grade para processar
	if tab.Next() {
		var desarme, armeIntervalo, desarmeIntervalo, arme, tol string
		if erro := tab.Scan(
			&desarme,
			&armeIntervalo,
			&desarmeIntervalo,
			&arme,
			&tol,
		); erro != nil {
			return erro
		}

		//
		//
		// Verifica se o evento esta fora do horario caso o cliente tenha uma grade
		// de horario e casso a variavel Alarme esteja seta em true e caso esteja ele
		// atera o evento para 1M01

		foraEntrada := false
		foraIntervalo := false
		foraSaida := false
		agora := time.Now()

		if desarme != "" && desarme != "00:00" {
			var entradaAm time.Time
			if erro := g.calculaGrade(desarme, tol, "SUB", &entradaAm); erro != nil {
				return erro
			}
			if agora.Before(entradaAm) {
				foraEntrada = true
			}
		}

		if armeIntervalo != "" && armeIntervalo != "00:00" &&
			desarmeIntervalo != "" && desarmeIntervalo != "00:00" {
			var saidaAm time.Time
			if erro := g.calculaGrade(armeIntervalo, tol, "ADD", &saidaAm); erro != nil {
				return erro
			}

			var entradaPm time.Time
			if erro := g.calculaGrade(desarmeIntervalo, tol, "SUB", &entradaPm); erro != nil {
				return erro
			}

			if agora.After(saidaAm) && agora.Before(entradaPm) {
				foraIntervalo = true
			}
		}

		if arme != "" && arme != "00:00" {
			var saidaPm time.Time
			if erro := g.calculaGrade(arme, tol, "ADD", &saidaPm); erro != nil {
				return erro
			}

			if agora.After(saidaPm) {
				foraSaida = true
			}
		}

		if foraEntrada || foraIntervalo || foraSaida {
			*codigo = "1M01"
			return nil
		}
	}

	return nil
}

// funções internas ===========================================================
// Auxilia a funcao verificaGrade
func (g *Grade) calculaGrade(inTime, tol, tipo string, outTime *time.Time) error {

	// Converte tolerancia para duração
	tolerancia, erro := time.ParseDuration(tol + "m")
	if erro != nil {
		return erro
	}

	var temp = strings.Split(inTime, ":")
	// Hora da grade
	hora, erro := strconv.Atoi(temp[0])
	if erro != nil {
		return erro
	}

	minuto, erro := strconv.Atoi(temp[1])
	if erro != nil {
		return erro
	}

	t := time.Now()
	gradeHora := time.Date(
		t.Year(), t.Month(), t.Day(),
		hora, minuto, t.Second(),
		t.Nanosecond(), t.Location(),
	)
	if tipo == "ADD" {
		*outTime = gradeHora.Add(tolerancia)
	} else {
		*outTime = gradeHora.Add(-tolerancia)
	}

	return nil
}

func sGradeToGrade(sg SGrade) (s Grade) {
	s.ID_Grade = sg.ID_Grade.String
	s.ID_Dispositivo = sg.ID_Dispositivo.String
	s.Nome = sg.Nome.String
	s.SegEam = sg.SegEam.String
	s.SegSam = sg.SegSam.String
	s.SegEpm = sg.SegEpm.String
	s.SegSpm = sg.SegSpm.String
	s.TerEam = sg.TerEam.String
	s.TerSam = sg.TerSam.String
	s.TerEpm = sg.TerEpm.String
	s.TerSpm = sg.TerSpm.String
	s.QuaEam = sg.QuaEam.String
	s.QuaSam = sg.QuaSam.String
	s.QuaEpm = sg.QuaEpm.String
	s.QuaSpm = sg.QuaSpm.String
	s.QuiEam = sg.QuiEam.String
	s.QuiSam = sg.QuiSam.String
	s.QuiEpm = sg.QuiEpm.String
	s.QuiSpm = sg.QuiSpm.String
	s.SexEam = sg.SexEam.String
	s.SexSam = sg.SexSam.String
	s.SexEpm = sg.SexEpm.String
	s.SexSpm = sg.SexSpm.String
	s.SabEam = sg.SabEam.String
	s.SabSam = sg.SabSam.String
	s.SabEpm = sg.SabEpm.String
	s.SabSpm = sg.SabSpm.String
	s.DomEam = sg.DomEam.String
	s.DomSam = sg.DomSam.String
	s.DomEpm = sg.DomEpm.String
	s.DomSpm = sg.DomSpm.String
	s.Tolerancia = sg.Tolerancia.String
	s.DataCadastro = sg.DataCadastro.Time.Format("02/01/2006 15:04:05")
	return
}

func getSelect(filtro string) string {
	return fmt.Sprintf(`
		SELECT 
			grade.ID_Grade, 
			grade.ID_Dispositivo, 
			grade.Nome, 
			grade.SegEam, 
			grade.SegSam, 
			grade.SegEpm, 
			grade.SegSpm, 
			grade.TerEam, 
			grade.TerSam, 
			grade.TerEpm, 
			grade.TerSpm, 
			grade.QuaEam, 
			grade.QuaSam, 
			grade.QuaEpm, 
			grade.QuaSpm, 
			grade.QuiEam, 
			grade.QuiSam, 
			grade.QuiEpm, 
			grade.QuiSpm, 
			grade.SexEam, 
			grade.SexSam, 
			grade.SexEpm, 
			grade.SexSpm, 
			grade.SabEam, 
			grade.SabSam, 
			grade.SabEpm, 
			grade.SabSpm, 
			grade.DomEam, 
			grade.DomSam, 
			grade.DomEpm, 
			grade.DomSpm, 
			grade.Tolerancia, 
			grade.DataCadastro,
			
			listaBloqueio.ID_Alvo
		FROM grade

		LEFT JOIN listaBloqueio 
		ON grade.ID_Grade = listaBloqueio.ID_Alvo
		%s
	`, filtro)
}

func processaItem(tab *sql.Rows, g *Grade) error {
	var (
		tmp   SGrade
		ativo sql.NullString
	)

	if err := tab.Scan(
		&tmp.ID_Grade,
		&tmp.ID_Dispositivo,
		&tmp.Nome,
		&tmp.SegEam,
		&tmp.SegSam,
		&tmp.SegEpm,
		&tmp.SegSpm,
		&tmp.TerEam,
		&tmp.TerSam,
		&tmp.TerEpm,
		&tmp.TerSpm,
		&tmp.QuaEam,
		&tmp.QuaSam,
		&tmp.QuaEpm,
		&tmp.QuaSpm,
		&tmp.QuiEam,
		&tmp.QuiSam,
		&tmp.QuiEpm,
		&tmp.QuiSpm,
		&tmp.SexEam,
		&tmp.SexSam,
		&tmp.SexEpm,
		&tmp.SexSpm,
		&tmp.SabEam,
		&tmp.SabSam,
		&tmp.SabEpm,
		&tmp.SabSpm,
		&tmp.DomEam,
		&tmp.DomSam,
		&tmp.DomEpm,
		&tmp.DomSpm,
		&tmp.Tolerancia,
		&tmp.DataCadastro,
		&ativo,
	); err != nil {
		return err
	}

	*g = sGradeToGrade(tmp)

	if ativo.Valid {
		g.Ativo = "N"
	} else {
		g.Ativo = "S"
	}
	return nil
}

func getIdDispByIDGrade(idGrade string, idDisp *string) error {
	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	tab, err := db.Query(`
		SELECT grade.ID_Dispositivo
		FROM grade		
		WHERE grade.ID_Grade = ?
	`, idGrade)
	if err != nil {
		return err
	}
	defer tab.Close()

	if tab.Next() {
		var id sql.NullString
		if err := tab.Scan(&id); err != nil {
			return err
		}

		*idDisp = id.String
		return nil
	}
	return errors.New("não encontrado")
}
