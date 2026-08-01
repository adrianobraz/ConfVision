// Resumo financeiro consolidado (dashboard, DRE, balancete, prestacao, fechamento)
// Escopo: id_representante (carteira REP) ou id_central (Central)
// Caixa/contas a pagar: exclusivos do admin_usuario
function fn_fp_fin_resumo {
  input {
    text competencia? filters=trim
    timestamp? data_inicio?
    timestamp? data_fim?
    text id_representante? filters=trim
    text id_central? filters=trim
    text admin_usuario? filters=trim
    // true soh para break-glass sem Central — nunca para CEN/REP normal
    bool permitir_global?=false
  }

  stack {
    var $id_rep {
      value = $input.id_representante|first_notempty:""|trim
    }
  
    var $id_cen {
      value = $input.id_central|first_notempty:""|trim
    }
  
    var $usuario {
      value = $input.admin_usuario|first_notempty:""|trim
    }
  
    var $escopo_rep {
      value = ($id_rep|is_empty) == false
    }
  
    var $permitir_global {
      value = $input.permitir_global == true
    }
  
    var $ids_fra {
      value = []
    }
  
    conditional {
      if ($escopo_rep) {
        function.run fn_fp_franqueados_ids_representante {
          input = {id_representante: $id_rep}
        } as $carteira
      
        var.update $ids_fra {
          value = $carteira.ids|first_notempty:[]
        }
      }
    }
  
    db.query fp_fatura {
      where = $db.fp_fatura.id > 0
      return = {type: "list"}
    } as $todas_faturas_raw
  
    var $todas_faturas {
      value = $todas_faturas_raw
    }
  
    conditional {
      if ($escopo_rep) {
        var $fats {
          value = []
        }
      
        foreach ($todas_faturas_raw) {
          each as $f {
            function.run fn_fp_fatura_na_carteira {
              input = {
                fatura          : $f
                id_representante: $id_rep
                ids_franqueado  : $ids_fra
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $fats {
                  value = $f
                }
              }
            }
          }
        }
      
        var.update $todas_faturas {
          value = $fats
        }
      }
    
      elseif (($id_cen|is_empty) == false) {
        var $fats_cen {
          value = []
        }
      
        foreach ($todas_faturas_raw) {
          each as $f {
            conditional {
              if (($f.id_central|trim) == $id_cen) {
                array.push $fats_cen {
                  value = $f
                }
              }
            }
          }
        }
      
        var.update $todas_faturas {
          value = $fats_cen
        }
      }
    
      elseif ($permitir_global == false) {
        // Sem escopo: NUNCA devolver dados de todas as Centrais
        var.update $todas_faturas {
          value = []
        }
      }
    }
  
    var $ids_fatura_escopo {
      value = []
    }
  
    foreach ($todas_faturas) {
      each as $fx {
        array.push $ids_fatura_escopo {
          value = $fx.id
        }
      }
    }
  
    db.query fp_pagamento {
      where = $db.fp_pagamento.id > 0
      return = {type: "list"}
    } as $pagamentos_raw
  
    var $pagamentos {
      value = []
    }
  
    // else: sem escopo → pagamentos permanece []
    // else: sem escopo → pagamentos permanece []
    conditional {
      if ($permitir_global == true) {
        var.update $pagamentos {
          value = $pagamentos_raw
        }
      }
    
      elseif ($escopo_rep || (($id_cen|is_empty) == false)) {
        var $pags {
          value = []
        }
      
        foreach ($pagamentos_raw) {
          each as $p {
            var $pag_ok {
              value = false
            }
          
            foreach ($ids_fatura_escopo) {
              each as $fid {
                conditional {
                  if ($pag_ok == false && $p.fp_fatura_id == $fid) {
                    var.update $pag_ok {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($pag_ok) {
                array.push $pags {
                  value = $p
                }
              }
            }
          }
        }
      
        var.update $pagamentos {
          value = $pags
        }
      }
    
      // else: sem escopo → pagamentos permanece []
    }
  
    // Caixa e contas a pagar: exclusivos do usuario logado (CEN ou REP)
    var $retiradas {
      value = []
    }
  
    var $contas_pagar {
      value = []
    }
  
    conditional {
      if (($usuario|is_empty) == false) {
        db.query fp_caixa_movimento {
          where = $db.fp_caixa_movimento.tipo == "retirada" && $db.fp_caixa_movimento.admin_usuario == $usuario
          return = {type: "list"}
        } as $retiradas_user
      
        db.query fp_conta_pagar {
          where = $db.fp_conta_pagar.id > 0 && $db.fp_conta_pagar.admin_usuario == $usuario
          return = {type: "list"}
        } as $contas_pagar_user
      
        var.update $retiradas {
          value = $retiradas_user
        }
      
        var.update $contas_pagar {
          value = $contas_pagar_user
        }
      }
    }
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.status == "ativa"
      return = {type: "list"}
    } as $assinaturas_ativas_raw
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.status == "pendente"
      return = {type: "list"}
    } as $assinaturas_pendentes_raw
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.status == "suspensa"
      return = {type: "list"}
    } as $assinaturas_suspensas_raw
  
    var $assinaturas_ativas {
      value = $assinaturas_ativas_raw
    }
  
    var $assinaturas_pendentes {
      value = $assinaturas_pendentes_raw
    }
  
    var $assinaturas_suspensas {
      value = $assinaturas_suspensas_raw
    }
  
    conditional {
      if ($escopo_rep) {
        var $aa {
          value = []
        }
      
        var $ap {
          value = []
        }
      
        var $as {
          value = []
        }
      
        foreach ($assinaturas_ativas_raw) {
          each as $a {
            function.run fn_fp_assinatura_na_carteira {
              input = {
                assinatura      : $a
                id_representante: $id_rep
                ids_franqueado  : $ids_fra
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $aa {
                  value = $a
                }
              }
            }
          }
        }
      
        foreach ($assinaturas_pendentes_raw) {
          each as $a {
            function.run fn_fp_assinatura_na_carteira {
              input = {
                assinatura      : $a
                id_representante: $id_rep
                ids_franqueado  : $ids_fra
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $ap {
                  value = $a
                }
              }
            }
          }
        }
      
        foreach ($assinaturas_suspensas_raw) {
          each as $a {
            function.run fn_fp_assinatura_na_carteira {
              input = {
                assinatura      : $a
                id_representante: $id_rep
                ids_franqueado  : $ids_fra
              }
            } as $chk
          
            conditional {
              if ($chk.ok) {
                array.push $as {
                  value = $a
                }
              }
            }
          }
        }
      
        var.update $assinaturas_ativas {
          value = $aa
        }
      
        var.update $assinaturas_pendentes {
          value = $ap
        }
      
        var.update $assinaturas_suspensas {
          value = $as
        }
      }
    
      elseif (($id_cen|is_empty) == false) {
        var $aa_cen {
          value = []
        }
      
        var $ap_cen {
          value = []
        }
      
        var $as_cen {
          value = []
        }
      
        foreach ($assinaturas_ativas_raw) {
          each as $a {
            conditional {
              if (($a.id_central|trim) == $id_cen) {
                array.push $aa_cen {
                  value = $a
                }
              }
            }
          }
        }
      
        foreach ($assinaturas_pendentes_raw) {
          each as $a {
            conditional {
              if (($a.id_central|trim) == $id_cen) {
                array.push $ap_cen {
                  value = $a
                }
              }
            }
          }
        }
      
        foreach ($assinaturas_suspensas_raw) {
          each as $a {
            conditional {
              if (($a.id_central|trim) == $id_cen) {
                array.push $as_cen {
                  value = $a
                }
              }
            }
          }
        }
      
        var.update $assinaturas_ativas {
          value = $aa_cen
        }
      
        var.update $assinaturas_pendentes {
          value = $ap_cen
        }
      
        var.update $assinaturas_suspensas {
          value = $as_cen
        }
      }
    
      elseif ($permitir_global == false) {
        var.update $assinaturas_ativas {
          value = []
        }
      
        var.update $assinaturas_pendentes {
          value = []
        }
      
        var.update $assinaturas_suspensas {
          value = []
        }
      }
    }
  
    var $caixa_saldo {
      value = 0
    }
  
    var $caixa_entradas {
      value = 0
    }
  
    var $caixa_saidas {
      value = 0
    }
  
    conditional {
      if (($usuario|is_empty) == false && ($escopo_rep || (($id_cen|is_empty) == false))) {
        function.run fn_fp_caixa_saldo {
          input = {
            id_central      : $id_cen
            id_representante: $id_rep
            admin_usuario   : $usuario
          }
        } as $caixa_escopo
      
        var.update $caixa_saldo {
          value = $caixa_escopo|get:"saldo":0
        }
      
        var.update $caixa_entradas {
          value = $caixa_escopo|get:"total_entradas":0
        }
      
        var.update $caixa_saidas {
          value = $caixa_escopo|get:"total_saidas":0
        }
      }
    }
  
    var $competencia_filtro {
      value = $input.competencia|first_notempty:""
    }
  
    var $tem_competencia {
      value = ($competencia_filtro|strlen) >= 7
    }
  
    var $valor_aberto {
      value = 0
    }
  
    var $valor_pago {
      value = 0
    }
  
    var $valor_vencido {
      value = 0
    }
  
    var $valor_cancelado {
      value = 0
    }
  
    var $valor_repasses_abertos {
      value = 0
    }
  
    var $repasses_periodo {
      value = 0
    }
  
    var $qtd_abertas {
      value = 0
    }
  
    var $qtd_pagas {
      value = 0
    }
  
    var $qtd_vencidas {
      value = 0
    }
  
    var $qtd_canceladas {
      value = 0
    }
  
    var $receita_periodo {
      value = 0
    }
  
    var $entradas_periodo {
      value = 0
    }
  
    var $retiradas_periodo {
      value = 0
    }
  
    var $despesas_periodo {
      value = 0
    }
  
    var $inadimplentes_map {
      value = []
    }
  
    var $ids_fat_repasse {
      value = []
    }
  
    foreach ($todas_faturas) {
      each as $f {
        var $st {
          value = $f.status|first_notempty:""
        }
      
        var $val {
          value = $f.valor_total|first_notempty:0
        }
      
        var $tipo_f {
          value = $f.tipo|first_notempty:""|trim
        }
      
        var $eh_repasse {
          value = $tipo_f == "repasse_rep_central" || $tipo_f == "repasse_central_breakglass" || ($tipo_f|contains:"repasse")
        }
      
        conditional {
          if ($eh_repasse) {
            array.push $ids_fat_repasse {
              value = $f.id
            }
          }
        }
      
        conditional {
          if ($st == "aberta") {
            conditional {
              if ($eh_repasse) {
                var.update $valor_repasses_abertos {
                  value = $valor_repasses_abertos + $val
                }
              }
            
              else {
                var.update $qtd_abertas {
                  value = $qtd_abertas + 1
                }
              
                var.update $valor_aberto {
                  value = $valor_aberto + $val
                }
              
                conditional {
                  if ($f.vencimento_em != null && $f.vencimento_em < now) {
                    var.update $qtd_vencidas {
                      value = $qtd_vencidas + 1
                    }
                  
                    var.update $valor_vencido {
                      value = $valor_vencido + $val
                    }
                  
                    var $ja {
                      value = false
                    }
                  
                    foreach ($inadimplentes_map) {
                      each as $idf {
                        conditional {
                          if ($idf == $f.id_franqueado) {
                            var.update $ja {
                              value = true
                            }
                          }
                        }
                      }
                    }
                  
                    conditional {
                      if ($ja == false && ($f.id_franqueado|strlen) > 0) {
                        var.update $inadimplentes_map {
                          value = $inadimplentes_map|push:$f.id_franqueado
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($st == "paga") {
            var $no_periodo {
              value = true
            }
          
            conditional {
              if ($tem_competencia) {
                var.update $no_periodo {
                  value = false
                }
              
                conditional {
                  if ($f.pago_em != null) {
                    conditional {
                      if (($f.pago_em|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                        var.update $no_periodo {
                          value = true
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($input.data_inicio != null) {
                var.update $no_periodo {
                  value = false
                }
              
                conditional {
                  if ($f.pago_em != null && $f.pago_em >= $input.data_inicio) {
                    conditional {
                      if ($input.data_fim == null || $f.pago_em <= $input.data_fim) {
                        var.update $no_periodo {
                          value = true
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($eh_repasse) {
                conditional {
                  if ($no_periodo) {
                    var.update $repasses_periodo {
                      value = $repasses_periodo + $val
                    }
                  }
                }
              }
            
              else {
                var.update $qtd_pagas {
                  value = $qtd_pagas + 1
                }
              
                var.update $valor_pago {
                  value = $valor_pago + $val
                }
              
                conditional {
                  if ($no_periodo) {
                    var.update $receita_periodo {
                      value = $receita_periodo + $val
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($st == "cancelada") {
            var.update $qtd_canceladas {
              value = $qtd_canceladas + 1
            }
          
            var.update $valor_cancelado {
              value = $valor_cancelado + $val
            }
          }
        }
      }
    }
  
    foreach ($pagamentos) {
      each as $p {
        var $incluir {
          value = true
        }
      
        conditional {
          if ($tem_competencia) {
            var.update $incluir {
              value = false
            }
          
            conditional {
              if ($p.pago_em != null) {
                conditional {
                  if (($p.pago_em|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                    var.update $incluir {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($input.data_inicio != null) {
            var.update $incluir {
              value = false
            }
          
            conditional {
              if ($p.pago_em != null && $p.pago_em >= $input.data_inicio) {
                conditional {
                  if ($input.data_fim == null || $p.pago_em <= $input.data_fim) {
                    var.update $incluir {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($incluir) {
            var $pag_eh_repasse {
              value = false
            }
          
            foreach ($ids_fat_repasse) {
              each as $rid {
                conditional {
                  if ($pag_eh_repasse == false && $p.fp_fatura_id == $rid) {
                    var.update $pag_eh_repasse {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($pag_eh_repasse) {
                // Repasse pago ja contabilizado em repasses_periodo via fatura
              }
            
              else {
                var.update $entradas_periodo {
                  value = $entradas_periodo + ($p.valor|first_notempty:0)
                }
              }
            }
          }
        }
      }
    }
  
    foreach ($retiradas) {
      each as $m {
        var $incluir_r {
          value = true
        }
      
        var $dt {
          value = $m.movimento_em|first_notempty:$m.created_at
        }
      
        conditional {
          if ($tem_competencia) {
            var.update $incluir_r {
              value = false
            }
          
            conditional {
              if ($dt != null) {
                conditional {
                  if (($dt|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                    var.update $incluir_r {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($input.data_inicio != null) {
            var.update $incluir_r {
              value = false
            }
          
            conditional {
              if ($dt != null && $dt >= $input.data_inicio) {
                conditional {
                  if ($input.data_fim == null || $dt <= $input.data_fim) {
                    var.update $incluir_r {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        conditional {
          if ($incluir_r) {
            var.update $retiradas_periodo {
              value = $retiradas_periodo + ($m.valor|first_notempty:0)
            }
          }
        }
      }
    }
  
    var $contas_pagar_abertas {
      value = 0
    }
  
    var $valor_a_pagar {
      value = 0
    }
  
    foreach ($contas_pagar) {
      each as $cp {
        conditional {
          if (($cp.status|first_notempty:"") == "aberta") {
            var.update $contas_pagar_abertas {
              value = $contas_pagar_abertas + 1
            }
          
            var.update $valor_a_pagar {
              value = $valor_a_pagar + ($cp.valor|first_notempty:0)
            }
          }
        }
      
        conditional {
          if (($cp.status|first_notempty:"") == "paga") {
            var $incluir_d {
              value = true
            }
          
            conditional {
              if ($tem_competencia) {
                var.update $incluir_d {
                  value = false
                }
              
                conditional {
                  if ($cp.pago_em != null) {
                    conditional {
                      if (($cp.pago_em|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                        var.update $incluir_d {
                          value = true
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($input.data_inicio != null) {
                var.update $incluir_d {
                  value = false
                }
              
                conditional {
                  if ($cp.pago_em != null && $cp.pago_em >= $input.data_inicio) {
                    conditional {
                      if ($input.data_fim == null || $cp.pago_em <= $input.data_fim) {
                        var.update $incluir_d {
                          value = true
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($incluir_d) {
                var.update $despesas_periodo {
                  value = $despesas_periodo + ($cp.valor|first_notempty:0)
                }
              }
            }
          }
        }
      }
    }
  
    // A pagar = contas a pagar abertas + repasses em aberto
    var.update $valor_a_pagar {
      value = $valor_a_pagar + $valor_repasses_abertos
    }
  
    var $total_despesas_periodo {
      value = $retiradas_periodo + $despesas_periodo + $repasses_periodo
    }
  
    var $resultado_periodo {
      value = $receita_periodo - $total_despesas_periodo
    }
  
    function.run fn_fp_fin_descontos_resumo {
      input = {
        competencia     : $competencia_filtro
        data_inicio     : $input.data_inicio
        data_fim        : $input.data_fim
        id_representante: $id_rep
        id_central      : $id_cen
        permitir_global : $permitir_global
        ids_franqueado  : $ids_fra
      }
    } as $descontos
  
    // MRR / contagem por produto (assinaturas ativas)
    var $produtos_keys {
      value = [
        "franqueadopro"
        "webterminal"
        "terminalmovel"
        "confvision"
        "webambiente"
        "dialyze"
      ]
    }
  
    var $assinaturas_por_produto {
      value = []
    }
  
    var $mrr_total {
      value = 0
    }
  
    var $mrr_avulso {
      value = 0
    }
  
    var $qtd_bundle {
      value = 0
    }
  
    foreach ($produtos_keys) {
      each as $pk {
        var $qtd_p {
          value = 0
        }
      
        var $mrr_p {
          value = 0
        }
      
        var $bundle_p {
          value = 0
        }
      
        foreach ($assinaturas_ativas) {
          each as $aa {
            conditional {
              if (($aa.produto|first_notempty:"") == $pk) {
                var.update $qtd_p {
                  value = $qtd_p + 1
                }
              
                var $val_a {
                  value = $aa.valor|first_notempty:0
                }
              
                var.update $mrr_p {
                  value = $mrr_p + $val_a
                }
              
                conditional {
                  if (($aa.observacao|first_notempty:"")|contains:"bundle_fp") {
                    var.update $bundle_p {
                      value = $bundle_p + 1
                    }
                  
                    var.update $qtd_bundle {
                      value = $qtd_bundle + 1
                    }
                  }
                
                  else {
                    var.update $mrr_avulso {
                      value = $mrr_avulso + $val_a
                    }
                  }
                }
              
                var.update $mrr_total {
                  value = $mrr_total + $val_a
                }
              }
            }
          }
        }
      
        array.push $assinaturas_por_produto {
          value = {
            produto: $pk
            ativas : $qtd_p
            mrr    : $mrr_p
            bundle : $bundle_p
          }
        }
      }
    }
  
    // Receita do periodo por tipo de fatura (aprox. por produto)
    var $receita_franqueadopro {
      value = 0
    }
  
    var $receita_confvision {
      value = 0
    }
  
    var $receita_outros {
      value = 0
    }
  
    foreach ($todas_faturas) {
      each as $fr {
        conditional {
          if (($fr.status|first_notempty:"") == "paga") {
            var $no_per {
              value = true
            }
          
            conditional {
              if ($tem_competencia) {
                var.update $no_per {
                  value = false
                }
              
                conditional {
                  if ($fr.pago_em != null && ($fr.pago_em|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                    var.update $no_per {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($no_per) {
                var $tipo_f {
                  value = $fr.tipo|first_notempty:"assinatura"
                }
              
                var $vf {
                  value = $fr.valor_total|first_notempty:0
                }
              
                conditional {
                  if ($tipo_f == "repasse_rep_central" || $tipo_f == "repasse_central_breakglass" || ($tipo_f|contains:"repasse")) {
                    // repasse nao e receita
                  }
                
                  elseif ($tipo_f == "confvision_venda" || $tipo_f == "confvision_renovacao" || ($tipo_f|contains:"confvision")) {
                    var.update $receita_confvision {
                      value = $receita_confvision + $vf
                    }
                  }
                
                  elseif ($tipo_f == "assinatura") {
                    var.update $receita_franqueadopro {
                      value = $receita_franqueadopro + $vf
                    }
                  }
                
                  else {
                    var.update $receita_outros {
                      value = $receita_outros + $vf
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

  response = {
    assinaturas_ativas     : $assinaturas_ativas|count
    assinaturas_pendentes  : $assinaturas_pendentes|count
    assinaturas_suspensas  : $assinaturas_suspensas|count
    faturas_abertas        : $qtd_abertas
    faturas_pagas          : $qtd_pagas
    faturas_canceladas     : $qtd_canceladas
    faturas_vencidas       : $qtd_vencidas
    valor_em_aberto        : $valor_aberto
    valor_pago             : $valor_pago
    valor_vencido          : $valor_vencido
    valor_cancelado        : $valor_cancelado
    qtd_inadimplentes      : $inadimplentes_map|count
    saldo_caixa            : $caixa_saldo
    total_entradas_caixa   : $caixa_entradas
    total_saidas_caixa     : $caixa_saidas
    receita_periodo        : $receita_periodo
    entradas_periodo       : $entradas_periodo
    retiradas_periodo      : $retiradas_periodo
    despesas_periodo       : $despesas_periodo
    repasses_periodo       : $repasses_periodo
    valor_repasses_abertos : $valor_repasses_abertos
    total_despesas_periodo : $total_despesas_periodo
    resultado_periodo      : $resultado_periodo
    contas_pagar_abertas   : $contas_pagar_abertas
    valor_a_pagar          : $valor_a_pagar
    competencia            : $competencia_filtro
    total_bruto_descontos  : $descontos.total_bruto
    total_descontos        : $descontos.total_descontos
    total_liquido_descontos: $descontos.total_liquido
    total_cupons_desconto  : $descontos.total_cupons
    total_pro_plus_desconto: $descontos.total_pro_plus
    total_credito_abatido  : $descontos.total_credito_abatido
    qtd_descontos          : $descontos.total
    assinaturas_por_produto: $assinaturas_por_produto
    mrr_total              : $mrr_total
    mrr_avulso             : $mrr_avulso
    qtd_assinaturas_bundle : $qtd_bundle
    receita_confvision     : $receita_confvision
    receita_assinaturas    : $receita_franqueadopro
    receita_outros         : $receita_outros
    escopo_rep             : $escopo_rep
    id_representante       : $id_rep

    qtd_franqueados_carteira: $ids_fra|count
  }
}