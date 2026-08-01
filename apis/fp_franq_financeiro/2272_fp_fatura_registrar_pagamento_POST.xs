// Registrar pagamento manual e renovar assinaturas/licencas vinculadas
query fp_fatura_registrar_pagamento verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int fatura_id? filters=min:1
    decimal valor?
    text metodo?=manual filters=trim
    text observacao? filters=trim
    text id_externo? filters=trim
    timestamp? pago_em?
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    db.get fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
    } as $fatura
  
    precondition ($fatura != null) {
      error = "Fatura nao encontrada"
    }
  
    precondition ($fatura.status == "aberta") {
      error = "Somente faturas abertas podem receber pagamento"
    }
  
    // REP: so paga faturas da carteira ou o proprio repasse (sem id_franqueado)
    conditional {
      if ($admin_check.userTipo == "REP") {
        var $tipo_check {
          value = $fatura.tipo|first_notempty:""|trim
        }
      
        conditional {
          if ($tipo_check == "repasse_rep_central" || $tipo_check == "repasse_central_breakglass") {
            precondition ($fatura.id_representante == $admin_check.idVinculo) {
              error = "Repasse fora da carteira deste Representante"
            }
          }
        
          else {
            function.run fn_fp_admin_assert_escopo_franqueado {
              input = {
                admin_token              : $input.admin_token
                id_franqueado            : $fatura.id_franqueado
                id_representante_registro: $fatura.id_representante
              }
            } as $escopo_pag
          }
        }
      }
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
  
    var $valor_pago {
      value = $input.valor
    }
  
    conditional {
      if ($valor_pago == null || $valor_pago == 0) {
        var.update $valor_pago {
          value = $fatura.valor_total
        }
      }
    }
  
    db.add fp_pagamento {
      data = {
        created_at  : "now"
        fp_fatura_id: $input.fatura_id
        valor       : $valor_pago
        metodo      : $input.metodo
        pago_em     : $pago
        id_externo  : $input.id_externo
        observacao  : $input.observacao
      }
    } as $pagamento
  
    db.patch fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
      data = {status: "paga", pago_em: $pago}
    } as $fatura_upd
  
    db.query fp_fatura_item {
      where = $db.fp_fatura_item.fp_fatura_id == $input.fatura_id
      return = {type: "list"}
    } as $itens
  
    foreach ($itens) {
      each as $item {
        conditional {
          if ($item.ref_tipo == "fp_assinatura_produto") {
            db.get fp_assinatura_produto {
              field_name = "id"
              field_value = $item.ref_id|to_int
            } as $ass
          
            conditional {
              if ($ass != null) {
                function.run fn_fp_assinatura_ativar_pos_pagamento {
                  input = {
                    assinatura_id: $ass.id
                    pago_em      : $pago
                    valor_fatura : $valor_pago
                  }
                } as $ass_sync
              }
            }
          }
        }
      
        conditional {
          if ($item.ref_tipo == "vis_licenca") {
            function.run fn_fp_confvision_ativar_licenca_pagamento {
              input = {
                vis_licenca_id : $item.ref_id|to_int
                fp_fatura_id   : $input.fatura_id
                fp_pagamento_id: $pagamento.id
                pago_em        : $pago
              }
            } as $lic_upd
          }
        }
      }
    }
  
    // Sync franqueado soh para faturas de assinatura/licenca (repasses nao tem id_franqueado)
    var $sync_cv {
      value = null
    }
  
    var $sync_term {
      value = null
    }
  
    conditional {
      if ((($fatura.id_franqueado|trim)|is_empty) == false) {
        function.run fn_franqueado_sync_usa_confvision {
          input = {id_franqueado: $fatura.id_franqueado}
        } as $sync_cv
      
        function.run fn_franqueado_sync_usuario_terminal {
          input = {id_franqueado: $fatura.id_franqueado}
        } as $sync_term
      }
    }
  
    var $detalhe_log {
      value = ($input.metodo|first_notempty:"manual") ~ " " ~ ($valor_pago|to_text)
    }
  
    var $acao_log {
      value = "fatura_pagamento"
    }
  
    var $tipo_pag {
      value = $fatura.tipo|first_notempty:""|trim
    }
  
    conditional {
      if ($tipo_pag == "repasse_rep_central" || $tipo_pag == "repasse_central_breakglass" || ($tipo_pag|contains:"repasse")) {
        var.update $acao_log {
          value = "repasse_pagamento"
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao            : $acao_log
        id_franqueado   : $fatura.id_franqueado
        id_representante: $fatura.id_representante
        id_central      : $fatura.id_central
        ref_tipo        : "fp_fatura"
        ref_id          : $input.fatura_id|to_text
        detalhe         : $detalhe_log
        origem          : "admin"
        valor           : $valor_pago
        produto         : $fatura.tipo
        admin_usuario   : $input.admin_usuario
      }
    } as $log
  
    // Split: gera fatura REP→Central (piso) quando pagamento e de franqueado (nao em repasses)
    var $repasse_result {
      value = null
    }
  
    var $tipo_fat {
      value = $fatura.tipo|first_notempty:""|trim
    }
  
    conditional {
      if ($tipo_fat != "repasse_rep_central" && $tipo_fat != "repasse_central_breakglass") {
        function.run fn_fp_repasse_gerar_apos_pagamento {
          input = {
            fatura_id    : $input.fatura_id
            admin_usuario: $input.admin_usuario
          }
        } as $repasse_result
      }
    }
  }

  response = {
    fatura         : $fatura_upd
    pagamento      : $pagamento
    sync_confvision: $sync_cv
    repasse        : $repasse_result
  }
}
