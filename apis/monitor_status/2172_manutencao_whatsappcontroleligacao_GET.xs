query "manutencao/whatsappcontroleligacao" verb=GET {
  api_group = "MonitorStatus"

  input {
  }

  stack {
    db.query whatsappcontroleligacao {
      return = {type: "list"}
    } as $whatsappcontroleligacao1
  }

  response = {dados: $whatsappcontroleligacao1}
}