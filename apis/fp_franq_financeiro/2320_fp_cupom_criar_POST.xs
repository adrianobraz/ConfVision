// Cria cupom hierarquico — CEN→REP (mesma Central) ou REP→FRA (carteira)
query fp_cupom_criar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text codigo? filters=trim
    text tipo?=percentual filters=trim
    decimal valor?
    text produto?=franqueadopro filters=trim
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $criado_por_tipo {
      value = $escopo|get:"userTipo":""
    }
  
    var $alvo {
      value = "FRA"
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    conditional {
      if (($escopo|get:"userTipo":"") == "CEN") {
        var.update $alvo {
          value = "REP"
        }
      
        precondition (($id_cen|is_empty) == false) {
          error = "Sessao Central sem idCentral — faca logout e login novamente"
        }
      }
    
      elseif (($escopo|get:"userTipo":"") == "REP") {
        var.update $criado_por_tipo {
          value = "REP"
        }
      
        var.update $alvo {
          value = "FRA"
        }
      
        var.update $id_rep {
          value = $escopo|get:"idRepresentante":""
        }
      }
    
      else {
        precondition (false) {
          error = "Somente Central ou Representante podem criar cupons"
        }
      }
    }
  
    function.run fn_fp_cupom_criar {
      input = {
        codigo           : $input.codigo
        tipo             : $input.tipo
        valor            : $input.valor
        produto          : $input.produto
        id_franqueado    : $input.id_franqueado
        id_representante : $id_rep
        id_central       : $id_cen
        criado_por_tipo  : $criado_por_tipo
        alvo_nivel       : $alvo
        observacao       : $input.observacao
        admin_usuario    : $input.admin_usuario|first_notempty:$escopo|get:"usuario":""
      }
    } as $resultado
  }

  response = $resultado
}
