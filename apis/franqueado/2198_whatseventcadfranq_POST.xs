// Add WhatsEventCadFranq record
query whatseventcadfranq verb=POST {
  api_group = "franqueado"

  input {
    dblink {
      table = "WhatsEventCadFranq"
    }
  }

  stack {
    db.add WhatsEventCadFranq {
      enforce_hidden_fields = false
      data = {
        created_at      : "now"
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