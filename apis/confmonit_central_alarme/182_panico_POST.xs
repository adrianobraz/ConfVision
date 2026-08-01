query Panico verb=POST {
  api_group = "ConfmonitCentralAlarme"

  input {
    text Authorization? filters=trim
    text Conta? filters=trim
    text Senha? filters=trim
    text IDFranqueado? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.3:2031/recebe-evento"
      method = "POST"
      params = {}
        |set:"idFranqueado":$input.IDFranqueado
        |set:"evento":($input.Conta|concat:"1M1300000":"")
        |set:"senha":$input.Senha
        |set:"emailOn":"SIM"
      headers = []
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
        |push:"Content-Type: application/json"
    } as $api1
  }

  response = $api1
}