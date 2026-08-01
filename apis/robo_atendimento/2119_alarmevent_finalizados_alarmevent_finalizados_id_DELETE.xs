// Delete alarmEvent_Finalizados record.
query "alarmevent_finalizados/{alarmevent_finalizados_id}" verb=DELETE {
  api_group = "RoboAtendimento"

  input {
    int alarmevent_finalizados_id? filters=min:1
  }

  stack {
    db.del alarmEvent_Finalizados {
      field_name = "id"
      field_value = $input.alarmevent_finalizados_id
    }
  }

  response = null
}