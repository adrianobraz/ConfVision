function uazapi_EnvioMsgAudio {
  input {
    text number? filters=trim
    text textMensagem? filters=trim
    int tblWhatsAppEnviados?
  }

  stack {
    api.request {
      url = "https://confianca-n8n.rkr351.easypanel.host/webhook/enviaaudiowhats"
      method = "POST"
      params = {}
        |set:"telefone":$input.number
        |set:"mensagem":$input.textMensagem
        |set:"tblWhatsAppEnviado":$input.tblWhatsAppEnviados
      headers = []
        |push:"Content-Type: application/json"
      timeout = 10
    } as $api1
  
    !db.edit WhatsAppEnviados {
      field_name = "id"
      field_value = $input.tblWhatsAppEnviados
      enforce_hidden_fields = false
      data = {
        whats_ID              : $api1.response.result.data.key.id
        whats_messageTimestamp: $api1.response.result.data.messageTimestamp
        whats_messageid       : $api1.response.result.data.instanceId
        whats_sender          : $api1.response.result.data.key.remoteJid
        audio                 : true
      }
    } as $WhatsAppEnviados1
  }

  response = $api1
}