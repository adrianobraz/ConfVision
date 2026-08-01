// Areas ativas — worker filtra por cameras do sync (vis_camera_query_ativas)
query vis_camera_area_query_ativas verb=GET {
  api_group = "confVision"

  input {
  }

  stack {
    db.query vis_camera_area {
      where = $db.vis_camera_area.ativo == true
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista}
}