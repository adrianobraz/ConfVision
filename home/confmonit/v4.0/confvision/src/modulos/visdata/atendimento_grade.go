package visdata

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	Cobertura24h   = "24h"
	CoberturaGrade = "grade"

	ResponsavelIA       = "inteligencia_artificial"
	ResponsavelParceiro = "parceiro"
)

var gruposEmergenciaAtendimento = map[string]struct{}{
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

func IsGrupoEmergenciaAtendimento(ctiGrupo string) bool {
	_, ok := gruposEmergenciaAtendimento[strings.ToUpper(strings.TrimSpace(ctiGrupo))]
	return ok
}

func ListGradeSlots(ctx context.Context, idFranqueado, idCliente, responsavel string) ([]GradeSlot, error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	responsavel = strings.TrimSpace(responsavel)
	if idFranqueado == "" || responsavel == "" {
		return nil, errors.New("id_franqueado e responsavel obrigatorios")
	}
	db, err := DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, `
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

func ReplaceGradeSlots(ctx context.Context, idFranqueado, idCliente, responsavel string, slots []GradeSlot) error {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	responsavel = strings.TrimSpace(responsavel)
	if idFranqueado == "" || responsavel == "" {
		return errors.New("id_franqueado e responsavel obrigatorios")
	}
	db, err := DB()
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
DELETE FROM ops_cliente_atendimento_grade
WHERE id_franqueado = $1 AND COALESCE(id_cliente,'') = $2 AND responsavel = $3`,
		idFranqueado, idCliente, responsavel)
	if err != nil {
		return err
	}

	for _, s := range slots {
		dias := s.DiasSemana
		if len(dias) == 0 {
			continue
		}
		hi := normalizeHoraMin(s.HoraInicio)
		hf := normalizeHoraMin(s.HoraFim)
		if hi == "" || hf == "" {
			continue
		}
		_, err = tx.ExecContext(ctx, `
INSERT INTO ops_cliente_atendimento_grade (id_franqueado, id_cliente, responsavel, dias_semana, hora_inicio, hora_fim)
VALUES ($1,$2,$3,$4,$5::time,$6::time)`,
			idFranqueado, idCliente, responsavel, intSliceToPGArray(dias), hi, hf)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func int64SliceToInt(in []int64) []int {
	out := make([]int, 0, len(in))
	for _, v := range in {
		out = append(out, int(v))
	}
	return out
}

func intSliceToPGArray(in []int) []int32 {
	out := make([]int32, 0, len(in))
	for _, v := range in {
		out = append(out, int32(v))
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

func coberturaHorariaRecurso(p PoliticaAtendimento, recurso string) string {
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

func parceiroGrupoPermitido(p PoliticaAtendimento, ctiGrupo string) bool {
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

func ResolveRecursoAtivoComHorario(ctx context.Context, idFranqueado, idCliente, recurso, ctiGrupo string, politica *PoliticaAtendimento) (bool, error) {
	ok, err := ResolveRecursoAtivo(ctx, idFranqueado, idCliente, recurso, politica)
	if err != nil || !ok {
		return ok, err
	}

	if IsGrupoEmergenciaAtendimento(ctiGrupo) {
		return true, nil
	}

	var p PoliticaAtendimento
	if politica != nil {
		p = *politica
	} else {
		p, err = GetPoliticaAtendimento(ctx, idFranqueado)
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

func ResolveRecursoAtivoComHorarioCompat(ctx context.Context, idFranqueado, idCliente, recurso, ctiGrupo string) (ativo bool, failOpen bool, err error) {
	idFranqueado = strings.TrimSpace(idFranqueado)
	idCliente = strings.TrimSpace(idCliente)
	recurso = strings.TrimSpace(recurso)
	if idFranqueado == "" || idCliente == "" || recurso == "" {
		return true, true, nil
	}

	hasPol, err := franqueadoTemPoliticaAtendimento(ctx, idFranqueado)
	if err != nil {
		return true, true, err
	}
	hasCli, err := clienteTemConfigAtendimento(ctx, idFranqueado, idCliente)
	if err != nil {
		return true, true, err
	}
	if !hasPol && !hasCli {
		return true, true, nil
	}

	if !hasPol {
		ok, err := resolveFromClienteConfig(ctx, idFranqueado, idCliente, recurso)
		if err != nil {
			return false, false, err
		}
		if !ok {
			return false, false, nil
		}
		if IsGrupoEmergenciaAtendimento(ctiGrupo) {
			return true, false, nil
		}
		resp := responsavelFromRecurso(recurso)
		if resp == "" {
			return true, false, nil
		}
		okGrade, err := DentroDaGrade(ctx, idFranqueado, idCliente, resp, time.Now())
		return okGrade, false, err
	}

	ok, err := ResolveRecursoAtivoComHorario(ctx, idFranqueado, idCliente, recurso, ctiGrupo, nil)
	return ok, false, err
}

func RegistrarParceiroEnvio(ctx context.Context, idFranqueado, idCliente, idProcesso string, alarmEventsID int64, idParceiro, ctiGrupo, acaoFinal string, detalhe map[string]any) error {
	db, err := DB()
	if err != nil {
		return err
	}
	if detalhe == nil {
		detalhe = map[string]any{}
	}
	raw, err := json.Marshal(detalhe)
	if err != nil {
		return err
	}
	var alarmID any
	if alarmEventsID > 0 {
		alarmID = alarmEventsID
	}
	_, err = db.ExecContext(ctx, `
INSERT INTO ops_parceiro_envio_log (
    id_franqueado, id_cliente, id_processo, alarm_events_id, id_parceiro, cti_grupo, acao_final, detalhe
) VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8::jsonb)`,
		strings.TrimSpace(idFranqueado), strings.TrimSpace(idCliente), strings.TrimSpace(idProcesso),
		alarmID, strings.TrimSpace(idParceiro), strings.TrimSpace(ctiGrupo), strings.TrimSpace(acaoFinal), string(raw))
	return err
}

func ProcessarEnvioParceiro(ctx context.Context, payload map[string]any) (map[string]any, error) {
	idFra := strVal(payload, "id_franqueado")
	idCli := strVal(payload, "id_cliente")
	idProc := strVal(payload, "id_processo")
	ctiGrupo := strVal(payload, "cti_grupo")
	idParceiro := strVal(payload, "id_parceiro")
	alarmEventsID := int64(intVal(payload, "alarm_events_id"))

	retorno, acao, err := opsProcessoEnd(ctx, idProc, "Enviado ao parceiro de monitoramento", "PARCEIRO")
	if err != nil {
		_ = RegistrarParceiroEnvio(ctx, idFra, idCli, idProc, alarmEventsID, idParceiro, ctiGrupo, "ERRO", map[string]any{
			"erro": err.Error(),
		})
		return nil, err
	}

	detalhe := map[string]any{"retorno_processo_end": retorno}
	if err := RegistrarParceiroEnvio(ctx, idFra, idCli, idProc, alarmEventsID, idParceiro, ctiGrupo, acao, detalhe); err != nil {
		return nil, err
	}

	return map[string]any{
		"ok":           true,
		"acao_final":   acao,
		"finalizacao":  retorno,
	}, nil
}

func resolveFromClienteConfig(ctx context.Context, idFranqueado, idCliente, recurso string) (bool, error) {
	return ResolveRecursoAtivo(ctx, idFranqueado, idCliente, recurso, nil)
}

func franqueadoTemPoliticaAtendimento(ctx context.Context, idFranqueado string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_franqueado_atendimento_politica WHERE id_franqueado = $1`, strings.TrimSpace(idFranqueado)).Scan(&n)
	return n > 0, err
}

func clienteTemConfigAtendimento(ctx context.Context, idFranqueado, idCliente string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var n int
	err = db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM ops_cliente_atendimento_config
WHERE id_franqueado = $1 AND id_cliente = $2`, strings.TrimSpace(idFranqueado), strings.TrimSpace(idCliente)).Scan(&n)
	return n > 0, err
}
