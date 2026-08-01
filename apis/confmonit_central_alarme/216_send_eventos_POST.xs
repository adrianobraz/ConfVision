query SendEventos verb=POST {
  api_group = "ConfmonitCentralAlarme"
  auth = "Usuario"

  input {
    text Authorization? filters=trim
    text idFranqueado? filters=trim
    text evento? filters=trim
    text senha? filters=trim
    text emailOn? filters=trim
    text Conta? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.3:2031/recebe-evento"
      method = "POST"
      params = {}
        |set:"idFranqueado":$input.idFranqueado
        |set:"evento":($input.Conta|concat:$input.evento:"")
        |set:"senha":$input.senha
        |set:"emailOn":"SIM"
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}