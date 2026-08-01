// Relatorio de repasses REP→Central e Central→Break-glass (escopo da Central)
query fp_repasse_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text status? filters=trim
    text id_representante? filters=trim
    text tipo? filters=trim
    int limite?=200 filters=min:1|max:500
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    conditional {
      if (($escopo|get:"userTipo":"") == "REP") {
        var.update $id_rep {
          value = $escopo|get:"idRepresentante":""
        }
      }
    }
  
    var $tipo {
      value = $input.tipo|first_notempty:"repasse_rep_central"|trim
    }
  
    var $lista {
      value = []
    }
  
    conditional {
      if ($tipo == "repasse_central_breakglass") {
        precondition (($escopo|get:"breakglass":false) == true || ($escopo|get:"userTipo":"") == "CEN") {
          error = "Sem permissao para listar repasse Break-glass"
        }
      
        db.query fp_fatura {
          where = $db.fp_fatura.tipo == "repasse_central_breakglass" && $db.fp_fatura.status ==? $input.status
          sort = {fp_fatura.created_at: "desc"}
          return = {type: "list"}
        } as $lista_raw
      
        conditional {
          if ($escopo|get:"permite_global":false) {
            var.update $lista {
              value = $lista_raw
            }
          }
        
          else {
            foreach ($lista_raw) {
              each as $f {
                conditional {
                  if (($f.id_central|trim) == $id_cen) {
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
    
      else {
        db.query fp_fatura {
          where = $db.fp_fatura.tipo == "repasse_rep_central" && $db.fp_fatura.status ==? $input.status && $db.fp_fatura.id_representante ==? $id_rep
          sort = {fp_fatura.created_at: "desc"}
          return = {type: "list"}
        } as $lista_raw
      
        conditional {
          if ($escopo|get:"permite_global":false) {
            var.update $lista {
              value = $lista_raw
            }
          }
        
          else {
            foreach ($lista_raw) {
              each as $f {
                conditional {
                  if (($f.id_central|trim) == $id_cen) {
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
    }
  }

  response = {
    dados    : $lista
    total    : $lista|count
    userTipo : $escopo|get:"userTipo":""
    idCentral: $id_cen
  }
}
