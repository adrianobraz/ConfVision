query WhatsappEventCadJanela verb=POST {
  api_group = "RoboAtendimento"

  input {
    int whatsappeventocad_id?
    enum[] dias? {
      values = ["0", "1", "2", "3", "4", "5", "6", "7"]
    }
  
    text hora_inicio? filters=trim
    text hora_fim? filters=trim
    text IdCliente? filters=trim
    text IdFranqueado? filters=trim
    text whatsapp? filters=trim
    text nome? filters=trim
    text idDispositivo? filters=trim
    text NomeDispositivo? filters=trim
    int id?
  }

  stack {
    !var $x1 {
      value = now
    }
  
    !var $x2 {
      value = ```
        ($x1|format_timestamp:"N":"America/Sao_Paulo") ~ " - " ~
        ($x1|to_timestamp|format_timestamp:"H:i":"America/Sao_Paulo")
        ```
    }
  
    db.get WhatsappEventCadJanela {
      field_name = "id"
      field_value = $input.id
    } as $WhatsappEventCadJanela2
  
    conditional {
      if ($WhatsappEventCadJanela2|is_empty) {
        db.add WhatsappEventCadJanela {
          enforce_hidden_fields = false
          data = {
            whatsappeventocad_id: $input.whatsappeventocad_id
            dias                : $input.dias
            hora_inicio         : $input.hora_inicio
            hora_fim            : $input.hora_fim
            IdCliente           : $input.IdCliente
            IdFranqueado        : $input.IdFranqueado
            whatsapp            : $input.whatsapp
            nome                : $input.nome
            idDispositivo       : $input.idDispositivo
            NomeDispositivo     : $input.NomeDispositivo
          }
        } as $WhatsappEventCadJanela1
      }
    
      else {
        db.edit WhatsappEventCadJanela {
          field_name = "id"
          field_value = $input.id
          enforce_hidden_fields = false
          data = {
            dias       : $input.dias
            hora_inicio: $input.hora_inicio
            hora_fim   : $input.hora_fim
          }
        } as $WhatsappEventCadJanela1
      }
    }
  }

  response = {dados: $WhatsappEventCadJanela1}
}