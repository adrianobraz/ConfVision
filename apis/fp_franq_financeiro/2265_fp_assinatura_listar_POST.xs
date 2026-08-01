// Listar assinaturas — CEN filtra id_central; REP filtra carteira
query fp_assinatura_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text status? filters=trim
    text produto? filters=trim
    text id_franqueado? filters=trim
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
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.status ==? $input.status && $db.fp_assinatura_produto.produto ==? $input.produto && $db.fp_assinatura_produto.id_franqueado ==? $input.id_franqueado
      sort = {fp_assinatura_produto.proxima_cobranca_em: "asc"}
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
          each as $ass {
            var $ok {
              value = false
            }
          
            conditional {
              if (($ass|get:"id_representante":"") == $id_rep) {
                var.update $ok {
                  value = true
                }
              }
            
              else {
                foreach ($ids) {
                  each as $id_fra {
                    conditional {
                      if (($ass|get:"id_franqueado":"") == $id_fra) {
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
              if ($ok) {
                array.push $lista {
                  value = $ass
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
          each as $ass {
            conditional {
              if ((($ass|get:"id_central":"")|trim) == $id_cen) {
                array.push $lista {
                  value = $ass
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
    userTipo : $user_tipo
    idCentral: $id_cen
  }
}
