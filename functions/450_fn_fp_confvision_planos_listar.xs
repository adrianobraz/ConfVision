// Lista catalogo canonico de planos ConfVision (18 slugs) com flags e precos
function fn_fp_confvision_planos_listar {
  input {
    text id_franqueado? filters=trim
  }

  stack {
    var $slugs {
      value = [
        "online"
        "sensor_foto"
        "sensor_foto_video"
        "analitico_armado_evento"
        "analitico_armado_foto"
        "analitico_armado_foto_video"
        "analitico_24h_evento"
        "analitico_24h_foto"
        "analitico_24h_foto_video"
        "gravacao_7d"
        "gravacao_15d"
        "gravacao_30d"
        "gravacao_movimento_7d"
        "gravacao_movimento_15d"
        "gravacao_movimento_30d"
        "gravacao_timelapse_7d"
        "gravacao_timelapse_15d"
        "gravacao_timelapse_30d"
      ]
    }
  
    var $planos {
      value = []
    }
  
    var $desconto_pro_plus {
      value = false
    }
  
    conditional {
      if (($input.id_franqueado|is_empty) == false) {
        function.run fn_fp_assinatura_get_ativa {
          input = {
            id_franqueado: $input.id_franqueado
            produto      : "franqueadopro"
          }
        } as $fp
      
        conditional {
          if ($fp.liberado && $fp.assinatura != null && $fp.assinatura.plano == "pro_plus") {
            var.update $desconto_pro_plus {
              value = true
            }
          }
        }
      }
    }
  
    foreach ($slugs) {
      each as $slug {
        function.run fn_vis_plano_flags {
          input = {plano: $slug}
        } as $flags
      
        var $valor_final {
          value = $flags.valor
        }
      
        var $desconto_pct {
          value = 0
        }
      
        conditional {
          if ($desconto_pro_plus) {
            var.update $valor_final {
              value = ($flags.valor * 0.8)|round:2
            }
          
            var.update $desconto_pct {
              value = 20
            }
          }
        }
      
        var.update $planos {
          value = $planos
            |push:```
              {
                plano             : $slug
                plano_label       : $flags.plano_label
                unidade           : $flags.unidade
                valor             : $flags.valor
                valor_com_desconto: $valor_final
                desconto_pct      : $desconto_pct
              }
              ```
        }
      }
    }
  }

  response = {
    planos           : $planos
    desconto_pro_plus: $desconto_pro_plus
    total            : $planos|count
  }
}