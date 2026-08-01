// Aplica cupom em fatura aberta (recalcula) ou paga (credito proxima cobranca). Nao toca ConfVision Pro+.
function fn_fp_cupom_aplicar_fatura {
  input {
    text codigo? filters=trim
    text id_franqueado? filters=trim
    int fatura_id? filters=min:1
    text origem?=franqueado filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    db.get fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
    } as $fatura
  
    precondition ($fatura != null) {
      error = "Fatura nao encontrada"
    }
  
    precondition ($fatura.id_franqueado == $input.id_franqueado) {
      error = "Fatura nao pertence a este franqueado"
    }
  
    // ConfVision: cupom nao interfere no beneficio Pro+
    conditional {
      if ($fatura.tipo == "confvision") {
        function.run fn_fp_assinatura_get_ativa {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "franqueadopro"
          }
        } as $fp
      
        conditional {
          if ($fp.liberado && $fp.assinatura != null && $fp.assinatura.plano == "pro_plus") {
            precondition (false) {
              error = "Cupom nao se aplica a faturas ConfVision para assinantes Pro+. O desconto de 20% do plano ja esta ativo."
            }
          }
        }
      }
    }
  
    var $produto_fatura {
      value = $fatura.tipo|first_notempty:"franqueadopro"
    }
  
    conditional {
      if ($produto_fatura == "assinatura") {
        var.update $produto_fatura {
          value = "franqueadopro"
        }
      }
    }
  
    var $valor_base {
      value = $fatura.valor_total|first_notempty:0
    }
  
    precondition ($valor_base > 0) {
      error = "Fatura sem valor para aplicar desconto"
    }
  
    // Piso congelado na fatura/assinatura — cupom REP nao pode furar
    var $piso {
      value = $fatura.valor_piso_central|first_notempty:0
    }
  
    conditional {
      if ($piso <= 0) {
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.fp_fatura_id == $fatura.id && $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto"
          return = {type: "list"}
        } as $itens_ass
      
        foreach ($itens_ass) {
          each as $it {
            conditional {
              if ($piso <= 0) {
                db.get fp_assinatura_produto {
                  field_name = "id"
                  field_value = $it.ref_id|to_int
                } as $ass_piso
              
                conditional {
                  if ($ass_piso != null && $ass_piso.valor_piso_central != null) {
                    var.update $piso {
                      value = $ass_piso.valor_piso_central
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  
    function.run fn_fp_cupom_aplicar {
      input = {
        codigo            : $input.codigo
        id_franqueado     : $input.id_franqueado
        produto           : $produto_fatura
        valor_base        : $valor_base
        valor_piso_minimo : $piso
        ref_tipo          : "fp_fatura"
        ref_id            : $fatura.id|to_text
        origem            : $input.origem|first_notempty:"franqueado"
        admin_usuario     : $input.admin_usuario
        contexto          : "franqueado"
      }
    } as $aplicado
  
    var $modo {
      value = "fatura_aberta"
    }
  
    var $credito {
      value = null
    }
  
    conditional {
      if ($fatura.status == "aberta") {
        db.patch fp_fatura {
          field_name = "id"
          field_value = $fatura.id
          data = {
            valor_total: $aplicado.valor_final
            observacao : (($fatura.observacao|first_notempty:"") ~ " | Cupom " ~ ($input.codigo|to_upper) ~ " -R$ " ~ ($aplicado.valor_desconto|to_text))|trim
          }
        } as $fatura_upd
      
        db.add fp_fatura_item {
          data = {
            created_at    : "now"
            fp_fatura_id  : $fatura.id
            descricao     : "Desconto cupom " ~ ($input.codigo|to_upper)
            quantidade    : 1
            valor_unitario: (0 - $aplicado.valor_desconto)
            valor_total   : (0 - $aplicado.valor_desconto)
            ref_tipo      : "fp_cupom_desconto"
            ref_id        : $aplicado.cupom.id|to_text
          }
        } as $item_desconto
      
        var.update $fatura {
          value = $fatura_upd
        }
      }
    
      elseif ($fatura.status == "paga") {
        var.update $modo {
          value = "credito_proxima_fatura"
        }
      
        function.run fn_fp_assinatura_get_ativa {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "franqueadopro"
          }
        } as $check_ass
      
        var $ass {
          value = $check_ass.assinatura
        }
      
        precondition ($ass != null) {
          error = "Nenhuma assinatura encontrada para creditar o desconto"
        }
      
        var $saldo_atual {
          value = $ass.credito_saldo|first_notempty:0
        }
      
        var $novo_saldo {
          value = ($saldo_atual + $aplicado.valor_desconto)|round:2
        }
      
        db.patch fp_assinatura_produto {
          field_name = "id"
          field_value = $ass.id
          data = {credito_saldo: $novo_saldo}
        } as $ass_upd
      
        var.update $credito {
          value = {
            assinatura_id     : $ass.id
            saldo_anterior    : $saldo_atual
            credito_adicionado: $aplicado.valor_desconto
            saldo_atual       : $novo_saldo
          }
        }
      
        function.run fn_fp_financeiro_log {
          input = {
            acao         : "cupom_credito"
            id_franqueado: $input.id_franqueado
            ref_tipo     : "fp_assinatura_produto"
            ref_id       : $ass.id|to_text
            detalhe      : "codigo=" ~ ($input.codigo|to_upper) ~ " credito=" ~ ($aplicado.valor_desconto|to_text) ~ " saldo=" ~ ($novo_saldo|to_text) ~ " fatura_paga=" ~ ($fatura.id|to_text)
            origem       : $input.origem|first_notempty:"franqueado"
            valor        : $aplicado.valor_desconto
            produto      : "franqueadopro"
            plano        : $ass.plano
            admin_usuario: $input.admin_usuario
          }
        } as $log_credito
      }
    
      else {
        precondition (false) {
          error = "Somente faturas abertas ou pagas aceitam cupom"
        }
      }
    }
  }

  response = {
    modo          : $modo
    fatura        : $fatura
    valor_base    : $aplicado.valor_base
    valor_desconto: $aplicado.valor_desconto
    valor_final   : $aplicado.valor_final
    cupom         : $aplicado.cupom
    credito       : $credito
  }
}