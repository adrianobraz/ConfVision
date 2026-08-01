package dispositivoV4

import (
	connV4 "api/src/V4/conexao"
	paginacaoV4 "api/src/V4/paginacaoV4"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Faixa no corpo da requisicao: "1","3","6","12","24" ou "todos" (>=24h).
// Campo adicionado em Dispositivo via json:"faixa" no struct principal.

type ContagemSemComunicacaoFaixas struct {
	H1    int `json:"1"`
	H3    int `json:"3"`
	H6    int `json:"6"`
	H12   int `json:"12"`
	H24   int `json:"24"`
	Todos int `json:"todos"`
}

func limitesFaixaSemComunicacao(faixa string) (minH, maxH int, err error) {
	switch faixa {
	case "1":
		return 1, 3, nil
	case "3":
		return 3, 6, nil
	case "6":
		return 6, 12, nil
	case "12":
		return 12, 24, nil
	case "24":
		return 24, 36, nil
	default:
		return 0, 0, errors.New("faixa invalida")
	}
}

func filtroFaixaSemComunicacao(idFranqueado string, horasMin, horasMax int) string {
	agora := time.Now()
	limiteMin := agora.Add(-time.Duration(horasMin) * time.Hour)
	limiteMax := agora.Add(-time.Duration(horasMax) * time.Hour)

	return fmt.Sprintf(`
		WHERE dispositivo.ID_Cliente IN (
			SELECT cliente.ID_Cliente 
			FROM cliente
			WHERE cliente.ID_Franqueado = '%s'  
		) 
		AND dispositivo.ID_Dispositivo NOT IN (
			SELECT listaBloqueio.ID_Alvo 
			FROM listaBloqueio
		)
		AND dispositivo.DataUltimoEvento <= '%s'
		AND dispositivo.DataUltimoEvento > '%s'
	`, idFranqueado, limiteMin.Format("2006-01-02 15:04:05"), limiteMax.Format("2006-01-02 15:04:05"))
}

func (d *Dispositivo) ListarSemComunicacaoByFaixa(lista *[]Dispositivo) (int, error) {
	if d.ID_Franqueado == "" {
		return 0, errors.New("um id de franqueado deve ser informado")
	}

	faixa := d.Faixa
	if faixa == "" {
		faixa = d.HorasToFaixa()
	}
	if faixa == "" {
		return 0, errors.New("faixa deve ser informada")
	}

	if faixa == "todos" {
		return d.ListarSemComunicacaoByIdFranqueado(lista)
	}

	minH, maxH, err := limitesFaixaSemComunicacao(faixa)
	if err != nil {
		return 0, err
	}

	db, err := connV4.Conectar()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	filtro := filtroFaixaSemComunicacao(d.ID_Franqueado, minH, maxH)
	filtro = d.appendFiltroTermoDispositivo(filtro)

	total := 0
	if d.Limit > 0 {
		total, err = d.contarDispositivos(db, filtro)
		if err != nil {
			return 0, err
		}
	}

	filtro += `
		ORDER BY dispositivo.DataUltimoEvento DESC
	`
	filtro += paginacaoV4.Clausula(d.Limit, d.Offset)

	tab, err := db.Query(getSelect(filtro))
	if err != nil {
		return 0, err
	}
	defer tab.Close()

	for tab.Next() {
		var item Dispositivo
		if err := processaItem(tab, &item); err != nil {
			return 0, err
		}
		*lista = append(*lista, item)
	}

	if d.Limit <= 0 {
		total = len(*lista)
	}
	return total, nil
}

func (d *Dispositivo) HorasToFaixa() string {
	switch d.Horas {
	case 1, 3, 6, 12, 24:
		return fmt.Sprintf("%d", d.Horas)
	default:
		return ""
	}
}

func (d *Dispositivo) ContarSemComunicacaoFaixas(contagem *ContagemSemComunicacaoFaixas) error {
	if d.ID_Franqueado == "" {
		return errors.New("um id de franqueado deve ser informado")
	}

	db, err := connV4.Conectar()
	if err != nil {
		return err
	}
	defer db.Close()

	agora := time.Now()
	t1 := agora.Add(-1 * time.Hour).Format("2006-01-02 15:04:05")
	t3 := agora.Add(-3 * time.Hour).Format("2006-01-02 15:04:05")
	t6 := agora.Add(-6 * time.Hour).Format("2006-01-02 15:04:05")
	t12 := agora.Add(-12 * time.Hour).Format("2006-01-02 15:04:05")
	t24 := agora.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	t36 := agora.Add(-36 * time.Hour).Format("2006-01-02 15:04:05")

	sqlContagem := fmt.Sprintf(`
		SELECT 
			SUM(CASE WHEN dispositivo.DataUltimoEvento <= '%s' AND dispositivo.DataUltimoEvento > '%s' THEN 1 ELSE 0 END),
			SUM(CASE WHEN dispositivo.DataUltimoEvento <= '%s' AND dispositivo.DataUltimoEvento > '%s' THEN 1 ELSE 0 END),
			SUM(CASE WHEN dispositivo.DataUltimoEvento <= '%s' AND dispositivo.DataUltimoEvento > '%s' THEN 1 ELSE 0 END),
			SUM(CASE WHEN dispositivo.DataUltimoEvento <= '%s' AND dispositivo.DataUltimoEvento > '%s' THEN 1 ELSE 0 END),
			SUM(CASE WHEN dispositivo.DataUltimoEvento <= '%s' AND dispositivo.DataUltimoEvento > '%s' THEN 1 ELSE 0 END),
			SUM(CASE WHEN dispositivo.DataUltimoEvento < '%s' THEN 1 ELSE 0 END)
		FROM dispositivo
		%s
	`, t1, t3, t3, t6, t6, t12, t12, t24, t24, t36, t24, whereSemComunicacaoFranqueado(d.ID_Franqueado))

	var c1, c3, c6, c12, c24, cTodos sql.NullInt64
	if err := db.QueryRow(sqlContagem).Scan(&c1, &c3, &c6, &c12, &c24, &cTodos); err != nil {
		return err
	}

	contagem.H1 = int(c1.Int64)
	contagem.H3 = int(c3.Int64)
	contagem.H6 = int(c6.Int64)
	contagem.H12 = int(c12.Int64)
	contagem.H24 = int(c24.Int64)
	contagem.Todos = int(cTodos.Int64)

	return nil
}
