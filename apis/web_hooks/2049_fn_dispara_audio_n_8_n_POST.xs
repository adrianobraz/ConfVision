query fnDisparaAudioN8n verb=POST {
  api_group = "WebHooks"

  input {
    int tblWhatsAppEnviados?
    text Whats_ID? filters=trim
    text WhatsMessageTimeStamp? filters=trim
    text WhatsMessageID? filters=trim
    text Whats_sender? filters=trim
  }

  stack {
    db.edit WhatsAppEnviados {
      field_name = "id"
      field_value = $input.tblWhatsAppEnviados
      enforce_hidden_fields = false
      data = {
        whats_ID              : $input.Whats_ID
        whats_messageTimestamp: $input.WhatsMessageTimeStamp
        whats_messageid       : $input.WhatsMessageID
        whats_sender          : $input.Whats_sender
        audio                 : true
      }
    } as $WhatsAppEnviados1
  }

  response = null
}