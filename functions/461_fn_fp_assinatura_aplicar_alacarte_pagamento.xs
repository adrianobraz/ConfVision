// Aplica modulos efetivos apos pagamento de fatura a la carte
function fn_fp_assinatura_aplicar_alacarte_pagamento {
  input {
    int assinatura_id? filters=min:1
    decimal valor_fatura?
  }

  stack {
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    } as $assinatura
  
    precondition ($assinatura != null) {
      error = "Assinatura nao encontrada"
    }
  
    var $model {
      value = $assinatura
    }
  
    var $pagamento_parcial {
      value = false
    }
  
    var $lista_addons {
      value = $assinatura.addons_json|first_notempty:[]
    }
  
    var $pendentes {
      value = $assinatura.addons_pendentes_json|first_notempty:[]
    }
  
    foreach ($pendentes) {
      each as $ch_pend {
        var $ja_na_lista {
          value = false
        }
      
        foreach ($lista_addons) {
          each as $ch_add {
            conditional {
              if ($ch_add == $ch_pend) {
                var.update $ja_na_lista {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($ja_na_lista == false) {
            var.update $lista_addons {
              value = $lista_addons|push:$ch_pend
            }
          }
        }
      }
    }
  
    function.run fn_fp_alacarte_addons_normalizar {
      input = {produto: $assinatura.produto, addons: $lista_addons}
    } as $norm
  
    var.update $lista_addons {
      value = $norm.addons|first_notempty:[]
    }
  
    var $eh_alacarte {
      value = $assinatura.tipo_contratacao == "alacarte" || ($lista_addons|count) > 0 || ($pendentes|count) > 0
    }
  
    conditional {
      if ($eh_alacarte) {
        var $valor_ass {
          value = $assinatura.valor|first_notempty:0
        }
      
        var $valor_pago {
          value = $input.valor_fatura|first_notempty:0
        }
      
        conditional {
          if (($pendentes|count) > 0 && $valor_pago > 0 && $valor_pago < $valor_ass) {
            var.update $pagamento_parcial {
              value = true
            }
          }
        }
      
        function.run fn_fp_alacarte_calcular {
          input = {
            produto   : $assinatura.produto
            plano     : $assinatura.plano
            id_central: $assinatura.id_central|first_notempty:"CENTRAL"
            addons    : $lista_addons
          }
        } as $calc
      
        db.patch fp_assinatura_produto {
          field_name = "id"
          field_value = $assinatura.id
          data = {
            modulos_json         : $calc.modulos_json
            limites_json         : $calc.limites_json
            valor                : $calc.valor_total
            tipo_contratacao     : "alacarte"
            addons_json          : $lista_addons
            addons_pendentes_json: []
          }
        } as $model
      }
    }
  }

  response = {
    assinatura       : $model
    pagamento_parcial: $pagamento_parcial
  }
}