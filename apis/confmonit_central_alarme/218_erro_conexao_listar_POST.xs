query ErroConexaoListar verb=POST {
  api_group = "ConfmonitCentralAlarme"

  input {
    text Authorization? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/erroConexao/listar"
      method = "POST"
      headers = []
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}