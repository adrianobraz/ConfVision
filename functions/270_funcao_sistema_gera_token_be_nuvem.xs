function FuncaoSistema_GeraTokenBeNuvem {
  input {
    text email? filters=trim
    text password? filters=trim
  }

  stack {
    api.request {
      url = "https://app.benuvem.com.br/api/v1/auth/login"
      method = "POST"
      params = {}
        |set:"email":$input.email
        |set:"password":$input.password
      headers = []
        |push:"Content-Type: application/x-www-form-urlencoded"
    } as $api1
  }

  response = $api1
    |set:"":`$api1.response.result.access_token`
}