// Delete mapa_setor record
query "mapa_setor/{mapa_setor_id}" verb=DELETE {
  api_group = "mapaAmbiente"

  input {
    int mapa_setor_id? filters=min:1
  }

  stack {
    db.del mapa_setor {
      field_name = "id"
      field_value = $input.mapa_setor_id
    }
  }

  response = null
}