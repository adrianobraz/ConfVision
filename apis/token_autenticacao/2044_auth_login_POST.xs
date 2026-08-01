// Login and retrieve an authentication token
query "auth/login" verb=POST {
  api_group = "TokenAutenticacao"

  input {
    email email? filters=trim|lower
    text senha?
  }

  stack {
    db.get UserCliConfToken {
      field_name = "email"
      field_value = $input.email
      output = ["id", "created_at", "email", "senha"]
    } as $UserCliConfToken
  
    precondition ($UserCliConfToken != null) {
      error = "Invalid Credentials."
    }
  
    security.check_password {
      text_password = $input.senha
      hash_password = $UserCliConfToken.senha
    } as $pass_result
  
    precondition ($pass_result) {
      error = "Invalid Credentials."
    }
  
    security.create_auth_token {
      table = "UserCliConfToken"
      extras = {}
      expiration = 86400
      id = $UserCliConfToken.id
    } as $authToken
  }

  response = {authToken: $authToken}
}