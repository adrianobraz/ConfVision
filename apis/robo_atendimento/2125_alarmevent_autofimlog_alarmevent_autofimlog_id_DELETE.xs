// Delete alarmEvent_autofimlog record.
query "alarmevent_autofimlog/{alarmevent_autofimlog_id}" verb=DELETE {
  api_group = "RoboAtendimento"

  input {
    int alarmevent_autofimlog_id? filters=min:1
  }

  stack {
    db.del alarmEvent_autofimlog {
      field_name = "id"
      field_value = $input.alarmevent_autofimlog_id
    }
  }

  response = null
}