// Query all alarmEvent_autofimlog records
query alarmevent_autofimlog verb=GET {
  api_group = "RoboAtendimento"

  input {
  }

  stack {
    db.query alarmEvent_autofimlog {
      return = {type: "list"}
    } as $alarmevent_autofimlog
  }

  response = $alarmevent_autofimlog
}