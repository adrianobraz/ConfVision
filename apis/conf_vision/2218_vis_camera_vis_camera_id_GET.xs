// Get vis_camera record
query "vis_camera/{vis_camera_id}" verb=GET {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = $model
}