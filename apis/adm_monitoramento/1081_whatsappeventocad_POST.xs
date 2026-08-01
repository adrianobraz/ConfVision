// Add WhatsappEventoCad record
query whatsappeventocad verb=POST {
  api_group = "admMonitoramento"

  input {
    text IdCliente? filters=trim
    text IdFranqueado? filters=trim
    text whatsapp? filters=trim
    text nome? filters=trim
    enum[] Tipo? {
      values = [
        "ALARME"
        "ARME"
        "DESARME"
        "EMERGENCIA"
        "FALHAS"
        "GERAL"
        "MEDICO"
        "PANICO"
        "SETUP"
        "TESTE"
      ]
    }
  
    text idDispositivo? filters=trim
    text NomeDispositivo? filters=trim
    bool audio?
  }

  stack {
    conditional {
      if (`$input.IdFranqueado|is_empty` == false) {
        db.add WhatsappEventoCad {
          enforce_hidden_fields = false
          data = {
            created_at     : "now"
            IdCliente      : $input.IdCliente
            IdFranqueado   : $input.IdFranqueado
            whatsapp       : $input.whatsapp
            nome           : $input.nome
            Tipo           : $input.Tipo
            idDispositivo  : $input.idDispositivo
            NomeDispositivo: $input.NomeDispositivo
            audio          : $input.audio
          }
        
          output = [
            "id"
            "created_at"
            "IdCliente"
            "IdFranqueado"
            "whatsapp"
            "nome"
            "Tipo"
            "idDispositivo"
            "NomeDispositivo"
            "audio"
          ]
        } as $model
      
        var.update $model {
          value = $model|safe_array
        }
      }
    
      else {
        db.query WhatsappEventoCad {
          where = $db.WhatsappEventoCad.IdCliente == $input.IdCliente
          return = {type: "list"}
        } as $model
      }
    }
  }

  response = $model
}