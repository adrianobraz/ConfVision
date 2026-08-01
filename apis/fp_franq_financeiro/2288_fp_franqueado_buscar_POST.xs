// Buscar franqueados — REP so ve a propria carteira
query fp_franqueado_buscar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text busca? filters=trim
    int limite?
    text id_representante? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    var $id_rep {
      value = ""
    }
  
    conditional {
      if ($admin.userTipo == "REP") {
        var.update $id_rep {
          value = $admin|get:"idVinculo":($admin|get:"idRepresentante":"")|trim
        }
      
        conditional {
          if (($id_rep|is_empty) == true) {
            var.update $id_rep {
              value = $input.id_representante|trim
            }
          }
        }
      
        precondition (($id_rep|is_empty) == false) {
          error = "Sessao Representante sem idVinculo — faca logout e login novamente"
        }
      }
    
      elseif (($input.id_representante|is_empty) == false) {
        // CEN / break-glass: filtra franqueados da carteira do representante escolhido
        var.update $id_rep {
          value = $input.id_representante|trim
        }
      }
    }
  
    function.run fn_fp_franqueado_buscar {
      input = {
        busca           : $input.busca
        limite          : $input.limite
        id_representante: $id_rep
      }
    } as $lista
  }

  response = {
    dados   : $lista.dados
    total   : $lista.total
    userTipo: $admin.userTipo
  }
}
