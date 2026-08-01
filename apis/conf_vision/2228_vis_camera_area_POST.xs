// Criar area de deteccao
query vis_camera_area verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_camera_area"
    }
  }

  stack {
    db.add vis_camera_area {
      enforce_hidden_fields = false
      data = {
        created_at   : "now"
        vis_camera_id: $input.vis_camera_id
        nome         : $input.nome
        ativo        : $input.ativo
        poligono_json: $input.poligono_json
        cor          : $input.cor
      }
    } as $model
  }

  response = $model
}