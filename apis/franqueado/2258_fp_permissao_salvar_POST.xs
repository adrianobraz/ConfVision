// Substitui todas as permissoes de um responsavel (delete + insert)
query fp_permissao_salvar verb=POST {
  api_group = "franqueado"

  input {
    // ID do franqueado
    text idFranqueado filters=trim
  
    // ID do responsavel
    text idUsuario filters=trim
  
    // Lista de permissoes {chave_menu, liberado}
    object[] permissoes? {
      schema {
        text chave_menu filters=trim
        text liberado filters=trim
      }
    }
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
    } as $existentes
  
    foreach ($existentes) {
      each as $item {
        db.del fp_usuario_permissao {
          field_name = "id"
          field_value = $item.id
        }
      }
    }
  
    var $inseridos {
      value = 0
    }
  
    conditional {
      if ($input.permissoes != null) {
        foreach ($input.permissoes) {
          each as $perm {
            conditional {
              if (($perm.chave_menu|is_empty) == false) {
                db.add fp_usuario_permissao {
                  data = {
                    created_at   : "now"
                    id_franqueado: $input.idFranqueado
                    id_usuario   : $input.idUsuario
                    chave_menu   : $perm.chave_menu
                    liberado     : $perm.liberado|first_notempty:"S"
                  }
                } as $novo
              
                var.update $inseridos {
                  value = $inseridos + 1
                }
              }
            }
          }
        }
      }
    }
  }

  response = {inseridos: $inseridos, status: "OK"}
}