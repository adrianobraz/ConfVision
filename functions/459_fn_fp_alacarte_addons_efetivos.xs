// Lista chaves de modulos avulsos efetivamente pagos (ativos no snapshot ou addons pagos)
function fn_fp_alacarte_addons_efetivos {
  input {
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    json modulos_snapshot?
    json addons_json?
    json addons_pendentes_json?
  }

  stack {
    function.run fn_fp_plano_regras_default {
      input = {produto: $input.produto, plano: $input.plano}
    } as $padrao
  
    var $lista {
      value = []
    }
  
    var $mods {
      value = $input.modulos_snapshot|first_notempty:{}
    }
  
    foreach ($mods|keys) {
      each as $chave {
        conditional {
          if (($chave|is_empty) == false) {
            var $ativo {
              value = $mods|get:$chave:false
            }
          
            conditional {
              if ($ativo != true) {
                var.update $ativo {
                  value = false
                }
              }
            }
          
            var $incluso_base {
              value = $padrao.modulos_json|get:$chave:false
            }
          
            db.query fp_modulo_catalogo {
              where = $db.fp_modulo_catalogo.id_central == "CENTRAL" && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto == $input.produto && $db.fp_modulo_catalogo.chave == $chave && $db.fp_modulo_catalogo.ativo == "S"
              return = {type: "single"}
            } as $item_cat
          
            conditional {
              if ($ativo && $incluso_base != true && $item_cat != null) {
                var $ja_tem_snap {
                  value = false
                }
              
                foreach ($lista) {
                  each as $ch_ex {
                    conditional {
                      if ($ch_ex == $chave) {
                        var.update $ja_tem_snap {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($ja_tem_snap == false) {
                    var.update $lista {
                      value = $lista|push:$chave
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  
    function.run fn_fp_alacarte_addons_normalizar {
      input = {
        produto: $input.produto
        addons : $input.addons_json|first_notempty:[]
      }
    } as $norm_addons
  
    var $pendentes {
      value = $input.addons_pendentes_json|first_notempty:[]
    }
  
    foreach ($norm_addons.addons|first_notempty:[]) {
      each as $ch_pago {
        var $esta_pendente {
          value = false
        }
      
        foreach ($pendentes) {
          each as $ch_pend {
            conditional {
              if ($ch_pend == $ch_pago) {
                var.update $esta_pendente {
                  value = true
                }
              }
            }
          }
        }
      
        conditional {
          if ($esta_pendente == false) {
            var $incluso_base_pago {
              value = $padrao.modulos_json|get:$ch_pago:false
            }
          
            conditional {
              if ($incluso_base_pago != true) {
                var $ja_tem_pago {
                  value = false
                }
              
                foreach ($lista) {
                  each as $ch_ex2 {
                    conditional {
                      if ($ch_ex2 == $ch_pago) {
                        var.update $ja_tem_pago {
                          value = true
                        }
                      }
                    }
                  }
                }
              
                conditional {
                  if ($ja_tem_pago == false) {
                    var.update $lista {
                      value = $lista|push:$ch_pago
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

  response = {addons_efetivos: $lista}
}