// Update WhatsEventCadBloq record
query "whatseventcadbloq/{whatseventcadbloq_id}" verb=PUT {
  api_group = "franqueado"

  input {
    int whatseventcadbloq_id? filters=min:1
    dblink {
      table = "WhatsEventCadBloq"
    }
  }

  stack {
    db.edit WhatsEventCadBloq {
      field_name = "id"
      field_value = $input.whatseventcadbloq_id
      enforce_hidden_fields = false
      data = {
        idFranqueado: $input.idFranqueado
        idCliente   : $input.idCliente
        nomeCliente : $input.nomeCliente
      }
    } as $model
  }

  response = $model
}