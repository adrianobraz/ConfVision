query retornaeventos verb=POST {
  api_group = "n8nBya"

  input {
    text idDispositivo? filters=trim
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.idDispositivo == $input.idDispositivo
      sort = {alarm_events.created_at: "desc"}
      return = {
        type  : "list"
        paging: {page: 1, per_page: 25, metadata: false}
      }
    
      output = ["particao", "zonaUser", "dataEntrada", "ctiGrupo", "ctiDescricao"]
    } as $alarm_events1
  }

  response = {dados: $alarm_events1}
}