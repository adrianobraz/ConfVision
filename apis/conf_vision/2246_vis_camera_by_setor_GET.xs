// Busca camera ConfVision vinculada a um setor (dispositivo + particao + zonauser)
query vis_camera_by_setor verb=GET {
  api_group = "confVision"

  input {
    text id_dispositivo? filters=trim
    text particao? filters=trim
    text zonauser? filters=trim
    text id_franqueado? filters=trim
  }

  stack {
    db.query vis_camera {
      where = $db.vis_camera.id_dispositivo == $input.id_dispositivo && $db.vis_camera.particao == $input.particao && $db.vis_camera.zonauser == $input.zonauser && $db.vis_camera.id_franqueado ==? $input.id_franqueado
      sort = {vis_camera.id: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista|first}
}