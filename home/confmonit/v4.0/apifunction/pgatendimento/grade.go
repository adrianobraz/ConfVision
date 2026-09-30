package pgatendimento

import (
	"context"
	"strings"
	"time"
)

const (
	Cobertura24h   = "24h"
	CoberturaGrade = "grade"

	ResponsavelIA       = "inteligencia_artificial"
	ResponsavelParceiro = "parceiro"
)

var gruposEmergencia = map[string]struct{}{
	"PANICO":     {},
	"EMERGENCIA": {},
	"COACAO":     {},
}

type GradeSlot struct {
	ID           int    `json:"id"`
	IDFranqueado string `json:"id_franqueado"`
	IDCliente    string `json:"id_cliente"`
	Responsavel  string `json:"responsavel"`
	DiasSemana   []int  `json:"dias_semana"`
	HoraInicio   string `json:"hora_inicio"`
	HoraFim      string `json:"hora_fim"`
}

func IsGrupoEmergencia(ctiGrupo string) bool {
	_, ok := gruposEmergencia[strings.ToUpper(strings.TrimSpace(ctiGrupo))]
	return ok
}

func ListGradeSlots(ctx context.Context, idFranqueado, idCliente, responsavel string) ([]GradeSlot, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	responsavel = strings.TrimSpace(responsavel)
	if idFranqueado == "" || responsavel == "" {
		return nil, nil
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := d.QueryContext(ctx, `
SELECT id, id_franqueado, COALESCE(id_cliente,''), responsavel, dias_semana,
       to_char(hora_inicio, 'HH24:MI'), to_char(hora_fim, 'HH24:MI')
FROM ops_cliente_atendimento_grade
WHERE id_franqueado = $1 AND COALESCE(id_cliente,'') = $2 AND responsavel = $3
ORDER BY hora_inicio`, idFranqueado, idCliente, responsavel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GradeSlot
	for rows.Next() {
		var s GradeSlot
		var dias []int64
		if err := rows.Scan(&s.ID, &s.IDFranqueado, &s.IDCliente, &s.Responsavel, &dias, &s.HoraInicio, &s.HoraFim); err != nil {
			return nil, err
		}
		s.DiasSemana = int64SliceToInt(dias)
		out = append(out, s)
	}
	return out, rows.Err()
}

func int64SliceToInt(in []int64) []int {
	out := make([]int, 0, len(in))
	for _, v := range in {
		out = append(out, int(v))
	}
	return out
}

func normalizeHoraMin(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) == 5 {
		return s + ":00"
	}
	return s
}

func coberturaHorariaRecurso(p Politica, recurso string) string {
	switch strings.TrimSpace(recurso) {
	case RecursoIA:
		if strings.TrimSpace(p.IACoberturaHoraria) != "" {
			return strings.TrimSpace(p.IACoberturaHoraria)
		}
		return strings.TrimSpace(p.CoberturaHoraria)
	case RecursoParceiro:
		if strings.TrimSpace(p.ParceiroCoberturaHoraria) != "" {
			return strings.TrimSpace(p.ParceiroCoberturaHoraria)
		}
		return strings.TrimSpace(p.CoberturaHoraria)
	default:
		return Cobertura24h
	}
}

func parceiroGrupoPermitido(p Politica, ctiGrupo string) bool {
	grupo := strings.ToUpper(strings.TrimSpace(ctiGrupo))
	if grupo == "" {
		return true
	}
	if len(p.GruposEventoParceiro) == 0 {
		return true
	}
	for _, g := range p.GruposEventoParceiro {
		if strings.ToUpper(strings.TrimSpace(g)) == grupo {
			return true
		}
	}
	return false
}

func DentroDaGrade(ctx context.Context, idFranqueado, idCliente, responsavel string, now time.Time) (bool, error) {
	slots, err := ListGradeSlots(ctx, idFranqueado, idCliente, responsavel)
	if err != nil {
		return false, err
	}
	if len(slots) == 0 && idCliente != "" {
		slots, err = ListGradeSlots(ctx, idFranqueado, "", responsavel)
		if err != nil {
			return false, err
		}
	}
	if len(slots) == 0 {
		return false, nil
	}
	weekday := int(now.Weekday())
	clock := now.Format("15:04:05")
	for _, s := range slots {
		if !containsInt(s.DiasSemana, weekday) {
			continue
		}
		if horarioDentroFaixa(clock, s.HoraInicio, s.HoraFim) {
			return true, nil
		}
	}
	return false, nil
}

func containsInt(list []int, v int) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func horarioDentroFaixa(clock, inicio, fim string) bool {
	inicio = normalizeHoraMin(inicio)
	fim = normalizeHoraMin(fim)
	clock = normalizeHoraMin(clock)
	if inicio == "" || fim == "" {
		return false
	}
	if inicio <= fim {
		return clock >= inicio && clock <= fim
	}
	return clock >= inicio || clock <= fim
}

func responsavelFromRecurso(recurso string) string {
	switch strings.TrimSpace(recurso) {
	case RecursoIA:
		return ResponsavelIA
	case RecursoParceiro:
		return ResponsavelParceiro
	default:
		return ""
	}
}

func ResolveComHorario(ctx context.Context, idFranqueado, idCliente, recurso, ctiGrupo string, politica *Politica) (bool, error) {
	ok, err := ResolveRecursoAtivo(ctx, idFranqueado, idCliente, recurso, politica)
	if err != nil || !ok {
		return ok, err
	}

	if IsGrupoEmergencia(ctiGrupo) {
		return true, nil
	}

	var p Politica
	if politica != nil {
		p = *politica
	} else {
		p, err = GetPolitica(ctx, idFranqueado)
		if err != nil {
			return false, err
		}
	}

	if recurso == RecursoParceiro && !parceiroGrupoPermitido(p, ctiGrupo) {
		return false, nil
	}

	cob := coberturaHorariaRecurso(p, recurso)
	if cob != CoberturaGrade {
		return true, nil
	}

	responsavel := responsavelFromRecurso(recurso)
	if responsavel == "" {
		return true, nil
	}
	return DentroDaGrade(ctx, idFranqueado, idCliente, responsavel, time.Now())
}
