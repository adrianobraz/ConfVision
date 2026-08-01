// Excluir area de deteccao
query "vis_camera_area/{id}" verb=DELETE {
  api_group = "confVision"

  input {
    int id? filters=min:1
  }

  stack {
    db.del vis_camera_area {
      field_name = "id"
      field_value = $input.id
    }
  }

  response = {ok: true}
}