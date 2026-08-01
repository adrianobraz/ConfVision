// Ativa assinatura apos pagamento de fatura (pacote ou alacarte)
// Status=ativa e valido_ate sao gravados PRIMEIRO; passos secundarios nao podem reverter a liberacao
function fn_fp_assinatura_ativar_pos_pagamento {
  input {
    int assinatura_id? filters=min:1
    timestamp? pago_em?
    decimal valor_fatura?
  }

  stack {
    precondition ($input.assinatura_id != null) {
      error = "assinatura_id obrigatorio"
    }
  
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    } as $assinatura
  
    precondition ($assinatura != null) {
      error = "Assinatura nao encontrada"
    }
  
    var $pago {
      value = $input.pago_em
    }
  
    conditional {
      if ($pago == null) {
        var.update $pago {
          value = now
        }
      }
    }
  
    function.run fn_fp_calcular_proxima_cobranca {
      input = {
        base_em      : now
        periodicidade: $assinatura.periodicidade|first_notempty:"mensal"
      }
    } as $calc
  
    function.run fn_fp_calcular_proxima_cobranca {
      input = {
        base_em      : $calc.proxima_cobranca_em
        periodicidade: $assinatura.periodicidade|first_notempty:"mensal"
      }
    } as $prox
  
    // Critico: libera acesso antes de syncs/bundle (falhas secundarias nao devem bloquear)
    db.patch fp_assinatura_produto {
      field_name = "id"
      field_value = $assinatura.id
      data = {
        status             : "ativa"
        valido_ate         : $calc.proxima_cobranca_em
        proxima_cobranca_em: $prox.proxima_cobranca_em
        ultima_cobranca_em : $pago
        ciclo_fatura_ref   : ""
      }
    } as $ass_upd
  
    try_catch {
      try {
        function.run fn_fp_assinatura_aplicar_catalogo {
          input = {assinatura_id: $assinatura.id}
        } as $cat_upd
      }
    
      catch {
        var $err_cat {
          value = true
        }
      }
    }
  
    try_catch {
      try {
        function.run fn_fp_assinatura_aplicar_alacarte_pagamento {
          input = {
            assinatura_id: $assinatura.id
            valor_fatura : $input.valor_fatura|first_notempty:0
          }
        } as $alacarte_upd
      }
    
      catch {
        var $err_alacarte {
          value = true
        }
      }
    }
  
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $assinatura.id
    } as $model
  
    try_catch {
      try {
        conditional {
          if (($model.addons_pendentes_json|count) > 0) {
            db.patch fp_assinatura_produto {
              field_name = "id"
              field_value = $assinatura.id
              data = {addons_pendentes_json: []}
            } as $model
          }
        }
      }
    
      catch {
        var $err_addons {
          value = true
        }
      }
    }
  
    try_catch {
      try {
        conditional {
          if ($model.produto == "franqueadopro") {
            function.run fn_fp_bundle_sincronizar {
              input = {id_franqueado: $model.id_franqueado}
            } as $bundle_sync
          }
        }
      }
    
      catch {
        var $err_bundle {
          value = true
        }
      }
    }
  
    try_catch {
      try {
        conditional {
          if ($model.produto == "franqueadopro" || $model.produto == "confvision") {
            function.run fn_franqueado_sync_usa_confvision {
              input = {id_franqueado: $model.id_franqueado}
            } as $sync_cv
          }
        }
      }
    
      catch {
        var $err_cv {
          value = true
        }
      }
    }
  
    try_catch {
      try {
        conditional {
          if ($model.produto == "franqueadopro" || $model.produto == "webterminal" || $model.produto == "terminalmovel" || $model.produto == "webambiente") {
            function.run fn_franqueado_sync_usuario_terminal {
              input = {id_franqueado: $model.id_franqueado}
            } as $sync_term
          }
        }
      }
    
      catch {
        var $err_term {
          value = true
        }
      }
    }
  
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $assinatura.id
    } as $model
  }

  response = {assinatura: $model, ativada: true}
}
