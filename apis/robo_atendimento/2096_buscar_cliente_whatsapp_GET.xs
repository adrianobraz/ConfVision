query buscar_cliente_Whatsapp verb=GET {
  api_group = "RoboAtendimento"

  input {
    text telefone? filters=trim
  }

  stack {
    db.query WhatsappEventoCad {
      where = $db.WhatsappEventoCad.whatsapp == $input.telefone
      return = {type: "list"}
    } as $WhatsappEventoCad1
  
    !function.run WebLogarCache as $func2
    !function.run ConfMonitCliente_DadosClienteID {
      input = {
        Authorization: $func2
        idcliente    : $WhatsappEventoCad1.IdCliente
      }
    } as $func1|set:"":`$func1.response.result.dados`
  }

  response = $WhatsappEventoCad1|diff_assoc:$WhatsappEventoCad1
}