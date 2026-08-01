// Worker — gera fatura de renovacao ConfVision para um franqueado
query fp_fatura_gerar_automatica_confvision verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text worker_key? filters=trim
    text id_franqueado? filters=trim
    int[] vis_licenca_ids?
    text ciclo_ref_prefix? filters=trim
  }

  stack {
    function.run fn_fp_worker_validar {
      input = {worker_key: $input.worker_key}
    } as $worker_check
  
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition ($input.vis_licenca_ids != null && ($input.vis_licenca_ids|count) > 0) {
      error = "vis_licenca_ids obrigatorio"
    }
  
    var $valor_total {
      value = 0
    }
  
    var $qtd_validas {
      value = 0
    }
  
    var $ciclo_primeiro {
      value = ""
    }
  
    foreach ($input.vis_licenca_ids) {
      each as $lid {
        db.get vis_licenca {
          field_name = "id"
          field_value = $lid
        } as $lic
      
        conditional {
          if ($lic != null && $lic.id_franqueado == $input.id_franqueado && ($lic.status == "em_uso" || $lic.status == "disponivel") && $lic.pago_em != null) {
            function.run fn_vis_plano_flags {
              input = {plano: $lic.plano}
            } as $flags
          
            function.run fn_fp_confvision_valor_com_desconto {
              input = {
                id_franqueado: $input.id_franqueado
                valor_base   : $lic.valor|first_notempty:$flags.valor
              }
            } as $preco
          
            var.update $valor_total {
              value = $valor_total + $preco.valor
            }
          
            var.update $qtd_validas {
              value = $qtd_validas + 1
            }
          
            conditional {
              if ($ciclo_primeiro|is_empty) {
                var.update $ciclo_primeiro {
                  value = "VIS-" ~ ($lic.id|to_text) ~ "-" ~ ($lic.valido_ate|format_timestamp:"Ymd":"UTC")
                }
              }
            }
          }
        }
      }
    }
  
    precondition ($qtd_validas > 0) {
      error = "Nenhuma licenca valida para faturar"
    }
  
    var $ciclo {
      value = $input.ciclo_ref_prefix|first_notempty:$ciclo_primeiro
    }
  
    conditional {
      if ($ciclo|is_empty) {
        var.update $ciclo {
          value = "CV-" ~ $input.id_franqueado ~ "-" ~ (now|format_timestamp:"Ymd":"UTC")
        }
      }
    }
  
    var $vencimento {
      value = now|add_secs_to_timestamp:5 * 86400
    }
  
    var $referencia {
      value = "FAT-CV-" ~ (now|format_timestamp:"YmdHis":"UTC") ~ "-" ~ $input.id_franqueado
    }
  
    db.add fp_fatura {
      data = {
        created_at   : "now"
        id_franqueado: $input.id_franqueado
        referencia   : $referencia
        status       : "aberta"
        tipo         : "confvision_renovacao"
        valor_total  : $valor_total
        vencimento_em: $vencimento
        ciclo_ref    : $ciclo
        observacao   : "Renovacao ConfVision automatica"
      }
    } as $fatura
  
    var $itens_criados {
      value = []
    }
  
    foreach ($input.vis_licenca_ids) {
      each as $lid {
        db.get vis_licenca {
          field_name = "id"
          field_value = $lid
        } as $lic
      
        conditional {
          if ($lic != null && $lic.id_franqueado == $input.id_franqueado && ($lic.status == "em_uso" || $lic.status == "disponivel") && $lic.pago_em != null) {
            function.run fn_vis_plano_flags {
              input = {plano: $lic.plano}
            } as $flags
          
            function.run fn_fp_confvision_valor_com_desconto {
              input = {
                id_franqueado: $input.id_franqueado
                valor_base   : $lic.valor|first_notempty:$flags.valor
              }
            } as $preco
          
            var $descricao {
              value = "ConfVision — " ~ $flags.plano_label ~ " (lic " ~ ($lic.id|to_text) ~ ")"
            }
          
            conditional {
              if ($preco.desconto_aplicado) {
                var.update $descricao {
                  value = $descricao ~ " (20% Pro+)"
                }
              }
            }
          
            db.add fp_fatura_item {
              data = {
                created_at    : "now"
                fp_fatura_id  : $fatura.id
                descricao     : $descricao
                quantidade    : 1
                valor_unitario: $preco.valor
                valor_total   : $preco.valor
                ref_tipo      : "vis_licenca"
                ref_id        : $lic.id|to_text
              }
            } as $item
          
            var.update $itens_criados {
              value = $itens_criados|push:$item
            }
          }
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "fatura_confvision_automatica"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_fatura"
        ref_id       : $fatura.id|to_text
        detalhe      : `($itens_criados|count)|to_text ~ " itens"`
        origem       : "worker"
      }
    } as $log
  }

  response = {fatura: $fatura, itens: $itens_criados}
}