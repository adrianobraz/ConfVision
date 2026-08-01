function evoSendTexto {
  input {
    text number? filters=trim
    text textMensagem? filters=trim
    text instancia? filters=trim
  }

  stack {
    api.request {
      url = "https://confianca-evolution-api.rkr351.easypanel.host/message/sendText/" ~ $input.instancia
      method = "POST"
      params = {}
        |set:"number":$input.number
        |set:"text":$input.textMensagem
        |set:"delay":0
        |set:"linkPreview":false
        |set:"mentionsEveryOne":false
      headers = []
        |push:"Content-Type: application/json"
        |push:"apikey: 429683C4C977415CAAFCCE10F7D57E11"
    } as $api1
  }

  response = $api1
}