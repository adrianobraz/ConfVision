// Resumo de licencas ConfVision do franqueado
query fp_confvision_resumo verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    db.query vis_licenca {
      where = $db.vis_licenca.id_franqueado == $input.id_franqueado
      return = {type: "list"}
    } as $lista
  
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
  
    foreach ($lista) {
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
  }

  response = {
    dados : $lista
    resumo: ```
      {
        pendente  : $pendente
        disponivel: $disponivel
        em_uso    : $em_uso
        expirada  : $expirada
        total     : $lista|count
      }
      ```
  }
}