query DadosClienteEmail verb=POST {
  api_group = "ConfMonitCliente"

  input {
    text Authorization? filters=trim
    email email1? filters=trim|lower
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2000/v4/cliente/getDadosByEmail"
      method = "POST"
      params = {}|set:"email1":$input.email1
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}