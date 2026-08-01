// Atualizar area de deteccao
query "vis_camera_area/{id}" verb=PUT {
  api_group = "confVision"

  input {
    int id? filters=min:1
    dblink {
      table = "vis_camera_area"
    }
  }

  stack {
    db.edit vis_camera_area {
      field_name = "id"
      field_value = $input.id
      enforce_hidden_fields = false
      data = {
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