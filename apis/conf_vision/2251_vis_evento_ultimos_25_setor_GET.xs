// Ultimos 25 eventos ConfVision de uma camera/setor (terminal)
query vis_evento_ultimos25_setor verb=GET {
  api_group = "confVision"

  input {
    text id_dispositivo? filters=trim
    text particao? filters=trim
    text zonauser? filters=trim
  }

  stack {
    db.query vis_camera {
      where = $db.vis_camera.id_dispositivo == $input.id_dispositivo && $db.vis_camera.particao == $input.particao && $db.vis_camera.zonauser == $input.zonauser
      sort = {vis_camera.id: "desc"}
      return = {type: "list"}
    } as $cameras
  
    var $camera {
      value = $cameras|first
    }
  
    precondition ($camera != null) {
      error = "Camera nao encontrada para o setor"
    }
  
    db.query vis_evento {
      where = $db.vis_evento.vis_camera_id == $camera.id && $db.vis_evento.id_dispositivo == $input.id_dispositivo
      sort = {vis_evento.created_at: "desc"}
      return = {
        type  : "list"
        paging: {page: 1, per_page: 25, metadata: false}
      }
    } as $resultado
  }

  response = {dados: $resultado, camera: $camera}
}