package pgatendimento

import (
	"context"
	"encoding/json"
	"strings"
)

func RegistrarParceiroEnvio(ctx context.Context, idFranqueado, idCliente, idProcesso string, alarmEventsID int64, idParceiro, ctiGrupo, acaoFinal string, detalhe map[string]any) error {
	d, err := db(ctx)
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
	_, err = d.ExecContext(ctx, `
INSERT INTO ops_parceiro_envio_log (
    id_franqueado, id_cliente, id_processo, alarm_events_id, id_parceiro, cti_grupo, acao_final, detalhe
) VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8::jsonb)`,
		strings.TrimSpace(idFranqueado), strings.TrimSpace(idCliente), strings.TrimSpace(idProcesso),
		alarmID, strings.TrimSpace(idParceiro), strings.TrimSpace(ctiGrupo), strings.TrimSpace(acaoFinal), string(raw))
	return err
}

func ProcessarEnvioParceiro(ctx context.Context, payload map[string]any) (map[string]any, error) {
	idFra := strMap(payload, "id_franqueado")
	idCli := strMap(payload, "id_cliente")
	idProc := strMap(payload, "id_processo")
	ctiGrupo := strMap(payload, "cti_grupo")
	idParceiro := strMap(payload, "id_parceiro")
	alarmEventsID := int64(intMap(payload, "alarm_events_id"))

	duplaCom := false
	if pol, err := GetPolitica(ctx, idFra); err == nil {
		duplaCom = pol.ParceiroDuplaComunicacao
	}

	var retorno map[string]any
	acao := "DUPLA_COM"
	if !duplaCom {
		var err error
		retorno, acao, err = processoEnd(ctx, idProc, "Enviado ao parceiro de monitoramento", "PARCEIRO")
		if err != nil {
			_ = RegistrarParceiroEnvio(ctx, idFra, idCli, idProc, alarmEventsID, idParceiro, ctiGrupo, "ERRO", map[string]any{
				"erro": err.Error(),
			})
			return nil, err
		}
	}

	detalhe := map[string]any{"dupla_comunicacao": duplaCom}
	if retorno != nil {
		detalhe["retorno_processo_end"] = retorno
	}
	if err := RegistrarParceiroEnvio(ctx, idFra, idCli, idProc, alarmEventsID, idParceiro, ctiGrupo, acao, detalhe); err != nil {
		return nil, err
	}

	return map[string]any{
		"ok":                true,
		"acao_final":        acao,
		"dupla_comunicacao": duplaCom,
		"finalizacao":       retorno,
	}, nil
}
