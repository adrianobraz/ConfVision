// Get WhatsappEventCadJanela record
query "whatsappeventcadjanela/{whatsappeventcadjanela_id}" verb=GET {
  api_group = "RoboAtendimento"

  input {
    int whatsappeventcadjanela_id? filters=min:1
  }

  stack {
    db.get WhatsappEventCadJanela {
      field_name = "whatsappeventocad_id"
      field_value = $input.whatsappeventcadjanela_id
    } as $model
  
    precondition ($model != null) {
      error_type = "notfound"
      error = "Not Found"
    }
  }

  response = {dados: $model}
}