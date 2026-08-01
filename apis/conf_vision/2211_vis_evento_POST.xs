// Add vis_evento record
query vis_evento verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_evento"
    }
  }

  stack {
    db.add vis_evento {
      enforce_hidden_fields = false
      data = {
        created_at     : "now"
        vis_camera_id  : $input.vis_camera_id
        id_franqueado  : $input.id_franqueado
        id_cliente     : $input.id_cliente
        id_dispositivo : $input.id_dispositivo
        conta          : $input.conta
        particao       : $input.particao
        canal          : $input.canal
        tipo_deteccao  : $input.tipo_deteccao
        confianca      : $input.confianca
        snapshot_url   : $input.snapshot_url
        video_url      : $input.video_url
        bbox_json      : $input.bbox_json
        processado     : $input.processado
        alarm_events_id: $input.alarm_events_id
        ignorado       : $input.ignorado
        status         : $input.status
        id_evento      : $input.id_evento
        id_processo    : $input.id_processo
        started_at     : $input.started_at
        ended_at       : $input.ended_at
        clip_count     : $input.clip_count
      }
    } as $model
  }

  response = $model
}