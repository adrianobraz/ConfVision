// Resumo licencas e acesso ConfVision para o franqueado
query fp_confvision_resumo_franqueado verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_confvision_acesso_efetivo {
      input = {id_franqueado: $input.id_franqueado}
    } as $acesso
  
    db.query vis_licenca {
      where = $db.vis_licenca.id_franqueado == $input.id_franqueado
      sort = {vis_licenca.created_at: "desc"}
      return = {type: "list"}
    } as $licencas
  
    var $pendente {
      value = 0
    }
  
    var $disponivel {
      value = 0
    }
  
    var $em_uso {
      value = 0
    }
  
    var $expirada {
      value = 0
    }
  
    foreach ($licencas) {
      each as $item {
        conditional {
          if ($item.status == "pendente") {
            var.update $pendente {
              value = $pendente + 1
            }
          }
        
          elseif ($item.status == "disponivel") {
            var.update $disponivel {
              value = $disponivel + 1
            }
          }
        
          elseif ($item.status == "em_uso") {
            var.update $em_uso {
              value = $em_uso + 1
            }
          }
        
          elseif ($item.status == "expirada") {
            var.update $expirada {
              value = $expirada + 1
            }
          }
        }
      }
    }
  
    db.query fp_fatura {
      where = $db.fp_fatura.id_franqueado == $input.id_franqueado && ($db.fp_fatura.tipo == "confvision_venda" || $db.fp_fatura.tipo == "confvision_renovacao") && $db.fp_fatura.status == "aberta"
      sort = {fp_fatura.vencimento_em: "asc"}
      return = {type: "list"}
    } as $faturas_abertas
  }

  response = {
    liberado       : $acesso.liberado
    motivo         : $acesso.motivo
    usa_confvision : $acesso.usa_confvision
    licencas       : $licencas
    resumo         : ```
      {
        pendente  : $pendente
        disponivel: $disponivel
        em_uso    : $em_uso
        expirada  : $expirada
        total     : $licencas|count
      }
      ```
    faturas_abertas: $faturas_abertas
  }
}