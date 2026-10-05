// Ativa vis_licenca pendente vinculada a faturas ja pagas (ref_id = id Postgres)
function fn_fp_confvision_reconciliar_licencas_pagas {
  input {
    text id_franqueado? filters=trim
    int fp_fatura_id?
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    var $ativadas {
      value = 0
    }

    var $faturas {
      value = []
    }

    conditional {
      if ($input.fp_fatura_id != null && $input.fp_fatura_id > 0) {
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $input.id_franqueado && $db.fp_fatura.status == "paga" && $db.fp_fatura.id == $input.fp_fatura_id
          return = {type: "list"}
        } as $faturas_filtro

        var.update $faturas {
          value = $faturas_filtro
        }
      }

      else {
        db.query fp_fatura {
          where = $db.fp_fatura.id_franqueado == $input.id_franqueado && $db.fp_fatura.status == "paga"
          return = {type: "list"}
        } as $faturas_todas

        var.update $faturas {
          value = $faturas_todas
        }
      }
    }

    foreach ($faturas) {
      each as $fatura {
        var $pago_em {
          value = $fatura.pago_em|first_notempty:now
        }

        db.query fp_pagamento {
          where = $db.fp_pagamento.fp_fatura_id == $fatura.id
          sort = {fp_pagamento.id: "desc"}
          return = {type: "single"}
        } as $pag

        var $pag_id {
          value = 0
        }

        conditional {
          if ($pag != null) {
            var.update $pag_id {
              value = $pag.id|first_notempty:0
            }
          }
        }

        db.query fp_fatura_item {
          where = $db.fp_fatura_item.fp_fatura_id == $fatura.id && $db.fp_fatura_item.ref_tipo == "vis_licenca"
          return = {type: "list"}
        } as $itens

        foreach ($itens) {
          each as $item {
            conditional {
              if (($item.ref_id|is_empty) == false) {
                var $lic_id {
                  value = $item.ref_id|to_int
                }

                conditional {
                  if ($lic_id > 0) {
                    try_catch {
                      try {
                        function.run fn_fp_confvision_ativar_licenca_pagamento {
                          input = {
                            vis_licenca_id : $lic_id
                            fp_fatura_id   : $fatura.id
                            fp_pagamento_id: $pag_id
                            pago_em        : $pago_em
                          }
                        } as $lic_upd

                        var.update $ativadas {
                          value = $ativadas + 1
                        }
                      }

                      catch {
                        // continua proxima licenca (ex.: ja disponivel)
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  response = {ativadas: $ativadas}
}
