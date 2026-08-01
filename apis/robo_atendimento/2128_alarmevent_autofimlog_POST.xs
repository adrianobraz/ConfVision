// Add alarmEvent_autofimlog record
query alarmevent_autofimlog verb=POST {
  api_group = "RoboAtendimento"

  input {
    dblink {
      table = "alarmEvent_autofimlog"
    }
  }

  stack {
    db.add alarmEvent_autofimlog {
      enforce_hidden_fields = false
      data = {created_at: "now"}
    } as $alarmevent_autofimlog
  }

  response = $alarmevent_autofimlog
}