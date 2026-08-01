query InserirCliente verb=POST {
  api_group = "ConfMonitCliente"

  input {
    text Authorization? filters=trim
    object Dados? {
      schema {
        text idFranqueado? filters=trim
        text idPacote? filters=trim
        text nome? filters=trim
        text nick? filters=trim
        text documento1? filters=trim
        text telefone1? filters=trim
        email email1? filters=trim|lower
        text documento2? filters=trim
        text cep? filters=trim
        text endereco? filters=trim
        text complemento? filters=trim
        text bairro? filters=trim
        text cidade? filters=trim
        text uf? filters=trim
        text telefone2? filters=trim
        text email2? filters=trim
      }
    }
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/insere"
      method = "POST"
      params = $input.Dados
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}