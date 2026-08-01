query testeopenai verb=GET {
  api_group = "testes"

  input {
  }

  stack {
    function.run OpenAi_Texto as $func_1
  }

  response = $func_1
}