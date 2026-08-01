// Update WhatsEventCadFranq record
query "whatseventcadfranq/franqueado/{idfranq}" verb=PUT {
  api_group = "franqueado"

  input {
    dblink {
      table = "WhatsEventCadFranq"
    }
  
    text idfranq? filters=trim
  }

  stack {
    precondition (($input.idfranq|is_empty) == false) {
      error = "Franqueado Vazio"
    }
  
    db.get WhatsEventCadFranq {
      field_name = "idFranqueado"
      field_value = $input.idfranq
    } as $WhatsEventCadFranq1
  
    conditional {
      if ($WhatsEventCadFranq1|is_empty) {
        db.add WhatsEventCadFranq {
          enforce_hidden_fields = false
          data = {
            created_at      : "now"
            idFranqueado    : $input.idfranq
            telMonitoramento: $input.telMonitoramento
            telViatura      : $input.telViatura
            telGerente      : $input.telGerente
            horaIni         : $input.horaIni
            horaFin         : $input.horaFin
          }
        } as $model
      }
    
      else {
        db.edit WhatsEventCadFranq {
          field_name = "idFranqueado"
          field_value = $input.idfranq
          enforce_hidden_fields = false
          data = {
            telMonitoramento: $input.telMonitoramento
            telViatura      : $input.telViatura
            telGerente      : $input.telGerente
            horaIni         : $input.horaIni
            horaFin         : $input.horaFin
          }
        } as $model
      }
    }
  }

  response = {dados: $model}
}