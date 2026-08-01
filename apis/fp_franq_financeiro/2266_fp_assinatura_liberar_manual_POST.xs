// Libera assinatura: fatura paga (sync) ou fatura aberta (paga + libera) via Central de Liberacao
query fp_assinatura_liberar_manual verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int assinatura_id? filters=min:1
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
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
  
    function.run fn_fp_admin_assert_escopo_franqueado {
      input = {
        admin_token              : $input.admin_token
        id_franqueado            : $assinatura.id_franqueado
        id_representante_registro: $assinatura.id_representante
      }
    } as $escopo
  
    conditional {
      if ($assinatura.status == "ativa" && ($assinatura.valido_ate == null || $assinatura.valido_ate > now)) {
        var $model {
          value = $assinatura
        }
      
        var $log {
          value = null
        }
      
        function.run fn_fp_assinatura_aplicar_alacarte_pagamento {
          input = {
            assinatura_id: $assinatura.id
            valor_fatura : $assinatura.valor|first_notempty:0
          }
        } as $alacarte_sync_ativa
      
        var.update $model {
          value = $alacarte_sync_ativa.assinatura|first_notempty:$assinatura
        }
      
        conditional {
          if ($assinatura.produto == "franqueadopro") {
            function.run fn_fp_bundle_sincronizar {
              input = {
                id_franqueado: $assinatura.id_franqueado
                admin_usuario: $input.admin_usuario
              }
            } as $bundle_sync_ativa
          }
        }
      
        conditional {
          if ($assinatura.produto == "franqueadopro" || $assinatura.produto == "confvision") {
            function.run fn_franqueado_sync_usa_confvision {
              input = {id_franqueado: $assinatura.id_franqueado}
            } as $sync_cv_ativa
          }
        }
      
        conditional {
          if ($assinatura.produto == "franqueadopro" || $assinatura.produto == "webterminal" || $assinatura.produto == "terminalmovel" || $assinatura.produto == "webambiente") {
            function.run fn_franqueado_sync_usuario_terminal {
              input = {id_franqueado: $assinatura.id_franqueado}
            } as $sync_term_ativa
          }
        }
      }
    
      else {
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto" && $db.fp_fatura_item.ref_id == ($input.assinatura_id|to_text)
          return = {type: "list"}
        } as $itens
      
        var $fatura_paga {
          value = null
        }
      
        var $fatura_aberta {
          value = null
        }
      
        foreach ($itens) {
          each as $item {
            db.get fp_fatura {
              field_name = "id"
              field_value = $item.fp_fatura_id
            } as $f
          
            conditional {
              if ($f != null && $f.status == "paga" && $fatura_paga == null) {
                var.update $fatura_paga {
                  value = $f
                }
              }
            
              elseif ($f != null && $f.status == "aberta" && $fatura_aberta == null) {
                var.update $fatura_aberta {
                  value = $f
                }
              }
            }
          }
        }
      
        // Liberar na Central: se ha fatura aberta, registra pagamento e libera
        conditional {
          if ($fatura_paga == null && $fatura_aberta != null) {
            var $pago_em {
              value = now
            }
          
            var $valor_pago {
              value = $fatura_aberta.valor_total|first_notempty:($assinatura.valor|first_notempty:0)
            }
          
            db.add fp_pagamento {
              data = {
                created_at  : "now"
                fp_fatura_id: $fatura_aberta.id
                valor       : $valor_pago
                metodo      : "manual"
                pago_em     : $pago_em
                observacao  : $input.observacao|first_notempty:"Pagamento via Liberar (Central de Liberacao)"
              }
            } as $pagamento
          
            db.patch fp_fatura {
              field_name = "id"
              field_value = $fatura_aberta.id
              data = {
                status : "paga"
                pago_em: $pago_em
              }
            } as $fatura_paga
          
            function.run fn_fp_repasse_gerar_apos_pagamento {
              input = {
                fatura_id    : $fatura_aberta.id
                admin_usuario: $input.admin_usuario
              }
            } as $repasse_lib
          }
        }
      
        precondition ($fatura_paga != null) {
          error = "Nenhuma fatura encontrada para esta assinatura. Clique em Gerar fatura e tente Liberar de novo, ou registre o pagamento em Financeiro."
        }
      
        function.run fn_fp_assinatura_ativar_pos_pagamento {
          input = {
            assinatura_id: $input.assinatura_id
            pago_em      : $fatura_paga.pago_em|first_notempty:now
            valor_fatura : $fatura_paga.valor_total|first_notempty:0
          }
        } as $ativacao
      
        var $model {
          value = $ativacao.assinatura|first_notempty:$assinatura
        }
      
        function.run fn_fp_financeiro_log {
          input = {
            acao         : "assinatura_liberar_pos_pagamento"
            id_franqueado: $assinatura.id_franqueado
            ref_tipo     : "fp_assinatura_produto"
            ref_id       : $input.assinatura_id|to_text
            detalhe      : $input.observacao|first_notempty:"Liberacao com pagamento da fatura"
            origem       : "admin"
            valor        : $assinatura.valor|first_notempty:0
            produto      : $assinatura.produto
            plano        : $assinatura.plano
            admin_usuario: $input.admin_usuario
          }
        } as $log
      
        conditional {
          if ($assinatura.produto == "franqueadopro") {
            function.run fn_fp_bundle_sincronizar {
              input = {
                id_franqueado: $assinatura.id_franqueado
                admin_usuario: $input.admin_usuario
              }
            } as $bundle_sync
          }
        }
      
        conditional {
          if ($assinatura.produto == "franqueadopro" || $assinatura.produto == "confvision") {
            function.run fn_franqueado_sync_usa_confvision {
              input = {id_franqueado: $assinatura.id_franqueado}
            } as $sync_cv
          }
        }
      
        conditional {
          if ($assinatura.produto == "franqueadopro" || $assinatura.produto == "webterminal" || $assinatura.produto == "terminalmovel" || $assinatura.produto == "webambiente") {
            function.run fn_franqueado_sync_usuario_terminal {
              input = {id_franqueado: $assinatura.id_franqueado}
            } as $sync_term
          }
        }
      }
    }
  }

  response = {assinatura: $model, registro_financeiro: $log}
}