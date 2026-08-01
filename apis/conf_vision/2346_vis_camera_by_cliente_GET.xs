// Listar cameras do cliente
query vis_camera_by_cliente verb=GET {
  api_group = "confVision"

  input {
    text id_cliente? filters=trim
  }

  stack {
    precondition (($input.id_cliente|is_empty) == false) {
      error = "id_cliente obrigatorio"
    }
  
    db.query vis_camera {
      where = $db.vis_camera.id_cliente == $input.id_cliente
      sort = {vis_camera.created_at: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista}
}
