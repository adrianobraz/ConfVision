// Delete WhatsappEventoCad record
query "whatsappeventocad/{whatsappeventocad_id}" verb=DELETE {
  api_group = "admMonitoramento"

  input {
    int whatsappeventocad_id? filters=min:1
  }

  stack {
    db.del WhatsappEventoCad {
      field_name = "id"
      field_value = $input.whatsappeventocad_id
    }
  }

  response = null
}