// Salva ou atualiza cor/iniciais do avatar do cliente no Centro Operacional
query co_cliente_config_salvar verb=POST {
  api_group = "centerOperacion"

  input {
    text idCliente filters=trim
    text idFranqueado? filters=trim
    text corAvatar? filters=trim
    text iniciais? filters=trim
  }

  stack {
    precondition (($input.idCliente|is_empty) == false) {
      error = "idCliente obrigatorio"
    }
  
    db.query co_cliente_config {
      where = $db.co_cliente_config.idCliente == $input.idCliente
      return = {type: "single"}
    } as $existente
  
    conditional {
      if ($existente != null) {
        db.edit co_cliente_config {
          field_name = "id"
          field_value = $existente.id
          data = {
            idFranqueado: $input.idFranqueado|first_notempty:$existente.idFranqueado
            corAvatar   : $input.corAvatar|first_notempty:$existente.corAvatar
            iniciais    : $input.iniciais|first_notempty:$existente.iniciais
          }
        } as $atualizado
      }
    
      else {
        db.add co_cliente_config {
          data = {
            created_at  : "now"
            idCliente   : $input.idCliente
            idFranqueado: $input.idFranqueado
            corAvatar   : $input.corAvatar|first_notempty:"#2563eb"
            iniciais    : $input.iniciais
          }
        } as $atualizado
      }
    }
  }

  response = $atualizado
}