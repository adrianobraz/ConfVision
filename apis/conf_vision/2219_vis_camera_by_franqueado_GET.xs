// Listar cameras do franqueado
query vis_camera_by_franqueado verb=GET {
  api_group = "confVision"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    db.query vis_camera {
      where = $db.vis_camera.id_franqueado == $input.id_franqueado
      sort = {vis_camera.created_at: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista}
}