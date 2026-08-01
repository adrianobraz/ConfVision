// Get mapa_setor record
query "mapa_setor/{mapa_setor_id}" verb=GET {
  api_group = "mapaAmbiente"

  input {
    int mapa_setor_id? filters=min:1
  }

  stack {
    db.get mapa_setor {
      field_name = "id"
      field_value = $input.mapa_setor_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = $model
}