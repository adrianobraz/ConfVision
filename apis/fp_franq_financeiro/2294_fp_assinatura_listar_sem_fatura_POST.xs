// Admin — assinaturas ativas sem fatura do ciclo (escopo CEN/REP)
query fp_assinatura_listar_sem_fatura verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text produto? filters=trim
    text id_franqueado? filters=trim
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
  
    var $ids_fra {
      value = []
    }
  
    conditional {
      if ($user_tipo == "REP") {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $id_rep}
        } as $carteira
      
        var.update $ids_fra {
          value = $carteira|get:"ids":[]
        }
      }
    }
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.status == "ativa"
      sort = {fp_assinatura_produto.proxima_cobranca_em: "asc"}
      return = {type: "list"}
    } as $candidatas_raw
  
    var $candidatas {
      value = []
    }
  
    conditional {
      if ($perm_global == true) {
        var.update $candidatas {
          value = $candidatas_raw
        }
      }
    
      elseif ($user_tipo == "REP") {
        foreach ($candidatas_raw) {
          each as $item {
            var $ok {
              value = false
            }
          
            conditional {
              if (($item|get:"id_representante":"") == $id_rep) {
                var.update $ok {
                  value = true
                }
              }
            
              else {
                foreach ($ids_fra) {
                  each as $idf {
                    conditional {
                      if (($item|get:"id_franqueado":"") == $idf) {
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
                array.push $candidatas {
                  value = $item
                }
              }
            }
          }
        }
      }
    
      else {
        foreach ($candidatas_raw) {
          each as $item {
            conditional {
              if ((($item|get:"id_central":"")|trim) == $id_cen) {
                array.push $candidatas {
                  value = $item
                }
              }
            }
          }
        }
      }
    }
  
    db.query fp_fatura {
      return = {type: "list"}
    } as $todas_faturas_raw
  
    var $todas_faturas {
      value = []
    }
  
    conditional {
      if ($perm_global == true) {
        var.update $todas_faturas {
          value = $todas_faturas_raw
        }
      }
    
      elseif ($user_tipo == "REP") {
        foreach ($todas_faturas_raw) {
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
            
              else {
                foreach ($ids_fra) {
                  each as $idf {
                    conditional {
                      if (($f|get:"id_franqueado":"") == $idf) {
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
                array.push $todas_faturas {
                  value = $f
                }
              }
            }
          }
        }
      }
    
      else {
        foreach ($todas_faturas_raw) {
          each as $f {
            conditional {
              if ((($f|get:"id_central":"")|trim) == $id_cen) {
                array.push $todas_faturas {
                  value = $f
                }
              }
            }
          }
        }
      }
    }
  
    var $pendentes {
      value = []
    }
  
    foreach ($candidatas) {
      each as $item {
        var $incluir {
          value = true
        }
      
        conditional {
          if ($item.proxima_cobranca_em == null) {
            var.update $incluir {
              value = false
            }
          }
        }
      
        conditional {
          if (($input.produto|is_empty) == false) {
            conditional {
              if ($item.produto != $input.produto) {
                var.update $incluir {
                  value = false
                }
              }
            }
          }
        }
      
        conditional {
          if (($input.id_franqueado|is_empty) == false) {
            conditional {
              if ($item.id_franqueado != $input.id_franqueado) {
                var.update $incluir {
                  value = false
                }
              }
            }
          }
        }
      
        conditional {
          if ($incluir) {
            function.run fn_fp_assinatura_ciclo_ref {
              input = {
                assinatura_id: $item.id
                data_cobranca: $item.proxima_cobranca_em
              }
            } as $ciclo_ref
          
            var $fatura_existe {
              value = null
            }
          
            foreach ($todas_faturas) {
              each as $f {
                conditional {
                  if ($fatura_existe == null && $f.ciclo_ref == $ciclo_ref && $f.status != "cancelada") {
                    var.update $fatura_existe {
                      value = $f
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($fatura_existe == null) {
                var.update $pendentes {
                  value = $pendentes
                    |push:{assinatura: $item, ciclo_ref: $ciclo_ref}
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    dados    : $pendentes
    total    : $pendentes|count
    idCentral: $id_cen
  }
}
