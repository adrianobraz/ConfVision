function ConfMonitAuth_Autenticar {
  input {
    text email? filters=trim
    text senha? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/cliente/logar"
      method = "POST"
      params = {}
        |set:"email1":$input.email
        |set:"senha":$input.senha
      headers = []
        |push:"Content-Type: application/json"
    } as $api1|set:"":`$api1.response.result.dados`
  }

  response = $api1
}