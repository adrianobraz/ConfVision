// Sinaliza flush imediato do segmento atual (continua ou timelapse)
query "vis_camera/gravacao/flush/{vis_camera_id}" verb=POST {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $camera
  
    precondition ($camera != null) {
      error = "Camera nao encontrada"
    }
  
    precondition ($camera.gravacao_status == "ativa") {
      error = "Camera sem gravacao ativa"
    }
  
    precondition ($camera.grava_continua || $camera.grava_timelapse) {
      error = "Flush disponivel apenas para gravacao continua ou timelapse"
    }
  
    db.patch vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      data = {}|set:"gravacao_flush_pedido":true
    } as $result
  
    var $response {
      value = {
        success      : true
        vis_camera_id: $input.vis_camera_id
        message      : "Sinal de envio imediato registrado"
      }
    }
  }

  response = $response[""]
}