query whatsappeventocad2 verb=POST {
  api_group = "admMonitoramento"

  input {
    text IdCliente? filters=trim
    text? IdFrank2? filters=trim
    text whatsapp? filters=trim
    text nome? filters=trim
    text idDispositivo? filters=trim
    text NomeDispositivo? filters=trim
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
        "GRADEHORARIO"
        "INTERNET"
        "ENERGIA"
      ]
    }
  
    bool EnviaAudio?
    bool ligar?
    bool texto?
    bool notificarsempre?
    int ordemligacao?
    text nomeCliente? filters=trim
    text NomeFranqueado? filters=trim
  }

  stack {
    conditional {
      if ($input.IdFrank2 != null) {
        api.request {
          url = "https://monitoramento.uazapi.com/chat/check"
          method = "POST"
          params = {}
            |set:"numbers":([]|push:$input.whatsapp)
          headers = []
            |push:"Accept: application/json"
            |push:"Content-Type: application/json"
            |push:"token: b014aad8-99e1-4115-9873-7d004ab91dc7"
        } as $api1
      
        conditional {
          if ($api1.response.result[0].isInWhatsapp) {
            db.add WhatsappEventoCad {
              enforce_hidden_fields = false
              data = {
                created_at     : "now"
                IdCliente      : $input.IdCliente
                nomeCliente    : $input.nomeCliente
                IdFranqueado   : $input.IdFrank2
                nomefranqueado : $input.NomeFranqueado
                whatsapp       : $input.whatsapp
                nome           : $input.nome
                Tipo           : $input.Tipo
                idDispositivo  : $input.idDispositivo
                NomeDispositivo: $input.NomeDispositivo
                audio          : $input.EnviaAudio
                ligar          : $input.ligar
                texto          : $input.texto
                notificarsempre: $input.notificarsempre
                ordemligacao   : $input.ordemligacao
              }
            } as $model
          
            var.update $model {
              value = $model|safe_array
            }
          }
        
          else {
            var $model {
              value = []
            }
          }
        }
      }
    
      else {
        db.query WhatsappEventoCad {
          where = $db.WhatsappEventoCad.IdCliente == $input.IdCliente
          sort = {whatsappeventocad.ordemligacao: "asc"}
          return = {type: "list"}
        } as $model
      }
    }
  }

  response = $model
}