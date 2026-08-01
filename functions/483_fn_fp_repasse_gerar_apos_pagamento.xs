// Split automatico apos pagar fatura do franqueado:
// 1) REP→Central (piso Central)
// 2) Central→Break-glass (piso global, se modo piso)
function fn_fp_repasse_gerar_apos_pagamento {
  input {
    int fatura_id? filters=min:1
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
  
    precondition ($fatura.status == "paga") {
      error = "Fatura precisa estar paga para gerar repasse"
    }
  
    var $gerado {
      value = false
    }
  
    var $gerado_bg {
      value = false
    }
  
    var $motivo {
      value = ""
    }
  
    var $repasse {
      value = null
    }
  
    var $repasse_bg {
      value = null
    }
  
    var $piso_total {
      value = 0
    }
  
    var $piso_bg_total {
      value = 0
    }
  
    var $margem_cen_total {
      value = 0
    }
  
    var $margem_total {
      value = 0
    }
  
    var $id_rep {
      value = $fatura.id_representante|first_notempty:""
    }
  
    var $id_central {
      value = $fatura.id_central|first_notempty:""
    }
  
    conditional {
      if ($fatura.tipo == "repasse_rep_central" || $fatura.tipo == "repasse_central_breakglass") {
        var.update $motivo {
          value = "fatura_ja_e_repasse"
        }
      }
    
      else {
        db.query fp_fatura_item {
          where = $db.fp_fatura_item.fp_fatura_id == $fatura.id
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
                    conditional {
                      if (($id_rep|is_empty) && (($ass.id_representante|is_empty) == false)) {
                        var.update $id_rep {
                          value = $ass.id_representante
                        }
                      }
                    }
                  
                    conditional {
                      if (($id_central|is_empty) && (($ass.id_central|is_empty) == false)) {
                        var.update $id_central {
                          value = $ass.id_central
                        }
                      }
                    }
                  
                    var $piso_item {
                      value = $ass.valor_piso_central|first_notempty:0
                    }
                  
                    var $piso_bg_item {
                      value = $ass.valor_piso_breakglass|first_notempty:0
                    }
                  
                    var $margem_cen_item {
                      value = $ass.margem_central|first_notempty:0
                    }
                  
                    var $margem_item {
                      value = $ass.margem_rep|first_notempty:0
                    }
                  
                    conditional {
                      if ($piso_item <= 0) {
                        var.update $piso_item {
                          value = $ass.valor
                            |first_notempty:($item.valor_total|first_notempty:0)
                        }
                      
                        var.update $margem_item {
                          value = 0
                        }
                      
                        var.update $piso_bg_item {
                          value = 0
                        }
                      
                        var.update $margem_cen_item {
                          value = $piso_item
                        }
                      }
                    }
                  
                    var.update $piso_total {
                      value = $piso_total + $piso_item
                    }
                  
                    var.update $piso_bg_total {
                      value = $piso_bg_total + $piso_bg_item
                    }
                  
                    var.update $margem_cen_total {
                      value = $margem_cen_total + $margem_cen_item
                    }
                  
                    var.update $margem_total {
                      value = $margem_total + $margem_item
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($piso_total <= 0) {
            var.update $piso_total {
              value = $fatura.valor_piso_central
                |first_notempty:($fatura.valor_total|first_notempty:0)
            }
          
            var.update $margem_total {
              value = $fatura.margem_rep|first_notempty:0
            }
          
            var.update $piso_bg_total {
              value = $fatura.valor_piso_breakglass|first_notempty:0
            }
          
            var.update $margem_cen_total {
              value = $fatura.margem_central|first_notempty:($piso_total - $piso_bg_total)
            }
          }
        }
      
        // Modo Livre: NUNCA gera repasse Central→Break-glass (mesmo com piso_bg legado na assinatura)
        conditional {
          if (($id_central|is_empty) == false) {
            function.run fn_fp_central_modo_preco {
              input = {id_central: $id_central}
            } as $modo_cen
          
            conditional {
              if (($modo_cen.modo_preco|first_notempty:"livre") != "piso") {
                var.update $piso_bg_total {
                  value = 0
                }
              
                var.update $margem_cen_total {
                  value = $piso_total
                }
              }
            }
          }
        
          else {
            var.update $piso_bg_total {
              value = 0
            }
          }
        }
      
        conditional {
          if (($id_rep|is_empty) && (($fatura.id_franqueado|is_empty) == false)) {
            function.run fn_fp_franqueado_id_representante {
              input = {id_franqueado: $fatura.id_franqueado}
            } as $rep_res
          
            var.update $id_rep {
              value = $rep_res.id_representante|first_notempty:""
            }
          }
        }
      
        // Nao inventar "CENTRAL" — isso misturava faturas de Centrais distintas no bucket da matriz
        db.patch fp_fatura {
          field_name = "id"
          field_value = $fatura.id
          data = {
            id_representante      : $id_rep
            id_central            : $id_central
            valor_piso_central    : $piso_total
            valor_piso_breakglass : $piso_bg_total
            margem_central        : $margem_cen_total
            margem_rep            : $margem_total
          }
        } as $fatura_upd
      
        // 1) REP → Central
        db.query fp_fatura {
          where = $db.fp_fatura.fatura_origem_id == $fatura.id && $db.fp_fatura.tipo == "repasse_rep_central" && $db.fp_fatura.status != "cancelada"
          return = {type: "single"}
        } as $ja
      
        conditional {
          if ($ja != null) {
            var.update $repasse {
              value = $ja
            }
          
            var.update $motivo {
              value = "repasse_rep_ja_existe"
            }
          }
        
          elseif (($id_rep|is_empty) == false && $piso_total > 0) {
            db.add fp_fatura {
              data = {
                created_at             : "now"
                id_franqueado          : ""
                id_representante       : $id_rep
                id_central             : $id_central
                referencia             : "REPASSE-" ~ ($fatura.id|to_text)
                status                 : "aberta"
                tipo                   : "repasse_rep_central"
                valor_total            : $piso_total
                valor_piso_central     : $piso_total
                valor_piso_breakglass  : $piso_bg_total
                margem_central         : $margem_cen_total
                margem_rep             : $margem_total
                fatura_origem_id       : $fatura.id
                vencimento_em          : now|add_secs_to_timestamp:7 * 86400
                ciclo_ref              : "REP-" ~ ($fatura.id|to_text)
                observacao             : "Repasse automatico REP→Central da fatura #" ~ ($fatura.id|to_text) ~ " (franqueado " ~ ($fatura.id_franqueado|first_notempty:"?") ~ ")"
              }
            } as $repasse
          
            db.add fp_fatura_item {
              data = {
                created_at    : "now"
                fp_fatura_id  : $repasse.id
                descricao     : "Piso Central — origem fatura #" ~ ($fatura.id|to_text)
                quantidade    : 1
                valor_unitario: $piso_total
                valor_total   : $piso_total
                ref_tipo      : "fp_fatura"
                ref_id        : $fatura.id|to_text
              }
            } as $item_rep
          
            function.run fn_fp_financeiro_log {
              input = {
                acao         : "repasse_rep_central"
                id_franqueado: $fatura.id_franqueado
                ref_tipo     : "fp_fatura"
                ref_id       : $repasse.id|to_text
                detalhe      : "origem=" ~ ($fatura.id|to_text) ~ " rep=" ~ $id_rep ~ " piso=" ~ ($piso_total|to_text) ~ " margem_rep=" ~ ($margem_total|to_text)
                origem       : "sistema"
                valor        : $piso_total
                produto      : $fatura.tipo
                admin_usuario: $input.admin_usuario
              }
            } as $log
          
            var.update $gerado {
              value = true
            }
          
            var.update $motivo {
              value = "ok"
            }
          }
        
          else {
            var.update $motivo {
              value = "sem_piso_ou_rep"
            }
          }
        }
      
        // 2) Central → Break-glass (somente se houver piso global)
        db.query fp_fatura {
          where = $db.fp_fatura.fatura_origem_id == $fatura.id && $db.fp_fatura.tipo == "repasse_central_breakglass" && $db.fp_fatura.status != "cancelada"
          return = {type: "single"}
        } as $ja_bg
      
        conditional {
          if ($ja_bg != null) {
            var.update $repasse_bg {
              value = $ja_bg
            }
          }
        
          elseif ($piso_bg_total > 0) {
            db.add fp_fatura {
              data = {
                created_at             : "now"
                id_franqueado          : ""
                id_representante       : $id_rep
                id_central             : $id_central
                referencia             : "REPASSE-BG-" ~ ($fatura.id|to_text)
                status                 : "aberta"
                tipo                   : "repasse_central_breakglass"
                valor_total            : $piso_bg_total
                valor_piso_central     : $piso_total
                valor_piso_breakglass  : $piso_bg_total
                margem_central         : $margem_cen_total
                margem_rep             : $margem_total
                fatura_origem_id       : $fatura.id
                vencimento_em          : now|add_secs_to_timestamp:7 * 86400
                ciclo_ref              : "BG-" ~ ($fatura.id|to_text)
                observacao             : "Repasse automatico Central→Break-glass da fatura #" ~ ($fatura.id|to_text) ~ " (central " ~ $id_central ~ ")"
              }
            } as $repasse_bg
          
            db.add fp_fatura_item {
              data = {
                created_at    : "now"
                fp_fatura_id  : $repasse_bg.id
                descricao     : "Piso Break-glass — origem fatura #" ~ ($fatura.id|to_text)
                quantidade    : 1
                valor_unitario: $piso_bg_total
                valor_total   : $piso_bg_total
                ref_tipo      : "fp_fatura"
                ref_id        : $fatura.id|to_text
              }
            } as $item_bg
          
            function.run fn_fp_financeiro_log {
              input = {
                acao         : "repasse_central_breakglass"
                id_franqueado: $fatura.id_franqueado
                ref_tipo     : "fp_fatura"
                ref_id       : $repasse_bg.id|to_text
                detalhe      : "origem=" ~ ($fatura.id|to_text) ~ " central=" ~ $id_central ~ " piso_bg=" ~ ($piso_bg_total|to_text) ~ " margem_cen=" ~ ($margem_cen_total|to_text)
                origem       : "sistema"
                valor        : $piso_bg_total
                produto      : $fatura.tipo
                admin_usuario: $input.admin_usuario
              }
            } as $log_bg
          
            var.update $gerado_bg {
              value = true
            }
          
            conditional {
              if ($motivo == "" || $motivo == "sem_piso_ou_rep" || $motivo == "repasse_rep_ja_existe") {
                var.update $motivo {
                  value = "ok_bg"
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    gerado                 : $gerado
    gerado_bg              : $gerado_bg
    motivo                 : $motivo
    fatura                 : $repasse
    fatura_breakglass      : $repasse_bg
    valor_piso_central     : $piso_total
    valor_piso_breakglass  : $piso_bg_total
    margem_central         : $margem_cen_total
    margem_rep             : $margem_total
    id_representante       : $id_rep
    id_central             : $id_central
  }
}
