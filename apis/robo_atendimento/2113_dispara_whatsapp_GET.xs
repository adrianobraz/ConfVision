query DisparaWhatsapp verb=GET {
  api_group = "RoboAtendimento"

  input {
  }

  stack {
    function.run fn_DisparaWhatsapp as $func_1
  }

  response = $func_1
}