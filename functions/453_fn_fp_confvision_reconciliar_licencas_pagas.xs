// Ativa vis_licenca pendente vinculada a faturas ja pagas (recuperacao)
function fn_fp_confvision_reconciliar_licencas_pagas {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    var $ativadas {
      value = 0
    }
  
    db.query fp_fatura {
      where = $db.fp_fatura.id_franqueado == $input.id_franqueado && $db.fp_fatura.status == "paga"
      return = {type: "list"}
    } as $faturas
  
    foreach ($faturas) {
      each as $fatura {
        var $pago_em {
          value = $fatura.pago_em|first_notempty:now
        }
      
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.fp_fatura_id == $fatura.id && $db.fp_fatura_item.ref_tipo == "vis_licenca"
          return = {type: "list"}
        } as $itens
      
        foreach ($itens) {
          each as $item {
            conditional {
              if (($item.ref_id|is_empty) == false) {
                db.get vis_licenca {
                  field_name = "id"
                  field_value = $item.ref_id|to_int
                } as $lic
              
                conditional {
                  if (($lic != null) && (($lic.status == "pendente") || ($lic.pago_em == null))) {
                    var $base {
                      value = $lic.valido_ate
                    }
                  
                    conditional {
                      if (($base == null) || ($base < now)) {
                        var.update $base {
                          value = now
                        }
                      }
                    }
                  
                    var $novo_valido {
                      value = $base|add_secs_to_timestamp:30 * 86400
                    }
                  
                    var $novo_status {
                      value = "disponivel"
                    }
                  
                    conditional {
                      if (($lic.status == "disponivel") || ($lic.status == "em_uso")) {
                        var.update $novo_status {
                          value = $lic.status
                        }
                      }
                    }
                  
                    db.patch vis_licenca {
                      field_name = "id"
                      field_value = $lic.id
                      data = {
                        valido_ate: $novo_valido
                        status    : $novo_status
                        pago_em   : $pago_em
                        id_fatura : $fatura.id|to_text
                      }
                    } as $lic_upd
                  
                    var.update $ativadas {
                      value = $ativadas + 1
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