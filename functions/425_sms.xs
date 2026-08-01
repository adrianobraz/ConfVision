function sms {
  input {
    text number? filters=trim
    text textMensagem? filters=trim
    int tblWhatsAppEnviados?
    text InstanceToken? filters=trim
    text tipoAPi? filters=trim
  }

  stack {
    api.lambda {
      code = """
        const texto = $input.textMensagem || "";
        
        let evento = "";
        
        const posInicio = texto.indexOf("Evento:");
        const posFim = texto.indexOf("Dispositivo:");
        
        if (posInicio >= 0) {
        
            evento = texto.substring(
                posInicio + 7,
                posFim >= 0 ? posFim : undefined
            ).trim();
        
        } else {
        
            // se não existir Evento:, usa o texto inteiro
            evento = texto.trim();
        }
        
        // remove acentos
        evento = evento
            .normalize("NFD")
            .replace(/[\u0300-\u036f]/g, "");
        
        // remove espaços duplicados
        evento = evento.replace(/\s+/g, " ").trim();
        
        // corta em 160
        if (evento.length > 160) {
            evento = evento.substring(0, 160);
        }
        
        return {
            evento
        };
        """
      timeout = 10
    } as $msgSMS
  
    api.request {
      url = "https://api.comtele.com.br/messages/sms/send"
      method = "POST"
      params = {}
        |set:"receivers":([]|push:$input.number)
        |set:"message":$msgSMS.evento
        |set:"tag":"API-sending"
        |set:"custom":$input.tblWhatsAppEnviados
        |set:"route":17
      headers = []
        |push:$input.InstanceToken
        |push:"Content-Type: application/json"
    } as $api1
  
    db.edit WhatsAppEnviados {
      field_name = "id"
      field_value = $input.tblWhatsAppEnviados
      enforce_hidden_fields = false
      data = {
        whats_ID        : "SMS"
        whats_messageid : "SMS"
        whats_sender    : "SMS"
        whats_senderName: "SMS"
        whats_text      : $msgSMS.evento
        texto           : true
        SMS             : true
      }
    } as $WhatsAppEnviados1
  }

  response = $api1
}