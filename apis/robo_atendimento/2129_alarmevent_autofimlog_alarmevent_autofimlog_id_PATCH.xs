// Edit alarmEvent_autofimlog record
query "alarmevent_autofimlog/{alarmevent_autofimlog_id}" verb=PATCH {
  api_group = "RoboAtendimento"

  input {
    int alarmevent_autofimlog_id? filters=min:1
    dblink {
      table = "alarmEvent_autofimlog"
    }
  }

  stack {
    util.get_raw_input {
      encoding = "json"
      exclude_middleware = false
    } as $raw_input
  
    db.patch alarmEvent_autofimlog {
      field_name = "id"
      field_value = $input.alarmevent_autofimlog_id
      data = `$input|pick:($raw_input|keys)`|filter_null|filter_empty_text
    } as $alarmevent_autofimlog
  }

  response = $alarmevent_autofimlog
}