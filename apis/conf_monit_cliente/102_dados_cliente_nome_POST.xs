query DadosClienteNome verb=POST {
  api_group = "ConfMonitCliente"

  input {
    text Authorization? filters=trim
    text nome? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2000/v4/cliente/getDadosByName"
      method = "POST"
      params = {}|set:"nome":$input.nome
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}