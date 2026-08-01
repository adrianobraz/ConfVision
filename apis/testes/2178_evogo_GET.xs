query evogo verb=GET {
  api_group = "testes"

  input {
  }

  stack {
    api.request {
      url = "https://confianca-evolution-api.rkr351.easypanel.host/user/check"
      method = "POST"
      params = {}
        |set:"formatJid":false
        |set:"number":([]|push:"5519982387120")
      headers = []
        |push:"Content-Type: application/json"
        |push:"apikey:5c98ad37-d80e-4a44-ab73-263c809a6b99"
    } as $api1
  }

  response = $api1
}