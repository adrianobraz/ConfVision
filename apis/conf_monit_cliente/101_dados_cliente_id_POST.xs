query DadosClienteID verb=POST {
  api_group = "ConfMonitCliente"

  input {
    text Authorization? filters=trim
    text idcliente? filters=trim
  }

  stack {
    function.run ConfMonitCliente_DadosClienteID {
      input = {
        Authorization: $input.Authorization
        idcliente    : $input.idcliente
      }
    } as $func_1
  }

  response = $func_1
}