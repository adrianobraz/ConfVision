// Login and retrieve an authentication token
query login verb=POST {
  api_group = "Autenticacao"

  input {
    email email? filters=trim|lower
    text senha?
  }

  stack {
    db.get Usuario {
      field_name = "email"
      field_value = $input.email
      output = ["id", "created_at", "email", "senha"]
    } as $Usuario
  
    precondition ($Usuario != null) {
      error = "Invalid Credentials."
    }
  
    security.check_password {
      text_password = $input.senha
      hash_password = $Usuario.senha
    } as $pass_result
  
    precondition ($pass_result) {
      error = "Invalid Credentials."
    }
  
    security.create_auth_token {
      table = "Usuario"
      extras = {}
      expiration = 86400
      id = $Usuario.id
    } as $authToken
  }

  response = {authToken: $authToken}
}