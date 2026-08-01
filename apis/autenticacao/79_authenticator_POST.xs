// Login and retrieve an authentication token
query authenticator verb=POST {
  api_group = "Autenticacao"
  auth = "Usuario"

  input {
    email email? filters=trim|lower
    text senha?
  }

  stack {
    db.get Usuario {
      field_name = "id"
      field_value = $auth.id
      output = ["id", "email", "senha"]
    } as $Usuario0
  
    db.get Usuario {
      field_name = "email"
      field_value = $input.email
      output = ["id", "email", "senha"]
    } as $Usuario1
  
    precondition ($Usuario1 != null && $Usuario0.id == $Usuario1.id) {
      error = "Invalid Credentials."
    }
  
    security.check_password {
      text_password = $input.senha
      hash_password = $Usuario1.senha
    } as $pass_result
  
    precondition ($pass_result) {
      error = "Invalid Credentials."
    }
  
    security.create_auth_token {
      table = "Usuario"
      extras = {}
      expiration = 86400
      id = $Usuario1.id
    } as $authToken
  }

  response = {authToken: $authToken}
}