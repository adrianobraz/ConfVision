addon Addon_Relatorio_Events {
  input {
    int alarm_events_id? {
      table = "alarm_events"
    }
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.id == $input.alarm_events_id
      sort = {alarm_events.created_at: "asc"}
      return = {type: "list"}
    }
  }
}