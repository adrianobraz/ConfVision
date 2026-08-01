query resetsenha verb=POST {
  api_group = "ConfMonitAuth"

  input {
    text idcliente? filters=trim
    text Authorization? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/resetarSenha"
      method = "POST"
      params = {}|set:"idCliente":$input.idcliente
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}