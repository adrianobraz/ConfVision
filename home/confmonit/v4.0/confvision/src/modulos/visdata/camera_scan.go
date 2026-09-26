package visdata

import "database/sql"

const cameraSelectCols = `
c.id, c.created_at, c.ativo, c.bloqueado, c.nome, c.id_franqueado, c.id_cliente,
c.id_dispositivo, c.conta, c.particao, c.canal, c.setor, c.protocolo, c.rtsp_url_sec,
c.onvif_host, c.onvif_porta, c.onvif_usuario, c.onvif_senha, c.confianca_min, c.cooldown_seg,
c.somente_armado, c.deteccao_humano, c.deteccao_veiculo, c.status, c.ultimo_evento_em,
c.worker_id, c.ultimo_ping_em, c.zonauser, c.captura_sensor, c.captura_analitico,
c.analitico_pausado, c.ultimo_stream_ok_em, c.stream_falhas_consecutivas, c.stream_tentativas_horarias,
c.stream_policy_generation, c.stream_motivo_pausa, c.id_setor, c.snapshot_url, c.vis_licenca_id, c.plano, c.ativado_em,
c.evento_grava_foto, c.evento_grava_video, c.vis_licenca_gravacao_id, c.grava_continua,
c.retencao_dias, c.gravacao_ativada_em, c.gravacao_status, c.grava_movimento,
c.grava_timelapse, c.gravacao_flush_pedido, c.modo_deteccao, c.vis_mediamtx_node_id,
COALESCE(n.rtsp_internal, '') AS mediamtx_rtsp_base
`

func scanCameraRows(rows *sql.Rows) ([]map[string]any, error) {
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		dest := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range dest {
			ptrs[i] = &dest[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := make(map[string]any, len(cols))
		for i, col := range cols {
			m[col] = normalizeValue(dest[i])
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
