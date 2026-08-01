// Get mapa_ambiente record
query "mapa_ambiente/{mapa_ambiente_id}" verb=GET {
  api_group = "mapaAmbiente"

  input {
    int mapa_ambiente_id? filters=min:1
  }

  stack {
    db.get mapa_ambiente {
      field_name = "id"
      field_value = $input.mapa_ambiente_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = $model
}