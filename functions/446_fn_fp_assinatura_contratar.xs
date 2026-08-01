// Cria ou atualiza assinatura pendente e gera fatura inicial (franqueado ou admin)
// Pacote de cota: input opcional; se vazio usa o vinculado no catalogo
function fn_fp_assinatura_contratar {
  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    int fp_pacote_cota_id?
    text admin_usuario? filters=trim
    text cupom_codigo? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.plano|is_empty) == false) {
      error = "plano obrigatorio"
    }
  
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $check
  
    precondition ($check.liberado == false) {
      error = "Plano ativo. Aguarde o vencimento ou contate a central para alteracao."
    }
  
    // Bundle FP: nao vende se ja incluso no plano FranqueadoPro
    conditional {
      if ($input.produto != "franqueadopro") {
        function.run fn_fp_produto_acesso_efetivo {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : $input.produto
          }
        } as $acesso
      
        conditional {
          if ($acesso.liberado && $acesso.origem == "bundle_fp") {
            var $plano_incluso {
              value = $acesso.plano|first_notempty:""
            }
          
            var $pedido {
              value = $input.plano|first_notempty:""
            }
          
            // Bloqueia se pedir o mesmo tier (ou inferior) ja incluso
            var $bloqueia_bundle {
              value = false
            }
          
            conditional {
              if ($pedido == $plano_incluso) {
                var.update $bloqueia_bundle {
                  value = true
                }
              }
            
              elseif ($input.produto == "webterminal" && $plano_incluso == "pro" && ($pedido == "lite" || $pedido == "pro")) {
                var.update $bloqueia_bundle {
                  value = true
                }
              }
            
              elseif ($input.produto == "webambiente" && $plano_incluso == "pro_plus" && ($pedido == "pro" || $pedido == "pro_plus")) {
                var.update $bloqueia_bundle {
                  value = true
                }
              }
            
              elseif ($input.produto == "webambiente" && $plano_incluso == "pro" && $pedido == "pro") {
                var.update $bloqueia_bundle {
                  value = true
                }
              }
            }
          
            precondition ($bloqueia_bundle == false) {
              error = "Produto ja incluido no seu plano FranqueadoPro (" ~ $plano_incluso ~ "). Nao e necessario comprar novamente."
            }
          }
        }
      }
    }
  
    db.query fp_fatura {
      where = ($db.fp_fatura.id_franqueado == $input.id_franqueado) && ($db.fp_fatura.status == "aberta")
      return = {type: "list"}
    } as $faturas_abertas
  
    precondition (($faturas_abertas|count) == 0) {
      error = "Existe fatura em aberto. Regularize o pagamento antes de contratar outro plano."
    }
  
    var $preco_cen {
      value = ""
    }
  
    conditional {
      if ($check.assinatura != null) {
        var.update $preco_cen {
          value = $check.assinatura.id_central|first_notempty:""|trim
        }
      }
    }
  
    // Primeira contratacao: sem assinatura nao ha id_central — resolve pelo franqueado
    conditional {
      if ($preco_cen|is_empty) {
        function.run fn_fp_franqueado_id_central {
          input = {id_franqueado: $input.id_franqueado}
        } as $cen_fra
      
        var.update $preco_cen {
          value = $cen_fra.id_central|first_notempty:"CENTRAL"|trim
        }
      }
    }
  
    function.run fn_fp_preco_efetivo {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
        plano        : $input.plano
        id_central   : $preco_cen
      }
    } as $preco
  
    var $cat {
      value = $preco.catalogo
    }
  
    precondition ($cat != null) {
      error = "Plano nao encontrado no catalogo"
    }
  
    // FranqueadoPro: licenca + pacote de cota (total mensal)
    var $cotas_itens {
      value = []
    }
  
    var $limites_finais {
      value = $cat.limites_json
    }
  
    var $valor_final {
      value = $preco.valor_venda
    }
  
    var $desc_cota {
      value = ""
    }
  
    conditional {
      if ($input.produto == "franqueadopro") {
        var $cota_id {
          value = $input.fp_pacote_cota_id|first_notempty:0
        }
      
        // Preferencia: cota vinculada no catalogo do plano
        conditional {
          if ($cat.fp_pacote_cota_id != null && $cat.fp_pacote_cota_id > 0) {
            var.update $cota_id {
              value = $cat.fp_pacote_cota_id
            }
          }
        }
      
        precondition ($cota_id != null && $cota_id > 0) {
          error = "Este plano FranqueadoPro nao tem Pacote de Cotas vinculado no catalogo. Peça à Central vincular um pacote."
        }
      
        function.run fn_fp_pacote_cota_valor_efetivo {
          input = {
            fp_pacote_cota_id: $cota_id
            id_franqueado    : $input.id_franqueado
            id_representante : $preco.id_representante
            id_central       : $preco_cen
          }
        } as $cota
      
        array.push $cotas_itens {
          value = {
            id        : $cota_id
            nome      : $cota.pacote.nome
            quantidade: $cota.quantidade
            valor     : $cota.valor_venda
          }
        }
      
        function.run fn_fp_cota_recalcular_assinatura {
          input = {
            cotas_json   : $cotas_itens
            valor_licenca: $preco.valor_venda
          }
        } as $calc_cota
      
        var.update $limites_finais {
          value = $calc_cota.limites_json
        }
      
        var.update $valor_final {
          value = $calc_cota.valor_mensal
        }
      
        var.update $desc_cota {
          value = " + cota " ~ ($cota.pacote.nome|first_notempty:"") ~ " (" ~ ($cota.quantidade|to_text) ~ ")"
        }
      }
    }
  
    function.run fn_fp_assinatura_regras_efetivas {
      input = {
        produto               : $input.produto
        plano                 : $input.plano
        modulos_snapshot      : $cat.modulos_json
        limites_snapshot      : $limites_finais
        fp_produto_catalogo_id: $cat.id
      }
    } as $regras
  
    var $assinatura {
      value = $check.assinatura
    }
  
    conditional {
      if ($assinatura != null) {
        db.patch fp_assinatura_produto {
          field_name = "id"
          field_value = $assinatura.id
          data = {
            plano                  : $input.plano
            status                 : "pendente"
            valor                  : $valor_final
            valor_piso_central     : $preco.valor_piso_central
            valor_piso_breakglass  : $preco.valor_piso_breakglass
            margem_central         : $preco.margem_central
            margem_rep             : $preco.margem_rep
            id_representante       : $preco.id_representante
            id_central             : $preco.id_central
            limites_json           : $limites_finais
            modulos_json           : $regras.modulos_json
            cotas_json             : $cotas_itens
            fp_produto_catalogo_id : $cat.id
            tipo_contratacao       : "pacote"
            valido_ate             : null
            proxima_cobranca_em    : null
            ciclo_fatura_ref       : ""

          }
        } as $assinatura
      }
    
      else {
        db.add fp_assinatura_produto {
          data = {
            created_at             : "now"
            id_franqueado          : $input.id_franqueado
            produto                : $input.produto
            plano                  : $input.plano
            status                 : "pendente"
            periodicidade          : "mensal"
            valor                  : $valor_final
            valor_piso_central     : $preco.valor_piso_central
            valor_piso_breakglass  : $preco.valor_piso_breakglass
            margem_central         : $preco.margem_central
            margem_rep             : $preco.margem_rep
            id_representante       : $preco.id_representante
            id_central             : $preco.id_central
            limites_json           : $limites_finais
            modulos_json           : $regras.modulos_json
            cotas_json             : $cotas_itens
            fp_produto_catalogo_id : $cat.id
            tipo_contratacao       : "pacote"

          }
        } as $assinatura
      }
    }
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "assinatura_contratar"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "fp_assinatura_produto"
        ref_id       : $assinatura.id|to_text
        detalhe      : $input.produto ~ "/" ~ $input.plano ~ $desc_cota ~ " total=" ~ ($valor_final|to_text)
        origem       : $input.origem|first_notempty:"franqueado"
        valor        : $valor_final
        produto      : $input.produto
        plano        : $input.plano
        admin_usuario: $input.admin_usuario
      }
    } as $log
  
    function.run fn_fp_fatura_gerar {
      input = {
        assinatura_id: $assinatura.id
        origem       : $input.origem|first_notempty:"franqueado"
        admin_usuario: $input.admin_usuario
        acao_log     : "fatura_gerar_contratacao"
        valor_fatura : $valor_final
      }
    } as $fatura_resultado
  
    var $cupom_resultado {
      value = null
    }
  
    conditional {
      if ((($input.cupom_codigo|is_empty) == false) && $fatura_resultado.criada && $fatura_resultado.fatura != null) {
        function.run fn_fp_cupom_aplicar_fatura {
          input = {
            codigo       : $input.cupom_codigo
            id_franqueado: $input.id_franqueado
            fatura_id    : $fatura_resultado.fatura.id
            origem       : $input.origem|first_notempty:"franqueado"
            admin_usuario: $input.admin_usuario
          }
        } as $cupom_resultado
      }
    }
  }

  response = {
    assinatura         : $assinatura
    fatura             : $fatura_resultado.fatura
    fatura_criada      : $fatura_resultado.criada
    cupom              : $cupom_resultado
    registro_financeiro: $log
  }
}