query EnviaMsgAtendente verb=POST {
  api_group = "ConfMonitCliente"

  input {
    text Authorization? filters=trim
    text idDispositivo? filters=trim
    text msgAtendente? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/dispositivo/setMsgAtendenteById"
      method = "POST"
      params = {}
        |set:"idDispositivo":$input.idDispositivo
        |set:"msgAtendente":$input.msgAtendente
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $api1
  }

  response = $api1
}