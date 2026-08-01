// Query all vis_camera records
query vis_camera verb=GET {
  api_group = "confVision"

  input {
  }

  stack {
    db.query vis_camera {
      return = {type: "list"}
    } as $model
  }

  response = $model
}