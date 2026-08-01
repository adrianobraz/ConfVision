// Delete whatsappcontroleligacao record
query "manutencao/whatsappcontroleligacao/{whatsappcontroleligacao_id}" verb=DELETE {
  api_group = "MonitorStatus"

  input {
    int whatsappcontroleligacao_id? filters=min:1
  }

  stack {
    db.del whatsappcontroleligacao {
      field_name = "id"
      field_value = $input.whatsappcontroleligacao_id
    }
  }

  response = {dados: $input.whatsappcontroleligacao_id}
}