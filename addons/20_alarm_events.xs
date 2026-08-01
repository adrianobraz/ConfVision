addon alarm_events {
  input {
    int alarm_events_id? {
      table = "alarm_events"
    }
  }

  stack {
    db.query alarm_events {
      where = $db.alarm_events.id == $input.alarm_events_id
      sort = {alarm_events.dataEntrada: "desc"}
      return = {type: "single"}
    }
  }
}