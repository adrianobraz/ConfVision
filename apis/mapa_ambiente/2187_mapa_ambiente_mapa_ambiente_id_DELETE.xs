// Delete mapa_ambiente record
query "mapa_ambiente/{mapa_ambiente_id}" verb=DELETE {
  api_group = "mapaAmbiente"

  input {
    int mapa_ambiente_id? filters=min:1
  }

  stack {
    db.del mapa_ambiente {
      field_name = "id"
      field_value = $input.mapa_ambiente_id
    }
  }

  response = null
}