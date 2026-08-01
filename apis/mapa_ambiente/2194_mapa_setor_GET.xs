// Query all mapa_setor records
query mapa_setor verb=GET {
  api_group = "mapaAmbiente"

  input {
  }

  stack {
    db.query mapa_setor {
      return = {type: "list"}
    } as $model
  }

  response = $model
}