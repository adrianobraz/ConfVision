function usercliconflog {
  input {
    dblink {
      table = "UserCliConfLog"
      override = {Data: {hidden: true}}
    }
  }

  stack {
    db.add UserCliConfLog {
      enforce_hidden_fields = false
      data = {
        created_at    : "now"
        usercliconf_id: $input.usercliconf_id
        idFranqueado  : $input.idFranqueado
        idCliente     : $input.idCliente
        Descricao     : $input.Descricao
        Ocorrencia    : $input.Ocorrencia
        Evento        : $input.Evento
        log           : $input.log
        idDisp        : $input.idDisp
        DispApp       : $input.DispApp
      }
    } as $model
  }

  response = $model
}