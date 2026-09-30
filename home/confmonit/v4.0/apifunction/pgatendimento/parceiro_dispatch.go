package pgatendimento

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"apifunction/confservice"
)

// ParceiroDispatchInput dados do evento para envio ao parceiro (EventGateway / autofim).
type ParceiroDispatchInput struct {
	AlarmEventsID int64  `json:"alarm_events_id"`
	IDEvento      string `json:"id_evento"`
	IDFranqueado  string `json:"id_franqueado"`
	IDCliente     string `json:"id_cliente"`
	IDProcesso    string `json:"id_processo"`
	IDDispositivo string `json:"id_dispositivo"`
	NomeCliente   string `json:"nome_cliente"`
	CtiGrupo      string `json:"cti_grupo"`
	CtiDescricao  string `json:"cti_descricao"`
	Codigo        string `json:"codigo"`
	Particao      string `json:"particao"`
	ZonaUser      string `json:"zona_user"`
	Conta         string `json:"conta"`
	DataEntrada   string `json:"data_entrada"`
	Nivel         string `json:"nivel"`
}

type alarmEventRow struct {
	ID            int64
	IDEvento      string
	IDFranqueado  string
	IDCliente     string
	IDProcesso    string
	IDDispositivo string
	NomeCliente   string
	CtiGrupo      string
	CtiDescricao  string
	Codigo        string
	Particao      string
	ZonaUser      string
	Conta         string
	DataEntrada   string
	Nivel         string
}

func loadAlarmEventRow(ctx context.Context, alarmEventsID int64) (*alarmEventRow, error) {
	if alarmEventsID <= 0 {
		return nil, errors.New("alarm_events_id invalido")
	}
	d, err := db(ctx)
	if err != nil {
		return nil, err
	}
	var row alarmEventRow
	var idEv, idFra, idCli, idProc, idDisp, nome, grupo, desc, cod, part, zona, conta, data, nivel sql.NullString
	err = d.QueryRowContext(ctx, `
SELECT id, id_evento, id_franqueado, id_cliente, id_processo, id_dispositivo,
       nome_cliente, cti_grupo, cti_descricao, codigo, particao, zona_user, conta, data_entrada, nivel
FROM ops_alarm_events WHERE id = $1`, alarmEventsID).Scan(
		&row.ID, &idEv, &idFra, &idCli, &idProc, &idDisp,
		&nome, &grupo, &desc, &cod, &part, &zona, &conta, &data, &nivel,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ops_alarm_events id=%d nao encontrado", alarmEventsID)
	}
	if err != nil {
		return nil, err
	}
	row.IDEvento = sqlString(idEv)
	row.IDFranqueado = sqlString(idFra)
	row.IDCliente = sqlString(idCli)
	row.IDProcesso = sqlString(idProc)
	row.IDDispositivo = sqlString(idDisp)
	row.NomeCliente = sqlString(nome)
	row.CtiGrupo = sqlString(grupo)
	row.CtiDescricao = sqlString(desc)
	row.Codigo = sqlString(cod)
	row.Particao = sqlString(part)
	row.ZonaUser = sqlString(zona)
	row.Conta = sqlString(conta)
	row.DataEntrada = sqlString(data)
	row.Nivel = sqlString(nivel)
	return &row, nil
}

func sqlString(v sql.NullString) string {
	if v.Valid {
		return strings.TrimSpace(v.String)
	}
	return ""
}

func mergeDispatchInput(ctx context.Context, in ParceiroDispatchInput) (ParceiroDispatchInput, error) {
	if in.AlarmEventsID > 0 {
		row, err := loadAlarmEventRow(ctx, in.AlarmEventsID)
		if err != nil {
			return in, err
		}
		fill := func(dst *string, val string) {
			if strings.TrimSpace(*dst) == "" {
				*dst = val
			}
		}
		fill(&in.IDEvento, row.IDEvento)
		fill(&in.IDFranqueado, row.IDFranqueado)
		fill(&in.IDCliente, row.IDCliente)
		fill(&in.IDProcesso, row.IDProcesso)
		fill(&in.IDDispositivo, row.IDDispositivo)
		fill(&in.NomeCliente, row.NomeCliente)
		fill(&in.CtiGrupo, row.CtiGrupo)
		fill(&in.CtiDescricao, row.CtiDescricao)
		fill(&in.Codigo, row.Codigo)
		fill(&in.Particao, row.Particao)
		fill(&in.ZonaUser, row.ZonaUser)
		fill(&in.Conta, row.Conta)
		fill(&in.DataEntrada, row.DataEntrada)
		fill(&in.Nivel, row.Nivel)
	}
	in.IDFranqueado = strings.TrimSpace(in.IDFranqueado)
	in.IDCliente = strings.TrimSpace(in.IDCliente)
	in.CtiGrupo = strings.TrimSpace(in.CtiGrupo)
	if in.IDFranqueado == "" || in.IDCliente == "" {
		return in, errors.New("id_franqueado e id_cliente obrigatorios")
	}
	return in, nil
}

// DispatchEventoParceiro resolve politica, enfileira ConfService e registra envio ops.
func DispatchEventoParceiro(ctx context.Context, in ParceiroDispatchInput) (map[string]any, error) {
	in, err := mergeDispatchInput(ctx, in)
	if err != nil {
		return nil, err
	}

	codigo := strings.TrimSpace(in.Codigo)
	if codigo == "1M01" || codigo == "3M02" {
		return map[string]any{"ok": false, "motivo": "codigo_ignorado"}, nil
	}

	ativo, err := ResolveComHorario(ctx, in.IDFranqueado, in.IDCliente, RecursoParceiro, in.CtiGrupo, nil)
	if err != nil {
		return nil, err
	}
	if !ativo {
		return map[string]any{"ok": false, "motivo": "parceiro_inativo"}, nil
	}

	liberado, motivoFin, err := ParceiroFinanceiroLiberado(ctx, in.IDFranqueado)
	if err != nil {
		return nil, err
	}
	if !liberado {
		return map[string]any{"ok": false, "motivo": motivoFin}, nil
	}

	cliAtivo, err := ClienteMonitoramentoAtivo(in.IDFranqueado, in.IDCliente)
	if err != nil {
		return nil, err
	}
	if !cliAtivo {
		return map[string]any{"ok": false, "motivo": "cliente_inativo"}, nil
	}

	vinc, err := ResolveParceiroVinculo(ctx, in.IDFranqueado, in.IDCliente, in.CtiGrupo)
	if err != nil {
		return nil, err
	}
	if !vinc.OK || strings.TrimSpace(vinc.IDParceiro) == "" {
		return map[string]any{"ok": false, "motivo": "sem_vinculo_parceiro"}, nil
	}

	alarmID := in.AlarmEventsID
	payloadEvento := map[string]any{
		"alarmEventsId": alarmID,
		"idEvento":      in.IDEvento,
		"idProcesso":    in.IDProcesso,
		"idFranqueado":  in.IDFranqueado,
		"idCliente":     in.IDCliente,
		"nomeCliente":   in.NomeCliente,
		"idDispositivo": in.IDDispositivo,
		"ctiGrupo":      in.CtiGrupo,
		"ctiDescricao":  in.CtiDescricao,
		"codigo":        in.Codigo,
		"particao":      in.Particao,
		"zonaUser":      in.ZonaUser,
		"conta":         in.Conta,
		"dataEntrada":   in.DataEntrada,
		"nivel":         in.Nivel,
		"codigoInterno": vinc.CodigoInterno,
		"codigo_interno": vinc.CodigoInterno,
	}

	csBody := map[string]any{
		"idFranqueado":  in.IDFranqueado,
		"idCliente":     in.IDCliente,
		"idParceiro":    vinc.IDParceiro,
		"codigoInterno": vinc.CodigoInterno,
		"nomeCliente":   in.NomeCliente,
		"payload":       payloadEvento,
	}
	if alarmID > 0 {
		csBody["alarmEventsId"] = alarmID
	}

	csResp, err := confservice.EnfileirarEvento(csBody)
	if err != nil {
		_ = RegistrarParceiroEnvio(ctx, in.IDFranqueado, in.IDCliente, in.IDProcesso, alarmID,
			vinc.IDParceiro, in.CtiGrupo, "ERRO", map[string]any{"erro": err.Error()})
		return nil, err
	}

	envio, err := ProcessarEnvioParceiro(ctx, map[string]any{
		"id_franqueado":   in.IDFranqueado,
		"id_cliente":      in.IDCliente,
		"id_processo":     in.IDProcesso,
		"alarm_events_id": alarmID,
		"cti_grupo":       in.CtiGrupo,
		"id_parceiro":     vinc.IDParceiro,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"ok":             true,
		"id_parceiro":    vinc.IDParceiro,
		"codigo_interno": vinc.CodigoInterno,
		"confservice":    csResp,
		"finalizacao":    envio,
	}, nil
}
