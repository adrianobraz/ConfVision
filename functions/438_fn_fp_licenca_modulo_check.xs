// Verifica modulo/chave_menu (reutilizavel)
function fn_fp_licenca_modulo_check {
  input {
    text id_franqueado? filters=trim
    text chave_menu? filters=trim
    text produto?=franqueadopro filters=trim
  }

  stack {
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $check
  
    var $liberado {
      value = $check.liberado
    }
  
    var $motivo {
      value = $check.motivo
    }
  
    var $plano {
      value = null
    }
  
    conditional {
      if ($liberado && $check.assinatura != null && ($input.chave_menu|is_empty) == false) {
        var.update $plano {
          value = $check.assinatura.plano
        }
      
        function.run fn_fp_assinatura_regras_efetivas {
          input = {
            produto               : $input.produto
            plano                 : $check.assinatura.plano
            tipo_contratacao      : $check.assinatura.tipo_contratacao
            modulos_snapshot      : $check.assinatura.modulos_json
            limites_snapshot      : $check.assinatura.limites_json
            fp_produto_catalogo_id: $check.assinatura.fp_produto_catalogo_id
          }
        } as $regras
      
        var $flag {
          value = $regras.modulos_json|get:$input.chave_menu:null
        }
      
        conditional {
          if ($flag != true) {
            var.update $liberado {
              value = false
            }
          
            var.update $motivo {
              value = "modulo_bloqueado_plano"
            }
          }
        }
      }
    }
  }

  response = {
    liberado  : $liberado
    motivo    : $motivo
    chave_menu: $input.chave_menu
    plano     : $plano
  }
}