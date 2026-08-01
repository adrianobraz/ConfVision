// Disparo de captura por evento de alarme (licenca sensor) — receptor chama apos ALARME/PANICO
query vis_evento_disparo_sensor verb=POST {
  api_group = "confVision"

  input {
    text id_evento? filters=trim
    text id_processo? filters=trim
    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    text conta? filters=trim
    text particao? filters=trim
    text zonauser? filters=trim
    int alarm_events_id?
  }

  stack {
    precondition ($env.$http_headers.Authorization == "Bearer 1e2d2eef75ccd7cfea2e06d7ce1c63c4") {
      error = "nao autorizado"
    }
  
    db.query vis_camera {
      where = $db.vis_camera.id_dispositivo == $input.id_dispositivo && $db.vis_camera.particao == $input.particao && $db.vis_camera.zonauser == $input.zonauser && $db.vis_camera.captura_sensor == true
      sort = {vis_camera.id: "desc"}
      return = {type: "list"}
    } as $cameras
  
    var $camera {
      value = $cameras|first
    }
  
    precondition ($camera != null) {
      error = "Camera sensor nao encontrada para o setor"
    }
  
    precondition ($camera.evento_grava_foto || $camera.evento_grava_video) {
      error = "Licenca sensor sem permissao de foto ou video"
    }
  
    db.add vis_evento {
      enforce_hidden_fields = false
      data = {
        created_at     : "now"
        vis_camera_id  : $camera.id
        id_franqueado  : $input.id_franqueado
        id_cliente     : $input.id_cliente
        id_dispositivo : $input.id_dispositivo
        conta          : $input.conta
        particao       : $input.particao
        canal          : $camera.canal
        tipo_deteccao  : "sensor"
        confianca      : 1
        processado     : false
        ignorado       : false
        status         : "capturando"
        id_evento      : $input.id_evento
        id_processo    : $input.id_processo
        alarm_events_id: $input.alarm_events_id
        clip_count     : 0
      }
    } as $evento
  }

  response = {dados: $evento, camera: $camera}
}