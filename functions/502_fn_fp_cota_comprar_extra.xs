// Compra pacote de cota extra (soma capacidade) — assinatura ativa ou pendente
function fn_fp_cota_comprar_extra {
  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    int fp_pacote_cota_id?
    text origem?=franqueado filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition ($input.fp_pacote_cota_id != null && $input.fp_pacote_cota_id > 0) {
      error = "fp_pacote_cota_id obrigatorio"
    }
  
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $check
  
    precondition ($check.assinatura != null) {
      error = "Contrate uma licenca (Lite/Pro/Pro+) antes de comprar cota"
    }
  
    var $assinatura {
      value = $check.assinatura
    }
  
    db.query fp_fatura {
      where = ($db.fp_fatura.id_franqueado == $input.id_franqueado) && ($db.fp_fatura.status == "aberta")
      return = {type: "list"}
    } as $faturas_abertas
  
    precondition (($faturas_abertas|count) == 0) {
      error = "Existe fatura em aberto. Regularize antes de comprar mais cota."
    }
  
    function.run fn_fp_pacote_cota_valor_efetivo {
      input = {
        fp_pacote_cota_id: $input.fp_pacote_cota_id
        id_franqueado    : $input.id_franqueado
        id_representante : $assinatura.id_representante
        id_central       : $assinatura.id_central
      }
    } as $cota
  
    var $itens {
      value = $assinatura.cotas_json|first_notempty:[]
    }
  
    array.push $itens {
      value = {
        id        : $cota.pacote.id
        nome      : $cota.pacote.nome
        quantidade: $cota.quantidade
        valor     : $cota.valor_venda
      }
    }
  
    // Valor da licenca = valor atual da assinatura menos cotas ja somadas
    function.run fn_fp_cota_recalcular_assinatura {
      input = {
        cotas_json   : $assinatura.cotas_json|first_notempty:[]
        valor_licenca: 0
      }
    } as $antes
  
    var $valor_licenca {
      value = ($assinatura.valor|first_notempty:0) - ($antes.valor_cotas|first_notempty:0)
    }
  
    conditional {
      if ($valor_licenca < 0) {
        var.update $valor_licenca {
          value = 0
        }
      }
    }
  
    function.run fn_fp_cota_recalcular_assinatura {
      input = {cotas_json: $itens, valor_licenca: $valor_licenca}
    } as $calc
  
    db.patch fp_assinatura_produto {
      field_name = "id"
      field_value = $assinatura.id
      data = {
        cotas_json  : $calc.cotas_json
        limites_json: $calc.limites_json
        valor       : $calc.valor_mensal
      }
    } as $assinatura
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "cota_comprar_extra"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_assinatura_produto"
        ref_id       : $assinatura.id|to_text
        detalhe      : "cota " ~ ($cota.pacote.nome|first_notempty:"") ~ " +" ~ ($cota.quantidade|to_text) ~ " valor=" ~ ($cota.valor_venda|to_text)
        origem       : $input.origem|first_notempty:"franqueado"
        valor        : $cota.valor_venda
        produto      : $input.produto
        plano        : $assinatura.plano
        admin_usuario: $input.admin_usuario
      }
    } as $log
  
    function.run fn_fp_fatura_gerar {
      input = {
        assinatura_id  : $assinatura.id
        origem         : $input.origem|first_notempty:"franqueado"
        admin_usuario  : $input.admin_usuario
        acao_log       : "fatura_gerar_cota_extra"
        valor_fatura   : $cota.valor_venda
        descricao_extra: "Pacote de cota: " ~ ($cota.pacote.nome|first_notempty:"")
      }
    } as $fatura_resultado
  }

  response = {
    assinatura         : $assinatura
    cota_comprada      : $cota.pacote
    quantidade_total   : $calc.quantidade_total
    limites_json       : $calc.limites_json
    valor_mensal       : $calc.valor_mensal
    fatura             : $fatura_resultado.fatura
    fatura_criada      : $fatura_resultado.criada
    registro_financeiro: $log
  }
}