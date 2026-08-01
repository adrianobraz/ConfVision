// Update WhatsEventCadFranq record
query "whatseventcadfranq/{whatseventcadfranq_id}" verb=PUT {
  api_group = "franqueado"

  input {
    int whatseventcadfranq_id? filters=min:1
    dblink {
      table = "WhatsEventCadFranq"
    }
  }

  stack {
    db.edit WhatsEventCadFranq {
      field_name = "id"
      field_value = $input.whatseventcadfranq_id
      enforce_hidden_fields = false
      data = {
        idFranqueado    : $input.idFranqueado
        telMonitoramento: $input.telMonitoramento
        telViatura      : $input.telViatura
        telGerente      : $input.telGerente
        horaIni         : $input.horaIni
        horaFin         : $input.horaFin
      }
    } as $model
  }

  response = $model
}