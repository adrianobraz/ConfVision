package modGrade

import (
	"encoding/json"
	"io"
	"net/http"
	aux "terminal/src/auxiliar"
	"terminal/src/tipos"
)

var Rotas = []tipos.Rota{
	{
		Uri:    "/modGrade/buscarDados",
		Metodo: http.MethodPost,
		Funcao: buscarDados,
	},
}

func buscarDados(w http.ResponseWriter, r *http.Request) {
	type objeto struct {
		IdDispositivo string `json:"idDispositivo"`
		NomeDisp      string `json:"nomeDisp"`
		Nome          string `json:"nome"`

		DomEam string `json:"domEam"`
		DomSam string `json:"domSam"`
		DomEpm string `json:"domEpm"`
		DomSpm string `json:"domSpm"`

		SegEam string `json:"segEam"`
		SegSam string `json:"segSam"`
		SegEpm string `json:"segEpm"`
		SegSpm string `json:"segSpm"`

		TerEam string `json:"terEam"`
		TerSam string `json:"terSam"`
		TerEpm string `json:"terEpm"`
		TerSpm string `json:"terSpm"`

		QuaEam string `json:"quaEam"`
		QuaSam string `json:"quaSam"`
		QuaEpm string `json:"quaEpm"`
		QuaSpm string `json:"quaSpm"`

		QuiEam string `json:"quiEam"`
		QuiSam string `json:"quiSam"`
		QuiEpm string `json:"quiEpm"`
		QuiSpm string `json:"quiSpm"`

		SexEam string `json:"sexEam"`
		SexSam string `json:"sexSam"`
		SexEpm string `json:"sexEpm"`
		SexSpm string `json:"sexSpm"`

		SabEam string `json:"sabEam"`
		SabSam string `json:"sabSam"`
		SabEpm string `json:"sabEpm"`
		SabSpm string `json:"sabSpm"`

		Tol string `json:"tolerancia"`
	}

	body, erro := io.ReadAll(r.Body)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	var obj objeto
	if erro := json.Unmarshal(body, &obj); erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}

	db, erro := aux.Conectar()
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer db.Close()

	tab, erro := db.Query(`
		SELECT
			grade.Nome,

			grade.DomEam,
			grade.DomSam,
			grade.DomEpm,
			grade.DomSpm,

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
			
			grade.Tolerancia,

			dispositivo.Nome

		FROM grade 

		LEFT JOIN dispositivo
		ON grade.ID_Dispositivo = dispositivo.ID_Dispositivo
		
		WHERE  grade.ID_Dispositivo = ?
		AND grade.ID_Grade NOT IN(
			SELECT listaBloqueio.ID_Alvo
			FROM listaBloqueio
			WHERE listaBloqueio.ID_Alvo = grade.ID_Grade
		)
	`, obj.IdDispositivo)
	if erro != nil {
		aux.RespostaErro(w, http.StatusBadRequest, erro)
		return
	}
	defer tab.Close()

	if tab.Next() {

		if erro := tab.Scan(
			&obj.Nome,

			&obj.DomEam,
			&obj.DomSam,
			&obj.DomEpm,
			&obj.DomSpm,

			&obj.SegEam,
			&obj.SegSam,
			&obj.SegEpm,
			&obj.SegSpm,

			&obj.TerEam,
			&obj.TerSam,
			&obj.TerEpm,
			&obj.TerSpm,

			&obj.QuaEam,
			&obj.QuaSam,
			&obj.QuaEpm,
			&obj.QuaSpm,

			&obj.QuiEam,
			&obj.QuiSam,
			&obj.QuiEpm,
			&obj.QuiSpm,

			&obj.SexEam,
			&obj.SexSam,
			&obj.SexEpm,
			&obj.SexSpm,

			&obj.SabEam,
			&obj.SabSam,
			&obj.SabEpm,
			&obj.SabSpm,

			&obj.Tol,

			&obj.NomeDisp,
		); erro != nil {
			aux.RespostaErro(w, http.StatusBadRequest, erro)
			return
		}

		aux.RespostaJsonDados(w, http.StatusOK, obj)
		return
	}

	aux.RespostaJsonVazio(w)
}
