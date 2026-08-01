// Calcula saldo de caixa no escopo do usuario (CEN/REP)
// Entradas: pagamentos de faturas de receita (assinatura/confvision/etc)
// Saidas: pagamentos de repasse + retiradas + contas pagas do admin_usuario
function fn_fp_caixa_saldo {
  input {
    text id_central? filters=trim
    text id_representante? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    var $id_cen {
      value = $input.id_central|first_notempty:""|trim
    }
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""|trim
    }
  
    var $usuario {
      value = $input.admin_usuario|first_notempty:""|trim
    }
  
    db.query fp_fatura {
      where = $db.fp_fatura.id > 0
      return = {type: "list"}
    } as $faturas_raw
  
    var $ids_entrada {
      value = []
    }
  
    var $ids_repasse {
      value = []
    }
  
    foreach ($faturas_raw) {
      each as $f {
        var $ok {
          value = false
        }
      
        conditional {
          if (($id_rep|is_empty) == false) {
            conditional {
              if (($f.id_representante|trim) == $id_rep) {
                var.update $ok {
                  value = true
                }
              }
            }
          }
        
          elseif (($id_cen|is_empty) == false) {
            conditional {
              if (($f.id_central|trim) == $id_cen) {
                var.update $ok {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($ok) {
            var $tipo_f {
              value = $f.tipo|first_notempty:""|trim
            }
          
            conditional {
              if ($tipo_f == "repasse_rep_central" || $tipo_f == "repasse_central_breakglass" || ($tipo_f|contains:"repasse")) {
                array.push $ids_repasse {
                  value = $f.id
                }
              }
            
              else {
                array.push $ids_entrada {
                  value = $f.id
                }
              }
            }
          }
        }
      }
    }
  
    db.query fp_pagamento {
      where = $db.fp_pagamento.id > 0
      return = {type: "list"}
    } as $pagamentos_raw
  
    var $total_entradas {
      value = 0
    }
  
    var $total_repasses {
      value = 0
    }
  
    var $qtd_pagamentos {
      value = 0
    }
  
    foreach ($pagamentos_raw) {
      each as $p {
        var $pag_entrada {
          value = false
        }
      
        var $pag_repasse {
          value = false
        }
      
        foreach ($ids_entrada) {
          each as $fid {
            conditional {
              if ($pag_entrada == false && $p.fp_fatura_id == $fid) {
                var.update $pag_entrada {
                  value = true
                }
              }
            }
          }
        }
      
        foreach ($ids_repasse) {
          each as $fidr {
            conditional {
              if ($pag_repasse == false && $p.fp_fatura_id == $fidr) {
                var.update $pag_repasse {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($pag_entrada) {
            var.update $total_entradas {
              value = $total_entradas + ($p.valor|first_notempty:0)
            }
          
            var.update $qtd_pagamentos {
              value = $qtd_pagamentos + 1
            }
          }
        }
      
        conditional {
          if ($pag_repasse) {
            var.update $total_repasses {
              value = $total_repasses + ($p.valor|first_notempty:0)
            }
          
            var.update $qtd_pagamentos {
              value = $qtd_pagamentos + 1
            }
          }
        }
      }
    }
  
    var $retiradas {
      value = []
    }
  
    var $contas_pagas {
      value = []
    }
  
    conditional {
      if (($usuario|is_empty) == false) {
        db.query fp_caixa_movimento {
          where = $db.fp_caixa_movimento.tipo == "retirada" && $db.fp_caixa_movimento.admin_usuario == $usuario
          return = {type: "list"}
        } as $retiradas
      
        db.query fp_conta_pagar {
          where = $db.fp_conta_pagar.status == "paga" && $db.fp_conta_pagar.admin_usuario == $usuario
          return = {type: "list"}
        } as $contas_pagas
      }
    }
  
    var $total_retiradas {
      value = 0
    }
  
    var $total_contas_pagas {
      value = 0
    }
  
    foreach ($retiradas) {
      each as $r {
        var.update $total_retiradas {
          value = $total_retiradas + ($r.valor|first_notempty:0)
        }
      }
    }
  
    foreach ($contas_pagas) {
      each as $c {
        var.update $total_contas_pagas {
          value = $total_contas_pagas + ($c.valor|first_notempty:0)
        }
      }
    }
  
    var $total_saidas {
      value = $total_retiradas + $total_contas_pagas + $total_repasses
    }
  
    var $saldo {
      value = $total_entradas - $total_saidas
    }
  }

  response = {
    saldo             : $saldo
    total_entradas    : $total_entradas
    total_retiradas   : $total_retiradas
    total_contas_pagas: $total_contas_pagas
    total_repasses    : $total_repasses
    total_saidas      : $total_saidas
    qtd_pagamentos    : $qtd_pagamentos
    qtd_retiradas     : $retiradas|count
  }
}
