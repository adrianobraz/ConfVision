// Retorna assinatura ativa (ou mais recente) do franqueado para um produto
function fn_fp_assinatura_get_ativa {
  input {
    text id_franqueado? filters=trim
    text produto? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.produto|is_empty) == false) {
      error = "produto obrigatorio"
    }
  
    var $produto {
      value = $input.produto|trim|to_lower
    }
  
    var $id_franqueado {
      value = $input.id_franqueado|trim
    }
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.id_franqueado == $id_franqueado && $db.fp_assinatura_produto.produto == $produto && $db.fp_assinatura_produto.status == "ativa"
      sort = {fp_assinatura_produto.valido_ate: "desc"}
      return = {type: "single"}
    } as $ativa
  
    conditional {
      if ($ativa == null) {
        db.query fp_assinatura_produto {
          where = $db.fp_assinatura_produto.id_franqueado == $id_franqueado && $db.fp_assinatura_produto.produto == $produto
          sort = {fp_assinatura_produto.created_at: "desc"}
          return = {type: "single"}
        } as $ativa
      }
    }
  
    var $liberado {
      value = false
    }
  
    var $motivo {
      value = "sem_assinatura"
    }
  
    // Auto-cura: pendente com fatura paga → tenta ativar (sem derrubar a leitura se sync falhar)
    conditional {
      if ($ativa != null && $ativa.status == "pendente") {
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto" && $db.fp_fatura_item.ref_id == ($ativa.id|to_text)
          return = {type: "list"}
        } as $itens_fat
      
        var $fatura_paga {
          value = null
        }
      
        foreach ($itens_fat) {
          each as $item_fat {
            conditional {
              if ($fatura_paga == null) {
                db.get fp_fatura {
                  field_name = "id"
                  field_value = $item_fat.fp_fatura_id
                } as $f
              
                conditional {
                  if ($f != null && $f.status == "paga") {
                    var.update $fatura_paga {
                      value = $f
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($fatura_paga != null) {
            try_catch {
              try {
                function.run fn_fp_assinatura_ativar_pos_pagamento {
                  input = {
                    assinatura_id: $ativa.id
                    pago_em      : $fatura_paga.pago_em|first_notempty:now
                    valor_fatura : $fatura_paga.valor_total
                  }
                } as $sync_pag
              
                var.update $ativa {
                  value = $sync_pag.assinatura|first_notempty:$ativa
                }
              }
            
              catch {
                db.get fp_assinatura_produto {
                  field_name = "id"
                  field_value = $ativa.id
                } as $ativa_reload
              
                conditional {
                  if ($ativa_reload != null) {
                    var.update $ativa {
                      value = $ativa_reload
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  
    conditional {
      if ($ativa != null) {
        conditional {
          if ($ativa.status == "ativa") {
            conditional {
              if ($ativa.valido_ate == null || $ativa.valido_ate > now) {
                var.update $liberado {
                  value = true
                }
              
                var.update $motivo {
                  value = "ok"
                }
              }
            
              else {
                var.update $motivo {
                  value = "vencida"
                }
              }
            }
          }
        
          elseif ($ativa.status == "suspensa") {
            var.update $motivo {
              value = "suspensa"
            }
          }
        
          elseif ($ativa.status == "pendente") {
            var.update $motivo {
              value = "pendente"
            }
          }
        }
      }
    }
  }

  response = {
    assinatura: $ativa
    liberado  : $liberado
    motivo    : $motivo
  }
}
