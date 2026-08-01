query AlteraSenha verb=POST {
  api_group = "ConfMonitAuth"

  input {
    text idCliente? filters=trim
    text senha? filters=trim
    text Authorization? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2000/v4/cliente/alterarSenhaById"
      method = "POST"
      params = {}
        |set:"idCliente":$input.idCliente
        |set:"senha":$input.senha
      headers = []
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
        |push:"Content-Type: application/json"
    } as $api1
  }

  response = $api1
}