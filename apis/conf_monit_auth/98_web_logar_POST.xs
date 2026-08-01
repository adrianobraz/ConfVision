query WebLogar verb=POST {
  api_group = "ConfMonitAuth"

  input {
  }

  stack {
    function.run ConfMonitAuth_WebLogar as $func_1
  }

  response = $func_1
}