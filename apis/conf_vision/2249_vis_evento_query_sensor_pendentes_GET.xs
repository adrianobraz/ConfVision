// Eventos sensor aguardando captura (worker ConfVision)
query vis_evento_query_sensor_pendentes verb=GET {
  api_group = "confVision"

  input {
    int limit?=10
  }

  stack {
    db.query vis_evento {
      where = $db.vis_evento.tipo_deteccao == "sensor" && $db.vis_evento.status == "capturando"
      sort = {vis_evento.created_at: "asc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista|slice:0:$input.limit}
}