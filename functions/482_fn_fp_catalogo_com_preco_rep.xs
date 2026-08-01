// Monta lista do catalogo com piso + valor_venda do REP (ou piso se sem markup)
function fn_fp_catalogo_com_preco_rep {
  input {
    text id_central?=CENTRAL filters=trim
    text id_representante? filters=trim
    text produto? filters=trim
    text ativo?=S filters=trim
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"
    }
  
    db.query fp_produto_catalogo {
      where = $db.fp_produto_catalogo.id_central == $id_central && $db.fp_produto_catalogo.id_representante == "" && $db.fp_produto_catalogo.produto ==? $input.produto && $db.fp_produto_catalogo.ativo ==? $input.ativo
      sort = {
        fp_produto_catalogo.produto     : "asc"
        fp_produto_catalogo.valor_mensal: "asc"
      }
    
      return = {type: "list"}
    } as $catalogo
  
    var $id_rep {
      value = $input.id_representante|first_notempty:""
    }
  
    var $precos {
      value = []
    }
  
    conditional {
      if (($id_rep|is_empty) == false) {
        db.query fp_preco_representante {
          where = $db.fp_preco_representante.id_representante == $id_rep && $db.fp_preco_representante.ativo == "S"
          return = {type: "list"}
        } as $precos
      }
    }
  
    var $saida {
      value = []
    }
  
    foreach ($catalogo) {
      each as $c {
        var $piso {
          value = $c.valor_mensal|first_notempty:0
        }
      
        var $venda {
          value = $piso
        }
      
        var $preco_id {
          value = null
        }
      
        var $tem_markup {
          value = false
        }
      
        foreach ($precos) {
          each as $p {
            conditional {
              if ($p.fp_produto_catalogo_id == $c.id || ($p.produto == $c.produto && $p.plano == $c.plano)) {
                conditional {
                  if ($p.valor_venda != null && $p.valor_venda >= $piso) {
                    var.update $venda {
                      value = $p.valor_venda
                    }
                  
                    var.update $preco_id {
                      value = $p.id
                    }
                  
                    var.update $tem_markup {
                      value = $p.valor_venda > $piso
                    }
                  }
                }
              }
            }
          }
        }
      
        var $item {
          value = $c
            |set:"valor_piso_central":$piso
            |set:"valor_venda":$venda
            |set:"margem_rep":$venda - $piso
            |set:"tem_markup":$tem_markup
            |set:"preco_rep_id":$preco_id
            |set:"id_representante":$id_rep
        }
      
        // Franqueado ve valor_mensal = valor_venda (compat UI)
        var.update $item {
          value = $item|set:"valor_mensal":$venda
        }
      
        // Cota vinculada (franqueadopro): enriquecimento opcional — nunca derruba o catalogo
        var $cota_id_item {
          value = $c.fp_pacote_cota_id|first_notempty:0
        }
      
        conditional {
          if ($cota_id_item > 0) {
            try_catch {
              try {
                db.get fp_pacote_cota {
                  field_name = "id"
                  field_value = $cota_id_item
                } as $pacote_row
              
                conditional {
                  if ($pacote_row != null && $pacote_row.ativo == "S") {
                    var $cota_valor {
                      value = $pacote_row.valor|first_notempty:0
                    }
                  
                    // Markup REP da cota (se houver)
                    conditional {
                      if (($id_rep|is_empty) == false) {
                        db.query fp_preco_pacote_cota {
                          where = $db.fp_preco_pacote_cota.id_representante == $id_rep && $db.fp_preco_pacote_cota.fp_pacote_cota_id == $cota_id_item && $db.fp_preco_pacote_cota.ativo == "S"
                          return = {type: "single"}
                        } as $preco_cota
                      
                        conditional {
                          if ($preco_cota != null && $preco_cota.valor_venda != null && $preco_cota.valor_venda >= $cota_valor) {
                            var.update $cota_valor {
                              value = $preco_cota.valor_venda
                            }
                          }
                        }
                      }
                    }
                  
                    var $qtd_cota {
                      value = $pacote_row.quantidade|first_notempty:0
                    }
                  
                    var $n_cli {
                      value = $qtd_cota
                    }
                  
                    var $n_disp {
                      value = $qtd_cota * 2
                    }
                  
                    var $n_usu {
                      value = $qtd_cota * 4
                    }
                  
                    var $n_set {
                      value = $qtd_cota * 10
                    }
                  
                    var $lim_cota {
                      value = {}
                        |set:"clientes_max":$n_cli
                        |set:"contas_max":$n_disp
                        |set:"usuarios_alarme_max":$n_usu
                        |set:"setores_alarme_max":$n_set
                        |set:"cota_quantidade":$qtd_cota
                    }
                  
                    var.update $item {
                      value = $item
                        |set:"fp_pacote_cota_id":$cota_id_item
                        |set:"pacote_cota_nome":$pacote_row.nome
                        |set:"pacote_cota_quantidade":$qtd_cota
                        |set:"valor_cota":$cota_valor
                        |set:"valor_total_mensal":($venda + $cota_valor)
                        |set:"limites_json":$lim_cota
                    }
                  }
                }
              }
            
              catch {
                // ignora cota invalida
              }
            }
          }
        }
      

        array.push $saida {
          value = $item
        }
      }
    }
  }

  response = {
    dados           : $saida
    total           : $saida|count
    id_representante: $id_rep
    id_central      : $id_central
  }
}
