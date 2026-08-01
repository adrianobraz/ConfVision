query FrandDadosID verb=POST {
  api_group = "ConfMonitFranqueado"

  input {
    text Authorization? filters=trim
    text fraid? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/franqueado/getDadosById"
      method = "POST"
      params = {}|set:"fraid":$input.fraid
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}