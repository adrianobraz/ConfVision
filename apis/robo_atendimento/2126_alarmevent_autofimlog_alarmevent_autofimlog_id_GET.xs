// Get alarmEvent_autofimlog record
query "alarmevent_autofimlog/{alarmevent_autofimlog_id}" verb=GET {
  api_group = "RoboAtendimento"

  input {
    int alarmevent_autofimlog_id? filters=min:1
  }

  stack {
    db.get alarmEvent_autofimlog {
      field_name = "id"
      field_value = $input.alarmevent_autofimlog_id
    } as $alarmevent_autofimlog
  
    precondition ($alarmevent_autofimlog != null) {
      error_type = "notfound"
      error = "Not Found."
    }
  }

  response = $alarmevent_autofimlog
}