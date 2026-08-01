// Add alarmEvent_Finalizados record
query alarmevent_finalizados verb=POST {
  api_group = "RoboAtendimento"

  input {
    dblink {
      table = "alarmEvent_Finalizados"
    }
  }

  stack {
    db.add alarmEvent_Finalizados {
      enforce_hidden_fields = false
      data = {created_at: "now"}
    } as $alarmevent_finalizados
  }

  response = $alarmevent_finalizados
}