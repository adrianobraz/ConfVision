// Get alarmEvent_Finalizados record
query "alarmevent_finalizados/{alarmevent_finalizados_id}" verb=GET {
  api_group = "RoboAtendimento"

  input {
    int alarmevent_finalizados_id? filters=min:1
  }

  stack {
    db.get alarmEvent_Finalizados {
      field_name = "id"
      field_value = $input.alarmevent_finalizados_id
    } as $alarmevent_finalizados
  
    precondition ($alarmevent_finalizados != null) {
      error_type = "notfound"
      error = "Not Found."
    }
  }

  response = $alarmevent_finalizados
}