query ArmeDesarme verb=POST {
  api_group = "ConfmonitCentralAlarme"
  auth = "Usuario"

  input {
    text Authorization? filters=trim
    text idDispositivo? filters=trim
    text usuario? filters=trim
    text senha? filters=trim
    text senhaWeb? filters=trim
    int numero?
    int acao?
  }

  stack {
    api.request {
      url = "http://185.130.61.3:2030/armar"
      method = "POST"
      params = {}
        |set:"idDispositivo":$input.idDispositivo
        |set:"numero":$input.numero
        |set:"usuario":$input.usuario
        |set:"acao":$input.acao
        |set:"senha":$input.senha
        |set:"senhaWeb":$input.senhaWeb
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}