// Listar areas de deteccao de uma camera
query vis_camera_area_by_camera verb=GET {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
  }

  stack {
    db.query vis_camera_area {
      where = $db.vis_camera_area.vis_camera_id == $input.vis_camera_id
      sort = {vis_camera_area.created_at: "asc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista}
}