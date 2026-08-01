function ConfMonitCliente_DadosClienteID {
  input {
    text Authorization? filters=trim
    text idcliente? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/getDadosById"
      method = "POST"
      params = {}|set:"idcliente":$input.idcliente
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}