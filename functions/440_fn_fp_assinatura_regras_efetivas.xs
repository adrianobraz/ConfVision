// Resolve modulos/limites efetivos: alacarte usa snapshot + addons pagos (nao pendentes)
function fn_fp_assinatura_regras_efetivas {
  input {
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    text tipo_contratacao? filters=trim
    json modulos_snapshot?
    json limites_snapshot?
    int fp_produto_catalogo_id?
    json addons_json?
    json addons_pendentes_json?
  }

  stack {
    function.run fn_fp_plano_regras_default {
      input = {produto: $input.produto, plano: $input.plano}
    } as $padrao
  
    var $modulos {
      value = $padrao.modulos_json
    }
  
    var $limites {
      value = $padrao.limites_json
    }
  
    var $retencao {
      value = $padrao.retencao_dias
    }
  
    conditional {
      if ($input.tipo_contratacao == "alacarte" && $input.modulos_snapshot != null) {
        var.update $modulos {
          value = $input.modulos_snapshot
        }
      }
    }
  
    conditional {
      if ($input.tipo_contratacao == "alacarte" || ($input.addons_json|count) > 0) {
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
          each as $ch_addon {
            var $esta_pendente {
              value = false
            }
          
            foreach ($pendentes) {
              each as $ch_pend {
                conditional {
                  if ($ch_pend == $ch_addon) {
                    var.update $esta_pendente {
                      value = true
                    }
                  }
                }
              }
            }
          
            conditional {
              if ($esta_pendente == false) {
                var.update $modulos {
                  value = $modulos|set:$ch_addon:true
                }
              }
            }
          }
        }
      }
    }
  
    conditional {
      if ($input.fp_produto_catalogo_id != null && $input.fp_produto_catalogo_id > 0) {
        db.get fp_produto_catalogo {
          field_name = "id"
          field_value = $input.fp_produto_catalogo_id
        } as $cat
      
        conditional {
          if ($cat != null) {
            conditional {
              if ($cat.limites_json != null) {
                var.update $limites {
                  value = $cat.limites_json
                }
              }
            }
          
            conditional {
              if ($cat.retencao_dias != null) {
                var.update $retencao {
                  value = $cat.retencao_dias
                }
              }
            }
          }
        }
      }
    }
  
    conditional {
      if ($input.limites_snapshot != null) {
        var.update $limites {
          value = $input.limites_snapshot
        }
      }
    }
  
    conditional {
      if ($input.modulos_snapshot != null) {
        foreach ($input.modulos_snapshot|keys) {
          each as $ch_snap {
            conditional {
              if (($ch_snap|is_empty) == false) {
                var $ativo_snap {
                  value = $input.modulos_snapshot|get:$ch_snap:false
                }
              
                conditional {
                  if ($ativo_snap) {
                    var $ch_norm {
                      value = $ch_snap
                    }
                  
                    conditional {
                      if ($ch_norm == "franqueadopro") {
                        var.update $ch_norm {
                          value = "franqueadopro.saas"
                        }
                      }
                    }
                  
                    var $incluso_snap {
                      value = $padrao.modulos_json|get:$ch_norm:false
                    }
                  
                    conditional {
                      if ($incluso_snap != true) {
                        var.update $modulos {
                          value = $modulos|set:$ch_norm:true
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
    }
  }

  response = {
    modulos_json : $modulos
    limites_json : $limites
    retencao_dias: $retencao
  }
}