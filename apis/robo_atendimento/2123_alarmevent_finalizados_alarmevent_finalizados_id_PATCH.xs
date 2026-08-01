// Edit alarmEvent_Finalizados record
query "alarmevent_finalizados/{alarmevent_finalizados_id}" verb=PATCH {
  api_group = "RoboAtendimento"

  input {
    int alarmevent_finalizados_id? filters=min:1
    dblink {
      table = "alarmEvent_Finalizados"
    }
  }

  stack {
    util.get_raw_input {
      encoding = "json"
      exclude_middleware = false
    } as $raw_input
  
    db.patch alarmEvent_Finalizados {
      field_name = "id"
      field_value = $input.alarmevent_finalizados_id
      data = `$input|pick:($raw_input|keys)`|filter_null|filter_empty_text
    } as $alarmevent_finalizados
  }

  response = $alarmevent_finalizados
}