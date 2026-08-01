// Listar licencas do franqueado com resumo por status
query vis_licenca_by_franqueado verb=GET {
  api_group = "confVision"

  input {
    text id_franqueado? filters=trim
    text status? filters=trim
    text unidade? filters=trim
  }

  stack {
    db.query vis_licenca {
      where = $db.vis_licenca.id_franqueado == $input.id_franqueado && $db.vis_licenca.status ==? $input.status && $db.vis_licenca.unidade ==? $input.unidade
      sort = {vis_licenca.created_at: "desc"}
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