query sms verb=GET {
  api_group = "testes"

  input {
    text telefone? filters=trim
    text evento? filters=trim
    text dispositivo? filters=trim
    text status? filters=trim
    text link_acesso? filters=trim
    text id_mensagem? filters=trim
  }

  stack {
    !function.run sms {
      input = {
        telefone   : $input.telefone
        evento     : $input.evento
        dispositivo: $input.dispositivo
        status     : $input.status
        link_acesso: $input.link_acesso
        id_mensagem: $input.id_mensagem
      }
    } as $func_1
  
    db.query WhatsappEventoCad {
      return = {type: "list"}
    } as $WhatsappEventoCad1
  
    foreach ($WhatsappEventoCad1) {
      each as $item {
        db.edit WhatsappEventoCad {
          field_name = "id"
          field_value = $item.id
          enforce_hidden_fields = false
          data = {
            Tipo: []|push:"ALARME"|push:"PANICO"|push:"GRADEHORARIO"
          }
        } as $WhatsappEventoCad2
      }
    }
  }

  response = $WhatsappEventoCad1
}