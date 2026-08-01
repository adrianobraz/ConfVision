// Update vis_camera record
query "vis_camera/status/{vis_camera_id}" verb=PUT {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
    dblink {
      table = "vis_camera"
    }
  }

  stack {
    db.edit vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      enforce_hidden_fields = false
      data = {
        ativo           : $input.ativo
        nome            : $input.nome
        id_franqueado   : $input.id_franqueado
        id_cliente      : $input.id_cliente
        id_dispositivo  : $input.id_dispositivo
        conta           : $input.conta
        particao        : $input.particao
        canal           : $input.canal
        setor           : $input.setor
        protocolo       : $input.protocolo
        rtsp_url_sec    : $input.rtsp_url_sec
        onvif_host      : $input.onvif_host
        onvif_porta     : $input.onvif_porta
        onvif_usuario   : $input.onvif_usuario
        onvif_senha     : $input.onvif_senha
        confianca_min   : $input.confianca_min
        cooldown_seg    : $input.cooldown_seg
        somente_armado  : $input.somente_armado
        deteccao_humano : $input.deteccao_humano
        deteccao_veiculo: $input.deteccao_veiculo
        status          : $input.status
        ultimo_evento_em: $input.ultimo_evento_em
        worker_id       : $input.worker_id
        ultimo_ping_em  : $input.ultimo_ping_em
      }
    } as $model
  }

  response = $model
}