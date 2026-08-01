// Movimentacao financeira (logs) no escopo CEN/REP / usuario
query fp_fin_movimentacao_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
    text natureza? filters=trim
    int limite?=200 filters=min:1|max:500
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    db.query fp_financeiro_log {
      where = $db.fp_financeiro_log.id > 0
      sort = {fp_financeiro_log.created_at: "desc"}
      return = {type: "list"}
    } as $logs_raw
  
    var $logs {
      value = []
    }
  
    conditional {
      if (($escopo|get:"userTipo":"") == "REP") {
        foreach ($logs_raw) {
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
                array.push $logs {
                  value = $l
                }
              }
            }
          }
        }
      }
    
      elseif ($escopo|get:"permite_global":false) {
        var.update $logs {
          value = $logs_raw
        }
      }
    
      else {
        foreach ($logs_raw) {
          each as $l {
            conditional {
              if (($l.id_central|trim) == ($escopo|get:"idCentral":""|trim)) {
                array.push $logs {
                  value = $l
                }
              }
            
              elseif (($l.id_central|is_empty) && ($l.admin_usuario|trim) == ($escopo|get:"usuario":""|trim)) {
                array.push $logs {
                  value = $l
                }
              }
            }
          }
        }
      }
    }
  
    var $comp {
      value = $input.competencia|first_notempty:""
    }
  
    var $dados {
      value = []
    }
  
    var $entradas {
      value = 0
    }
  
    var $saidas {
      value = 0
    }
  
    foreach ($logs) {
      each as $l {
        var $ok {
          value = true
        }
      
        conditional {
          if (($comp|strlen) >= 7 && $l.created_at != null) {
            var.update $ok {
              value = false
            }
          
            conditional {
              if (($l.created_at|format_timestamp:"Y-m":"UTC") == $comp) {
                var.update $ok {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($ok) {
            var $nat {
              value = "neutro"
            }
          
            conditional {
              if ($l.acao == "fatura_pagamento") {
                var $prod_log {
                  value = $l.produto|first_notempty:""|trim
                }
              
                conditional {
                  if ($prod_log == "repasse_rep_central" || $prod_log == "repasse_central_breakglass" || ($prod_log|contains:"repasse")) {
                    var.update $nat {
                      value = "saida"
                    }
                  
                    var.update $saidas {
                      value = $saidas + ($l.valor|first_notempty:0)
                    }
                  }
                
                  else {
                    var.update $nat {
                      value = "entrada"
                    }
                  
                    var.update $entradas {
                      value = $entradas + ($l.valor|first_notempty:0)
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($l.acao == "caixa_retirada" || $l.acao == "conta_pagar_pagar" || $l.acao == "repasse_pagamento") {
                var.update $nat {
                  value = "saida"
                }
              
                var.update $saidas {
                  value = $saidas + ($l.valor|first_notempty:0)
                }
              }
            }
          
            var $filtro_ok {
              value = true
            }
          
            conditional {
              if (($input.natureza|strlen) > 0 && $input.natureza != $nat) {
                var.update $filtro_ok {
                  value = false
                }
              }
            }
          
            conditional {
              if ($filtro_ok) {
                array.push $dados {
                  value = {
                    data         : $l.created_at
                    tipo         : $l.acao
                    natureza     : $nat
                    valor        : $l.valor
                    id_franqueado: $l.id_franqueado
                    descricao    : $l.detalhe
                    produto      : $l.produto
                    plano        : $l.plano
                    admin_usuario: $l.admin_usuario
                    ref_tipo     : $l.ref_tipo
                    ref_id       : $l.ref_id
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
    dados      : $dados
    total      : $dados|count
    entradas   : $entradas
    saidas     : $saidas
    competencia: $comp
    idCentral  : $escopo|get:"idCentral":""
  }
}
