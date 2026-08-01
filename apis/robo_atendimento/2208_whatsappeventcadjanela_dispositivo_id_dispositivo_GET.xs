// Get WhatsappEventCadJanela record
query "whatsappeventcadjanela/dispositivo/{idDispositivo}" verb=GET {
  api_group = "RoboAtendimento"

  input {
    text idDispositivo? filters=trim
  }

  stack {
    db.get WhatsappEventCadJanela {
      field_name = "idDispositivo"
      field_value = $input.idDispositivo
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = {dados: $model}
}