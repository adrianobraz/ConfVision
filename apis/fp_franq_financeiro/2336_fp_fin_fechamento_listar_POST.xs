// Lista fechamentos mensais no escopo da Central
query fp_fin_fechamento_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    var $perm_global {
      value = $escopo|get:"permite_global":false
    }
  
    db.query fp_fechamento_mes {
      where = $db.fp_fechamento_mes.id > 0
      sort = {fp_fechamento_mes.competencia: "desc"}
      return = {type: "list"}
    } as $lista_raw
  
    var $lista {
      value = []
    }
  
    conditional {
      if ($perm_global == true) {
        var.update $lista {
          value = $lista_raw
        }
      }
    
      else {
        foreach ($lista_raw) {
          each as $f {
            conditional {
              if ((($f|get:"id_central":"")|trim) == $id_cen) {
                array.push $lista {
                  value = $f
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    dados    : $lista
    total    : $lista|count
    idCentral: $id_cen
  }
}
