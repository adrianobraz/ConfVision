// Lista permissoes de menu de um responsavel do franqueado
query fp_permissao_listar_by_usuario verb=POST {
  api_group = "franqueado"

  input {
    // ID do franqueado
    text idFranqueado filters=trim
  
    // ID do responsavel
    text idUsuario filters=trim
  }

  stack {
    precondition (($input.idFranqueado|is_empty) == false) {
      error = "idFranqueado obrigatorio"
    }
  
    precondition (($input.idUsuario|is_empty) == false) {
      error = "idUsuario obrigatorio"
    }
  
    db.query fp_usuario_permissao {
      where = $db.fp_usuario_permissao.id_franqueado == $input.idFranqueado && $db.fp_usuario_permissao.id_usuario == $input.idUsuario
      return = {type: "list"}
    } as $lista
  }

  response = $lista
}