// Lista registros financeiros / contabeis (auditoria) no escopo CEN/REP
query fp_financeiro_log_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    text acao? filters=trim
    text produto? filters=trim
    text busca_admin? filters=trim
    int limite?=200 filters=min:1|max:500
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    db.query fp_financeiro_log {
      where = $db.fp_financeiro_log.id_franqueado ==? $input.id_franqueado && $db.fp_financeiro_log.acao ==? $input.acao && $db.fp_financeiro_log.produto ==? $input.produto && $db.fp_financeiro_log.admin_usuario ==? $input.busca_admin
      sort = {fp_financeiro_log.created_at: "desc"}
      return = {type: "list"}
    } as $lista_raw
  
    var $lista {
      value = []
    }
  
    conditional {
      if (($escopo|get:"userTipo":"") == "REP") {
        foreach ($lista_raw) {
          each as $l {
            var $ok {
              value = false
            }
          
            conditional {
              if (($l.id_representante|trim) == ($escopo|get:"idRepresentante":""|trim)) {
                var.update $ok {
                  value = true
                }
              }
            
              elseif (($l.admin_usuario|trim) == ($escopo|get:"usuario":""|trim)) {
                var.update $ok {
                  value = true
                }
              }
            
              elseif (($l.id_usuario|trim) == ($escopo|get:"idUsuario":""|trim)) {
                var.update $ok {
                  value = true
                }
              }
            }
          
            conditional {
              if ($ok) {
                array.push $lista {
                  value = $l
                }
              }
            }
          }
        }
      }
    
      elseif ($escopo|get:"permite_global":false) {
        var.update $lista {
          value = $lista_raw
        }
      }
    
      else {
        foreach ($lista_raw) {
          each as $l {
            conditional {
              if (($l.id_central|trim) == ($escopo|get:"idCentral":""|trim)) {
                array.push $lista {
                  value = $l
                }
              }
            
              elseif (($l.id_central|is_empty) && ($l.admin_usuario|trim) == ($escopo|get:"usuario":""|trim)) {
                array.push $lista {
                  value = $l
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
    idCentral: $escopo|get:"idCentral":""
  }
}
