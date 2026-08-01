// Atualiza apenas snapshot_url da câmera (cadastro)
query "vis_camera/snapshot/{vis_camera_id}" verb=PUT {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
    text snapshot_url? filters=trim
  }

  stack {
    db.edit vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      enforce_hidden_fields = false
      data = {snapshot_url: $input.snapshot_url}
    } as $model
  }

  response = $model
}