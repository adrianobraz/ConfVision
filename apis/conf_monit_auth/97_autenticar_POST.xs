query Autenticar verb=POST {
  api_group = "ConfMonitAuth"

  input {
    text email? filters=trim
    text senha? filters=trim
  }

  stack {
    function.run ConfMonitAuth_Autenticar {
      input = {email: $input.email, senha: $input.senha}
    } as $func_1
  }

  response = $func_1
}