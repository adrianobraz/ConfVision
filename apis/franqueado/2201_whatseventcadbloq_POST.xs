// Add WhatsEventCadBloq record
query whatseventcadbloq verb=POST {
  api_group = "franqueado"

  input {
    dblink {
      table = "WhatsEventCadBloq"
    }
  }

  stack {
    db.query WhatsEventCadBloq {
      where = $db.WhatsEventCadBloq.idFranqueado == $input.idFranqueado && $db.WhatsEventCadBloq.idCliente == $input.idCliente
      return = {type: "single"}
    } as $WhatsEventCadBloq1
  
    precondition ($WhatsEventCadBloq1|is_empty) {
      error = "Registro ja Adicionado"
    }
  
    db.add WhatsEventCadBloq {
      enforce_hidden_fields = false
      data = {
        created_at  : "now"
        idFranqueado: $input.idFranqueado
        idCliente   : $input.idCliente
        nomeCliente : $input.nomeCliente
      }
    } as $model
  }

  response = $model
}