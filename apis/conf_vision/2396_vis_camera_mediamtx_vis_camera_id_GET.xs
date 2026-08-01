// URLs MediaMTX da camera (atribui no se necessario)
query "vis_camera/mediamtx/{vis_camera_id}" verb=GET {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
    bool? atribuir_se_ausente?=true
  }

  stack {
    function.run fn_vis_mediamtx_resolve_camera {
      input = {
        vis_camera_id     : $input.vis_camera_id
        atribuir_se_ausente: $input.atribuir_se_ausente|first_notempty:true
      }
    } as $mtx
  }

  response = $mtx
}
