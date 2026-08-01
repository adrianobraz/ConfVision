query ListaEquipamentos verb=POST {
  api_group = "ConfMonitEquipamento"

  input {
    text Authorization? filters=trim
    text idCliente? filters=trim
  }

  stack {
    function.run ListaEquipamentos {
      input = {
        Authorization: $input.Authorization
        idCliente    : $input.idCliente
      }
    } as $func_1
  }

  response = $func_1
}