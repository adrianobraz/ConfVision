// Query all mapa_ambiente records
query mapa_ambiente verb=GET {
  api_group = "mapaAmbiente"

  input {
  }

  stack {
    db.query mapa_ambiente {
      return = {type: "list"}
    } as $model
  }

  response = $model
}