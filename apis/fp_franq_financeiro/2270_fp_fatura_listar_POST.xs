// Listar faturas — CEN filtra id_central; REP filtra carteira
query fp_fatura_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    text status? filters=trim
    text tipo? filters=trim
    text id_representante? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $user_tipo {
      value = $escopo|get:"userTipo":""
    }
  
    var $id_rep {
      value = $escopo|get:"idRepresentante":""
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    var $perm_global {
      value = $escopo|get:"permite_global":false
    }
  
    db.query fp_fatura {
      where = ($db.fp_fatura.id > 0) && ($db.fp_fatura.id_franqueado ==? $input.id_franqueado) && ($db.fp_fatura.status ==? $input.status) && ($db.fp_fatura.tipo ==? $input.tipo)
      sort = {fp_fatura.status: "asc", fp_fatura.id: "desc"}
      return = {type: "list"}
    } as $lista_raw
  
    var $lista {
      value = []
    }
  
    conditional {
      if ($user_tipo == "REP") {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $id_rep}
        } as $carteira
      
        var $ids {
          value = $carteira|get:"ids":[]
        }
      
        foreach ($lista_raw) {
          each as $f {
            var $ok {
              value = false
            }
          
            conditional {
              if (($f|get:"id_representante":"") == $id_rep) {
                var.update $ok {
                  value = true
                }
              }
            
              elseif ((($f|get:"id_franqueado":"")|is_empty) == false) {
                foreach ($ids) {
                  each as $id_fra {
                    conditional {
                      if (($f|get:"id_franqueado":"") == $id_fra) {
                        var.update $ok {
                          value = true
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if (($f|get:"tipo":"") == "repasse_rep_central" && ($f|get:"id_representante":"") == $id_rep) {
                var.update $ok {
                  value = true
                }
              }
            }
          
            conditional {
              if ($ok) {
                array.push $lista {
                  value = $f
                }
              }
            }
          }
        }
      }
    
      elseif ($perm_global == true) {
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
    dados          : $lista
    total          : $lista|count
    userTipo       : $user_tipo
    idCentral      : $id_cen
    permite_global : $perm_global

  }
}
