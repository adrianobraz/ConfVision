query sendWhatsEventCentralAlarm verb=POST {
  api_group = "FuncaoSistema"

  input {
    int idEvento?
  }

  stack {
    function.run FuncaoSistema_sendWhatsEventCentralAlarm {
      input = {idEvento: $input.idEvento}
    } as $func_1
  }

  response = $func_1
}