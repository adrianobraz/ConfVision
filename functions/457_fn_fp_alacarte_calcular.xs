// Calcula valor e modulos efetivos para contratacao a la carte
function fn_fp_alacarte_calcular {
  input {
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    text id_central? filters=trim
    json addons?
  }

  stack {
    precondition (($input.plano|is_empty) == false) {
      error = "plano base obrigatorio"
    }
  
    var $id_cen {
      value = $input.id_central|first_notempty:""|trim
    }
  
    conditional {
      if ($id_cen|is_empty) {
        var.update $id_cen {
          value = "CENTRAL"
        }
      }
    }
  
    function.run fn_fp_catalogo_get_por_plano {
      input = {
        produto   : $input.produto
        plano     : $input.plano
        id_central: $id_cen
      }
    } as $cat
  
    conditional {
      if ($cat == null && $id_cen != "CENTRAL") {
        function.run fn_fp_catalogo_get_por_plano {
          input = {
            produto   : $input.produto
            plano     : $input.plano
            id_central: "CENTRAL"
          }
        } as $cat
      }
    }
  
    precondition ($cat != null) {
      error = "Plano base nao encontrado no catalogo"
    }
  
    function.run fn_fp_plano_regras_default {
      input = {produto: $input.produto, plano: $input.plano}
    } as $padrao
  
    var $modulos {
      value = $padrao.modulos_json
    }
  
    var $valor_addons {
      value = 0
    }
  
    var $addons_cobrados {
      value = []
    }
  
    var $itens_detalhe {
      value = []
    }
  
    var $lista_addons {
      value = $input.addons|first_notempty:[]
    }
  
    foreach ($lista_addons) {
      each as $chave {
        conditional {
          if (($chave|is_empty) == false) {
            var $ja_incluso {
              value = $padrao.modulos_json|get:$chave:false
            }
          
            var.update $modulos {
              value = $modulos|set:$chave:true
            }
          
            conditional {
              if ($ja_incluso != true) {
                db.query fp_modulo_catalogo {
                  where = $db.fp_modulo_catalogo.id_central == "CENTRAL" && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.chave == $chave && $db.fp_modulo_catalogo.ativo == "S"
                  return = {type: "single"}
                } as $item_cat
              
                conditional {
                  if ($item_cat != null) {
                    var $preco {
                      value = $item_cat.valor_mensal|first_notempty:0
                    }
                  
                    var.update $valor_addons {
                      value = $valor_addons + $preco
                    }
                  
                    var.update $addons_cobrados {
                      value = $addons_cobrados|push:$chave
                    }
                  
                    var.update $itens_detalhe {
                      value = $itens_detalhe
                        |push:```
                          {
                            chave       : $chave
                            label       : $item_cat.label
                            valor_mensal: $preco
                            cobrado     : true
                          }
                          ```
                    }
                  }
                }
              }
            
              else {
                db.query fp_modulo_catalogo {
                  where = $db.fp_modulo_catalogo.id_central == "CENTRAL" && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.chave == $chave
                  return = {type: "single"}
                } as $item_inc
              
                var.update $itens_detalhe {
                  value = $itens_detalhe
                    |push:```
                      {
                        chave       : $chave
                        label       : $item_inc.label|first_notempty:$chave
                        valor_mensal: 0
                        cobrado     : false
                        incluso_base: true
                      }
                      ```
                }
              }
            }
          }
        }
      }
    }
  
    var $valor_total {
      value = ($cat.valor_mensal + $valor_addons)|round:2
    }
  }

  response = {
    produto               : $input.produto
    plano_base            : $input.plano
    valor_base            : $cat.valor_mensal
    valor_addons          : $valor_addons
    valor_total           : $valor_total
    modulos_json          : $modulos
    limites_json          : $padrao.limites_json
    retencao_dias         : $padrao.retencao_dias
    fp_produto_catalogo_id: $cat.id
    addons_cobrados       : $addons_cobrados
    itens_detalhe         : $itens_detalhe
  }
}