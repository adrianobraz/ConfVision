// Get WhatsappEventCadJanela Delete
query "whatsappeventcadjanela/{id}" verb=DELETE {
  api_group = "RoboAtendimento"

  input {
    int id? filters=min:1
  }

  stack {
    db.del WhatsappEventCadJanela {
      field_name = "id"
      field_value = $input.id
    }
  }

  response = null
}