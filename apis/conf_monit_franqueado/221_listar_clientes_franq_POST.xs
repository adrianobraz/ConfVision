query ListarClientesFranq verb=POST {
  api_group = "ConfMonitFranqueado"

  input {
    text Authorization? filters=trim
    text idFranqueado? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/listarByIdFranqueado"
      method = "POST"
      params = {}
        |set:"idFranqueado":$input.idFranqueado
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}