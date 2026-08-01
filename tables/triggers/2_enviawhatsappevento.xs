table_trigger enviawhatsappevento {
  table = "alarm_events"

  input {
    json new
    json old
    enum action {
      values = ["insert", "update", "delete", "truncate"]
    }
  
    text datasource
  }

  stack {
    conditional {
      if (($input.new|is_empty) == false) {
        function.run FuncaoSistema_sendWhatsEventCentralAlarm {
          input = {idEvento: $input.new.id}
        } as $func1
      }
    }
  }

  actions = {insert: true}
  datasources = ["live"]
}