// Acesso efetivo ao produto: assinatura propria ATIVA ou incluso no bundle FranqueadoPro
function fn_fp_produto_acesso_efetivo {
  input {
    text id_franqueado? filters=trim
    text produto? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.produto|is_empty) == false) {
      error = "produto obrigatorio"
    }
  
    var $produto {
      value = $input.produto|to_lower
    }
  
    function.run fn_fp_assinatura_get_ativa {
      input = {id_franqueado: $input.id_franqueado, produto: $produto}
    } as $check
  
    var $liberado {
      value = $check.liberado
    }
  
    var $motivo {
      value = $check.motivo
    }
  
    var $origem {
      value = "nenhum"
    }
  
    var $plano {
      value = ""
    }
  
    var $assinatura {
      value = $check.assinatura
    }
  
    conditional {
      if ($liberado) {
        var.update $origem {
          value = "assinatura"
        }
      
        conditional {
          if ($assinatura != null) {
            var.update $plano {
              value = $assinatura.plano|first_notempty:""
            }
          
            conditional {
              if (($assinatura.observacao|first_notempty:"")|contains:"bundle_fp") {
                var.update $origem {
                  value = "bundle_fp"
                }
              }
            }
          }
        }
      }
    }
  
    // Bundle FP: produto incluido no tier ativo do FranqueadoPro
    conditional {
      if ($liberado == false && $produto != "franqueadopro") {
        function.run fn_fp_assinatura_get_ativa {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "franqueadopro"
          }
        } as $fp
      
        conditional {
          if ($fp.liberado && $fp.assinatura != null) {
            function.run fn_fp_bundle_mapa {
              input = {plano_fp: $fp.assinatura.plano}
            } as $mapa
          
            foreach ($mapa.itens) {
              each as $it {
                conditional {
                  if ($liberado == false && $it.produto == $produto) {
                    var.update $liberado {
                      value = true
                    }
                  
                    var.update $motivo {
                      value = "incluso_franqueadopro"
                    }
                  
                    var.update $origem {
                      value = "bundle_fp"
                    }
                  
                    var.update $plano {
                      value = $it.plano
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
    liberado  : $liberado
    motivo    : $motivo
    origem    : $origem
    produto   : $produto
    plano     : $plano
    assinatura: $assinatura
  }
}