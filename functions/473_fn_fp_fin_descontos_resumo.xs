// Consolida descontos do periodo: cupons usados + abatimentos de credito + Pro+ ConfVision
// Escopo: REP (carteira) ou CEN (id_central). Sem escopo e sem permitir_global → totais zerados.
function fn_fp_fin_descontos_resumo {
  input {
    text competencia? filters=trim
    timestamp? data_inicio?
    timestamp? data_fim?
    text id_franqueado? filters=trim
    text produto? filters=trim
    text tipo_desconto? filters=trim
    text codigo? filters=trim
    text id_representante? filters=trim
    text id_central? filters=trim
    bool permitir_global?=false

    json ids_franqueado?
  }

  stack {
    var $competencia_filtro {
      value = $input.competencia|first_notempty:""
    }
  
    var $tem_competencia {
      value = ($competencia_filtro|strlen) >= 7
    }
  
    var $tipo_filtro {
      value = $input.tipo_desconto|first_notempty:"todos"|to_lower
    }
  
    var $codigo_filtro {
      value = $input.codigo
        |first_notempty:""
        |trim
        |to_upper
    }
  
    var $produto_filtro {
      value = $input.produto
        |first_notempty:""
        |trim
        |to_lower
    }
  
    var $franq_filtro {
      value = $input.id_franqueado|first_notempty:""|trim
    }
  
    var $ids_carteira {
      value = $input.ids_franqueado|first_notempty:[]
    }
  
    var $id_cen {
      value = $input.id_central|first_notempty:""|trim
    }
  
    var $escopo_rep {
      value = (($input.id_representante|first_notempty:"")|is_empty) == false
    }
  
    var $escopo_cen {
      value = ($id_cen|is_empty) == false
    }
  
    var $permitir_global {
      value = $input.permitir_global == true
    }
  
    // Sem REP, sem Central e sem break-glass global: nao consolida nada
    var $sem_escopo {
      value = $escopo_rep == false && $escopo_cen == false && $permitir_global == false
    }
  
    db.query fp_cupom_desconto {
      where = $db.fp_cupom_desconto.status == "usado"
      sort = {fp_cupom_desconto.usado_em: "desc"}
      return = {type: "list"}
    } as $cupons
  
    db.query fp_fatura_item {
      where = $db.fp_fatura_item.id > 0
      return = {type: "list"}
    } as $itens
  
    db.query fp_fatura {
      where = $db.fp_fatura.id > 0
      return = {type: "list"}
    } as $faturas
  
    db.query fp_financeiro_log {
      where = $db.fp_financeiro_log.acao == "cupom_credito_abatido"
      sort = {fp_financeiro_log.created_at: "desc"}
      return = {type: "list"}
    } as $logs_credito
  
    var $dados {
      value = []
    }
  
    var $total_bruto {
      value = 0
    }
  
    var $total_descontos {
      value = 0
    }
  
    var $total_liquido {
      value = 0
    }
  
    var $total_cupons {
      value = 0
    }
  
    var $total_pro_plus {
      value = 0
    }
  
    var $total_credito_abatido {
      value = 0
    }
  
    var $qtd_cupons {
      value = 0
    }
  
    var $qtd_pro_plus {
      value = 0
    }
  
    var $qtd_credito {
      value = 0
    }
  
    var $franqueados_map {
      value = []
    }
  
    // --- Cupons usados (valor_base / desconto / final gravados no cupom) ---
    foreach ($cupons) {
      each as $c {
        var $ok {
          value = true
        }
      
        var $dt {
          value = $c.usado_em|first_notempty:$c.created_at
        }
      
        conditional {
          if ($tem_competencia) {
            var.update $ok {
              value = false
            }
          
            conditional {
              if ($dt != null && ($dt|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                var.update $ok {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($input.data_inicio != null) {
            var.update $ok {
              value = false
            }
          
            conditional {
              if ($dt != null && $dt >= $input.data_inicio) {
                conditional {
                  if ($input.data_fim == null || $dt <= $input.data_fim) {
                    var.update $ok {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        var $idf {
          value = $c.usado_por_id_franqueado
            |first_notempty:($c.id_franqueado|first_notempty:"")
        }
      
        conditional {
          if (($franq_filtro|strlen) > 0 && $idf != $franq_filtro) {
            var.update $ok {
              value = false
            }
          }
        }
      
        conditional {
          if ($sem_escopo) {
            var.update $ok {
              value = false
            }
          }
        }
      
        conditional {
          if ($escopo_cen && $permitir_global == false && $ok) {
            conditional {
              if ((($c.id_central|trim)|first_notempty:"") != $id_cen) {
                var.update $ok {
                  value = false
                }
              }
            }
          }
        }
      
        conditional {
          if ($escopo_rep && $ok) {
            var $na_cart {
              value = false
            }
          
            conditional {
              if ($c.id_representante == $input.id_representante) {
                var.update $na_cart {
                  value = true
                }
              }
            }
          
            foreach ($ids_carteira) {
              each as $idc {
                conditional {
                  if ($idf == $idc) {
                    var.update $na_cart {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($na_cart == false) {
                var.update $ok {
                  value = false
                }
              }
            }
          }
        }
      
        var $prod {
          value = $c.produto
            |first_notempty:"franqueadopro"
            |to_lower
        }
      
        conditional {
          if (($produto_filtro|strlen) > 0 && $prod != $produto_filtro && $prod != "todos") {
            var.update $ok {
              value = false
            }
          }
        }
      
        conditional {
          if (($codigo_filtro|strlen) > 0 && ($c.codigo|first_notempty:""|to_upper) != $codigo_filtro) {
            var.update $ok {
              value = false
            }
          }
        }
      
        conditional {
          if ($tipo_filtro != "todos" && $tipo_filtro != "cupom") {
            var.update $ok {
              value = false
            }
          }
        }
      
        conditional {
          if ($ok) {
            var $vb {
              value = $c.valor_base|first_notempty:0
            }
          
            var $vd {
              value = $c.valor_desconto|first_notempty:0
            }
          
            var $vf {
              value = $c.valor_final|first_notempty:0
            }
          
            var $modo {
              value = "cupom"
            }
          
            conditional {
              if (($c.ref_tipo|first_notempty:"") == "fp_fatura") {
                var.update $modo {
                  value = "fatura"
                }
              }
            }
          
            array.push $dados {
              value = {
                data          : $dt
                tipo_desconto : "cupom"
                modo          : $modo
                codigo_cupom  : $c.codigo
                tipo_cupom    : $c.tipo
                id_franqueado : $idf
                produto       : $prod
                plano         : ""
                valor_base    : $vb
                valor_desconto: $vd
                valor_final   : $vf
                fatura_id     : $c.ref_id
                fatura_status : ""
                cupom_id      : $c.id
                ref_tipo      : $c.ref_tipo
                ref_id        : $c.ref_id
                observacao    : $c.observacao
              }
            }
          
            var.update $total_bruto {
              value = $total_bruto + $vb
            }
          
            var.update $total_descontos {
              value = $total_descontos + $vd
            }
          
            var.update $total_liquido {
              value = $total_liquido + $vf
            }
          
            var.update $total_cupons {
              value = $total_cupons + $vd
            }
          
            var.update $qtd_cupons {
              value = $qtd_cupons + 1
            }
          
            var $ja {
              value = false
            }
          
            foreach ($franqueados_map) {
              each as $fmap {
                conditional {
                  if ($fmap == $idf) {
                    var.update $ja {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($ja == false && ($idf|strlen) > 0) {
                var.update $franqueados_map {
                  value = $franqueados_map|push:$idf
                }
              }
            }
          }
        }
      }
    }
  
    // --- Desconto automatico Pro+ ConfVision (itens com "20% Pro+") ---
    foreach ($itens) {
      each as $it {
        var $desc {
          value = $it.descricao|first_notempty:""
        }
      
        conditional {
          if ($desc|contains:"20% Pro+") {
            var $fatura_pp {
              value = null
            }
          
            foreach ($faturas) {
              each as $f {
                conditional {
                  if ($f.id == $it.fp_fatura_id) {
                    var.update $fatura_pp {
                      value = $f
                    }
                  }
                }
              }
            }
          
            var $idf_pp {
              value = ""
            }
          
            var $dt_pp {
              value = $it.created_at
            }
          
            var $fat_status {
              value = ""
            }
          
            conditional {
              if ($fatura_pp != null) {
                var.update $idf_pp {
                  value = $fatura_pp.id_franqueado|first_notempty:""
                }
              
                var.update $fat_status {
                  value = $fatura_pp.status|first_notempty:""
                }
              
                conditional {
                  if ($fatura_pp.created_at != null) {
                    var.update $dt_pp {
                      value = $fatura_pp.created_at
                    }
                  }
                }
              }
            }
          
            var $ok_pp {
              value = true
            }
          
            conditional {
              if ($tem_competencia) {
                var.update $ok_pp {
                  value = false
                }
              
                conditional {
                  if ($dt_pp != null && ($dt_pp|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                    var.update $ok_pp {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($input.data_inicio != null) {
                var.update $ok_pp {
                  value = false
                }
              
                conditional {
                  if ($dt_pp != null && $dt_pp >= $input.data_inicio) {
                    conditional {
                      if ($input.data_fim == null || $dt_pp <= $input.data_fim) {
                        var.update $ok_pp {
                          value = true
                        }
                      }
                    }
                  }
                }
              }
            }
          
            conditional {
              if (($franq_filtro|strlen) > 0 && $idf_pp != $franq_filtro) {
                var.update $ok_pp {
                  value = false
                }
              }
            }
          
            conditional {
              if ($sem_escopo) {
                var.update $ok_pp {
                  value = false
                }
              }
            }
          
            conditional {
              if ($escopo_cen && $permitir_global == false && $ok_pp) {
                var $cen_pp {
                  value = ""
                }
              
                conditional {
                  if ($fatura_pp != null) {
                    var.update $cen_pp {
                      value = $fatura_pp.id_central|trim
                    }
                  }
                }
              
                conditional {
                  if ($cen_pp != $id_cen) {
                    var.update $ok_pp {
                      value = false
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($escopo_rep && $ok_pp) {
                var $na_cart_pp {
                  value = false
                }
              
                foreach ($ids_carteira) {
                  each as $idc {
                    conditional {
                      if ($idf_pp == $idc) {
                        var.update $na_cart_pp {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($na_cart_pp == false) {
                    var.update $ok_pp {
                      value = false
                    }
                  }
                }
              }
            }
          
            conditional {
              if (($produto_filtro|strlen) > 0 && $produto_filtro != "confvision") {
                var.update $ok_pp {
                  value = false
                }
              }
            }
          
            conditional {
              if (($codigo_filtro|strlen) > 0) {
                var.update $ok_pp {
                  value = false
                }
              }
            }
          
            conditional {
              if ($tipo_filtro != "todos" && $tipo_filtro != "pro_plus") {
                var.update $ok_pp {
                  value = false
                }
              }
            }
          
            conditional {
              if ($ok_pp) {
                var $vf_pp {
                  value = $it.valor_total
                    |first_notempty:($it.valor_unitario|first_notempty:0)
                }
              
                // Preco cobrado ja e 80% do valor de tabela
                var $vb_pp {
                  value = (($vf_pp / 0.8)|round:2)
                }
              
                var $vd_pp {
                  value = ($vb_pp - $vf_pp)|round:2
                }
              
                array.push $dados {
                  value = ```
                    {
                      data          : $dt_pp
                      tipo_desconto : "pro_plus"
                      modo          : "automatico"
                      codigo_cupom  : ""
                      tipo_cupom    : "percentual"
                      id_franqueado : $idf_pp
                      produto       : "confvision"
                      plano         : ""
                      valor_base    : $vb_pp
                      valor_desconto: $vd_pp
                      valor_final   : $vf_pp
                      fatura_id     : ($it.fp_fatura_id|to_text)
                      fatura_status : $fat_status
                      cupom_id      : 0
                      ref_tipo      : $it.ref_tipo
                      ref_id        : $it.ref_id
                      observacao    : $desc
                    }
                    ```
                }
              
                var.update $total_bruto {
                  value = $total_bruto + $vb_pp
                }
              
                var.update $total_descontos {
                  value = $total_descontos + $vd_pp
                }
              
                var.update $total_liquido {
                  value = $total_liquido + $vf_pp
                }
              
                var.update $total_pro_plus {
                  value = $total_pro_plus + $vd_pp
                }
              
                var.update $qtd_pro_plus {
                  value = $qtd_pro_plus + 1
                }
              
                var $ja_pp {
                  value = false
                }
              
                foreach ($franqueados_map) {
                  each as $fmap2 {
                    conditional {
                      if ($fmap2 == $idf_pp) {
                        var.update $ja_pp {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($ja_pp == false && ($idf_pp|strlen) > 0) {
                    var.update $franqueados_map {
                      value = $franqueados_map|push:$idf_pp
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  
    // --- Credito de cupom abatido em fatura (visibilidade; nao soma em total_descontos para evitar double-count) ---
    foreach ($logs_credito) {
      each as $lc {
        var $ok_cr {
          value = true
        }
      
        var $dt_cr {
          value = $lc.created_at
        }
      
        conditional {
          if ($tem_competencia) {
            var.update $ok_cr {
              value = false
            }
          
            conditional {
              if ($dt_cr != null && ($dt_cr|format_timestamp:"Y-m":"UTC") == $competencia_filtro) {
                var.update $ok_cr {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($input.data_inicio != null) {
            var.update $ok_cr {
              value = false
            }
          
            conditional {
              if ($dt_cr != null && $dt_cr >= $input.data_inicio) {
                conditional {
                  if ($input.data_fim == null || $dt_cr <= $input.data_fim) {
                    var.update $ok_cr {
                      value = true
                    }
                  }
                }
              }
            }
          }
        }
      
        var $idf_cr {
          value = $lc.id_franqueado|first_notempty:""
        }
      
        conditional {
          if (($franq_filtro|strlen) > 0 && $idf_cr != $franq_filtro) {
            var.update $ok_cr {
              value = false
            }
          }
        }
      
        conditional {
          if ($sem_escopo) {
            var.update $ok_cr {
              value = false
            }
          }
        }
      
        conditional {
          if ($escopo_cen && $permitir_global == false && $ok_cr) {
            conditional {
              if ((($lc.id_central|trim)|first_notempty:"") != $id_cen) {
                var.update $ok_cr {
                  value = false
                }
              }
            }
          }
        }
      
        conditional {
          if ($escopo_rep && $ok_cr) {
            var $na_cart_cr {
              value = false
            }
          
            foreach ($ids_carteira) {
              each as $idc {
                conditional {
                  if ($idf_cr == $idc) {
                    var.update $na_cart_cr {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($na_cart_cr == false) {
                var.update $ok_cr {
                  value = false
                }
              }
            }
          }
        }
      
        var $prod_cr {
          value = $lc.produto
            |first_notempty:"franqueadopro"
            |to_lower
        }
      
        conditional {
          if (($produto_filtro|strlen) > 0 && $prod_cr != $produto_filtro) {
            var.update $ok_cr {
              value = false
            }
          }
        }
      
        conditional {
          if (($codigo_filtro|strlen) > 0) {
            var.update $ok_cr {
              value = false
            }
          }
        }
      
        conditional {
          if ($tipo_filtro != "todos" && $tipo_filtro != "credito") {
            var.update $ok_cr {
              value = false
            }
          }
        }
      
        conditional {
          if ($ok_cr) {
            var $vd_cr {
              value = $lc.valor|first_notempty:0
            }
          
            array.push $dados {
              value = ```
                {
                  data          : $dt_cr
                  tipo_desconto : "credito"
                  modo          : "credito_abatido"
                  codigo_cupom  : ""
                  tipo_cupom    : ""
                  id_franqueado : $idf_cr
                  produto       : $prod_cr
                  plano         : $lc.plano|first_notempty:""
                  valor_base    : 0
                  valor_desconto: $vd_cr
                  valor_final   : 0
                  fatura_id     : $lc.ref_id
                  fatura_status : ""
                  cupom_id      : 0
                  ref_tipo      : $lc.ref_tipo
                  ref_id        : $lc.ref_id
                  observacao    : $lc.detalhe
                }
                ```
            }
          
            var.update $total_credito_abatido {
              value = $total_credito_abatido + $vd_cr
            }
          
            var.update $qtd_credito {
              value = $qtd_credito + 1
            }
          
            // Credito so entra no bruto/liquido quando filtro e so credito (senao ja contou no cupom)
            conditional {
              if ($tipo_filtro == "credito") {
                var.update $total_descontos {
                  value = $total_descontos + $vd_cr
                }
              }
            }
          
            var $ja_cr {
              value = false
            }
          
            foreach ($franqueados_map) {
              each as $fmap3 {
                conditional {
                  if ($fmap3 == $idf_cr) {
                    var.update $ja_cr {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($ja_cr == false && ($idf_cr|strlen) > 0) {
                var.update $franqueados_map {
                  value = $franqueados_map|push:$idf_cr
                }
              }
            }
          }
        }
      }
    }
  }

  response = {
    dados                       : $dados
    total                       : $dados|count
    total_bruto                 : $total_bruto
    total_descontos             : $total_descontos
    total_liquido               : $total_liquido
    total_cupons                : $total_cupons
    total_pro_plus              : $total_pro_plus
    total_credito_abatido       : $total_credito_abatido
    qtd_cupons                  : $qtd_cupons
    qtd_pro_plus                : $qtd_pro_plus
    qtd_credito                 : $qtd_credito
    qtd_franqueados_beneficiados: $franqueados_map|count
    competencia                 : $competencia_filtro
  }
}