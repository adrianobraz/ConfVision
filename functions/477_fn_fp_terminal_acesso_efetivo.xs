// Define se master deve ter UsuarioTeminal=S (WebTerminal, Terminal Movel ou webAmbiente liberados)
function fn_fp_terminal_acesso_efetivo {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    var $liberado {
      value = false
    }
  
    var $motivo {
      value = "sem_acesso_terminal"
    }
  
    var $origem {
      value = "nenhum"
    }
  
    var $produto {
      value = ""
    }
  
    function.run fn_fp_produto_acesso_efetivo {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : "webterminal"
      }
    } as $wt
  
    conditional {
      if ($wt.liberado) {
        var.update $liberado {
          value = true
        }
      
        var.update $motivo {
          value = "webterminal_" ~ ($wt.origem|first_notempty:"ativo")
        }
      
        var.update $origem {
          value = $wt.origem
        }
      
        var.update $produto {
          value = "webterminal"
        }
      }
    }
  
    conditional {
      if ($liberado == false) {
        function.run fn_fp_produto_acesso_efetivo {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "terminalmovel"
          }
        } as $tm
      
        conditional {
          if ($tm.liberado) {
            var.update $liberado {
              value = true
            }
          
            var.update $motivo {
              value = "terminalmovel_" ~ ($tm.origem|first_notempty:"ativo")
            }
          
            var.update $origem {
              value = $tm.origem
            }
          
            var.update $produto {
              value = "terminalmovel"
            }
          }
        }
      }
    }
  
    conditional {
      if ($liberado == false) {
        function.run fn_fp_produto_acesso_efetivo {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "webambiente"
          }
        } as $wa
      
        conditional {
          if ($wa.liberado) {
            var.update $liberado {
              value = true
            }
          
            var.update $motivo {
              value = "webambiente_" ~ ($wa.origem|first_notempty:"ativo")
            }
          
            var.update $origem {
              value = $wa.origem
            }
          
            var.update $produto {
              value = "webambiente"
            }
          }
        }
      }
    }
  
    var $usuario_terminal {
      value = "N"
    }
  
    conditional {
      if ($liberado) {
        var.update $usuario_terminal {
          value = "S"
        }
      }
    }
  }

  response = {
    liberado        : $liberado
    motivo          : $motivo
    origem          : $origem
    produto         : $produto
    usuario_terminal: $usuario_terminal
  }
}