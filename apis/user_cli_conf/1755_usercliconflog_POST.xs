query usercliconflog verb=POST {
  api_group = "UserCliConf"

  input {
    dblink {
      table = "UserCliConfLog"
      override = {Data: {hidden: true}}
    }
  }

  stack {
    function.run usercliconflog {
      input = {
        usercliconf_id: $input.usercliconf_id
        idFranqueado  : $input.idFranqueado
        idCliente     : $input.idCliente
        Data          : $input.Data
        Descricao     : $input.Descricao
        Ocorrencia    : $input.Ocorrencia
        Evento        : $input.Evento
        log           : $input.log
        idDisp        : $input.idDisp
        DispApp       : $input.DispApp
      }
    } as $func_1
  }

  response = $func_1
}