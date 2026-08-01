// Suspende assinaturas com fatura aberta vencida (worker ou admin)
function fn_fp_fatura_suspender_vencidas {
  input {
    text origem?=worker filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    db.query fp_fatura {
      where = $db.fp_fatura.status == "aberta" && $db.fp_fatura.vencimento_em != null && $db.fp_fatura.vencimento_em < now
      return = {type: "list"}
    } as $faturas_vencidas
  
    var $suspensas {
      value = 0
    }
  
    foreach ($faturas_vencidas) {
      each as $fatura {
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.fp_fatura_id == $fatura.id && $db.fp_fatura_item.ref_tipo == "fp_assinatura_produto"
          return = {type: "list"}
        } as $itens
      
        foreach ($itens) {
          each as $item {
            db.get fp_assinatura_produto {
              field_name = "id"
              field_value = $item.ref_id|to_int
            } as $ass
          
            conditional {
              if ($ass != null && ($ass.status == "ativa" || $ass.status == "pendente")) {
                db.patch fp_assinatura_produto {
                  field_name = "id"
                  field_value = $ass.id
                  data = {
                    status    : "suspensa"
                    observacao: "Suspensa automaticamente — fatura vencida " ~ ($fatura.referencia|first_notempty:"")
                  }
                } as $ass_upd
              
                var.update $suspensas {
                  value = $suspensas + 1
                }
              
                function.run fn_fp_financeiro_log {
                  input = {
                    acao         : "assinatura_suspender_vencida"
                    id_franqueado: $ass.id_franqueado
                    ref_tipo     : "fp_assinatura_produto"
                    ref_id       : $ass.id|to_text
                    detalhe      : "fatura_id=" ~ ($fatura.id|to_text)
                    origem       : $input.origem|first_notempty:"worker"
                    valor        : $fatura.valor_total
                    produto      : $ass.produto
                    plano        : $ass.plano
                    admin_usuario: $input.admin_usuario
                  }
                } as $log_item
              
                conditional {
                  if ($ass.produto == "franqueado_pro" && $ass.plano == "pro_plus") {
                    function.run "" {
                      input = {id_franqueado: $ass.id_franqueado}
                    } as $sync_pro
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($fatura.tipo == "confvision_renovacao") {
            db.query fp_fatura_item {
              where = $db.fp_fatura_item.fp_fatura_id == $fatura.id && $db.fp_fatura_item.ref_tipo == "vis_licenca"
              return = {type: "list"}
            } as $itens_cv
          
            foreach ($itens_cv) {
              each as $item_cv {
                db.get vis_licenca {
                  field_name = "id"
                  field_value = $item_cv.ref_id|to_int
                } as $lic
              
                conditional {
                  if ($lic != null && ($lic.status == "disponivel" || $lic.status == "em_uso")) {
                    db.patch vis_licenca {
                      field_name = "id"
                      field_value = $lic.id
                      data = {
                        observacao: "Renovacao em atraso — fatura vencida " ~ ($fatura.referencia|first_notempty:"")
                      }
                    } as $lic_obs
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  response = {suspensas: $suspensas}
}