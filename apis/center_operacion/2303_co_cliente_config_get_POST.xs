// Busca configuração visual do cliente no Centro Operacional
query co_cliente_config_get verb=POST {
  api_group = "centerOperacion"

  input {
    text idCliente filters=trim
    text idFranqueado? filters=trim
  }

  stack {
    precondition (($input.idCliente|is_empty) == false) {
      error = "idCliente obrigatorio"
    }
  
    db.query co_cliente_config {
      where = $db.co_cliente_config.idCliente == $input.idCliente
      return = {type: "single"}
    } as $config
  }

  response = $config
}