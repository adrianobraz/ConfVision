query eventos verb=GET {
  api_group = "testes"

  input {
  }

  stack {
    function.run func_eventos_terminal_atendimento as $func_1
  }

  response = $func_1
}