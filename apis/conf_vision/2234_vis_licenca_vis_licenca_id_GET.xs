// Obter licenca por id
query "vis_licenca/{vis_licenca_id}" verb=GET {
  api_group = "confVision"

  input {
    int vis_licenca_id? filters=min:1
  }

  stack {
    db.get vis_licenca {
      field_name = "id"
      field_value = $input.vis_licenca_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Licenca nao encontrada"
    }
  }

  response = $model
}