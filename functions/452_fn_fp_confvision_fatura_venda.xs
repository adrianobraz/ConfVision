// Venda de licencas ConfVision: cria vis_licenca pendente + fatura aberta
function fn_fp_confvision_fatura_venda {
  input {
    text id_franqueado? filters=trim
    object[] itens? {
      schema {
        text plano? filters=trim
        int quantidade?=1 filters=min:1
      }
    }
  
    text origem?=admin filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.itens != null) && (($input.itens|count) > 0)) {
      error = "itens obrigatorio"
    }
  
    var $valor_total {
      value = 0
    }
  
    var $licencas_criadas {
      value = []
    }
  
    var $itens_fatura {
      value = []
    }
  
    foreach ($input.itens) {
      each as $item {
        precondition (($item.plano|is_empty) == false) {
          error = "plano obrigatorio em cada item"
        }
      
        function.run fn_vis_plano_flags {
          input = {plano: $item.plano}
        } as $flags
      
        precondition ($flags.plano_label != "Nenhum") {
          error = "Plano invalido: " ~ $item.plano
        }
      
        function.run fn_fp_confvision_valor_com_desconto {
          input = {
            id_franqueado: $input.id_franqueado
            valor_base   : $flags.valor
          }
        } as $preco
      
        var $qtd {
          value = $item.quantidade|first_notempty:1
        }
      
        for (50) {
          each as $n {
            conditional {
              if ((($n + 1) > $qtd)) {
                break
              }
            }
          
            db.add vis_licenca {
              data = {
                created_at   : "now"
                id_franqueado: $input.id_franqueado
                plano        : $item.plano
                unidade      : $flags.unidade
                valor        : $preco.valor
                status       : "pendente"
                observacao   : "Aguardando pagamento — " ~ $flags.plano_label
              }
            } as $lic
          
            var.update $licencas_criadas {
              value = $licencas_criadas|push:$lic
            }
          
            var $desc_item {
              value = "ConfVision — " ~ $flags.plano_label
            }
          
            conditional {
              if ($preco.desconto_aplicado) {
                var.update $desc_item {
                  value = $desc_item ~ " (20% Pro+)"
                }
              }
            }
          
            var $ref_lic {
              value = $lic.id|to_text
            }
          
            var $item_fatura {
              value = {
                descricao     : $desc_item
                valor_unitario: $preco.valor
                ref_tipo      : "vis_licenca"
                ref_id        : $ref_lic
              }
            }
          
            var.update $itens_fatura {
              value = $itens_fatura|push:$item_fatura
            }
          
            var.update $valor_total {
              value = $valor_total + $preco.valor
            }
          }
        }
      }
    }
  
    var $vencimento {
      value = now|add_secs_to_timestamp:5 * 86400
    }
  
    var $referencia {
      value = "FAT-CV-" ~ (now|format_timestamp:"YmdHis":"UTC") ~ "-" ~ $input.id_franqueado
    }
  
    var $obs {
      value = "Venda ConfVision — aguardando pagamento"
    }
  
    conditional {
      if ($input.origem == "franqueado") {
        var.update $obs {
          value = "Compra ConfVision pelo portal — aguardando confirmacao da central"
        }
      }
    }
  
    db.add fp_fatura {
      data = {
        created_at   : "now"
        id_franqueado: $input.id_franqueado
        referencia   : $referencia
        status       : "aberta"
        tipo         : "confvision_venda"
        valor_total  : $valor_total
        vencimento_em: $vencimento
        observacao   : $obs
      }
    } as $fatura
  
    var $itens_db {
      value = []
    }
  
    foreach ($itens_fatura) {
      each as $fi {
        db.add fp_fatura_item {
          data = {
            created_at    : "now"
            fp_fatura_id  : $fatura.id
            descricao     : $fi.descricao
            quantidade    : 1
            valor_unitario: $fi.valor_unitario
            valor_total   : $fi.valor_unitario
            ref_tipo      : $fi.ref_tipo
            ref_id        : $fi.ref_id
          }
        } as $item_db
      
        var.update $itens_db {
          value = $itens_db|push:$item_db
        }
      }
    }
  
    var $qtd_lic {
      value = $licencas_criadas|count
    }
  
    var $detalhe_log {
      value = $qtd_lic|to_text
    }
  
    var.update $detalhe_log {
      value = $detalhe_log ~ " licencas"
    }
  
    conditional {
      if ($input.origem == "franqueado") {
        var.update $detalhe_log {
          value = $detalhe_log ~ " (portal ConfVision)"
        }
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "confvision_fatura_venda"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_fatura"
        ref_id       : $fatura.id|to_text
        detalhe      : $detalhe_log
        origem       : $input.origem|first_notempty:"admin"
        valor        : $valor_total
        produto      : "confvision"
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = {
    fatura  : $fatura
    licencas: $licencas_criadas
    itens   : $itens_db
    log     : $log
  }
}