// Get the record belonging to the authentication token
query "auth/me" verb=GET {
  api_group = "TokenAutenticacao"
  auth = "UserCliConfToken"

  input {
  }

  stack {
    db.get UserCliConfToken {
      field_name = "id"
      field_value = $auth.id
      output = ["usuario", "password"]
    } as $UserCliConfToken
  }

  response = {dados: $UserCliConfToken}
}