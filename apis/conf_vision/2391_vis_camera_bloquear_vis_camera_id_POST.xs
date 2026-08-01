// Bloqueia ou desbloqueia publish RTMP da camera
query "vis_camera/bloquear/{vis_camera_id}" verb=POST {
  api_group = "confVision"

  input {
    int vis_camera_id filters=min:1
    bool bloqueado?
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $atual
  
    precondition ($atual != null) {
      error_type = "notfound"
      error = "Camera nao encontrada"
    }
  
    var $flag {
      value = $input.bloqueado
    }
  
    conditional {
      if ($flag == null) {
        var.update $flag {
          value = true
        }
      }
    }
  
    db.patch vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
      data = {bloqueado: $flag}
    } as $model
  }

  response = $model
}
