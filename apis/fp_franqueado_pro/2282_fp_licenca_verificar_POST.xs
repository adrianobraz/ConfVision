// Verifica licenca do produto para o franqueado (runtime)
query fp_licenca_verificar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    // Usa acesso efetivo (assinatura propria ou bundle FP) para qualquer produto
    function.run fn_fp_produto_acesso_efetivo {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $acesso_efetivo
  
    function.run fn_fp_assinatura_get_ativa {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $check
  
    var $liberado {
      value = $acesso_efetivo.liberado
    }
  
    var $motivo {
      value = $acesso_efetivo.motivo
    }
  
    var $regras {
      value = null
    }
  
    var $addons_efetivos {
      value = []
    }
  
    var $plano_efetivo {
      value = $acesso_efetivo.plano|first_notempty:""
    }
  
    conditional {
      if ($check.assinatura != null && $check.liberado) {
        function.run fn_fp_assinatura_regras_efetivas {
          input = {
            produto               : $input.produto
            plano                 : $check.assinatura.plano
            tipo_contratacao      : $check.assinatura.tipo_contratacao
            modulos_snapshot      : $check.assinatura.modulos_json
            limites_snapshot      : $check.assinatura.limites_json
            fp_produto_catalogo_id: $check.assinatura.fp_produto_catalogo_id
            addons_json           : $check.assinatura.addons_json
            addons_pendentes_json : $check.assinatura.addons_pendentes_json
          }
        } as $regras
      
        function.run fn_fp_alacarte_addons_efetivos {
          input = {
            produto              : $input.produto
            plano                : $check.assinatura.plano
            modulos_snapshot     : $check.assinatura.modulos_json
            addons_json          : $check.assinatura.addons_json
            addons_pendentes_json: $check.assinatura.addons_pendentes_json
          }
        } as $addons_info
      
        var.update $addons_efetivos {
          value = $addons_info.addons_efetivos|first_notempty:[]
        }
      }
    
      elseif ($liberado && ($plano_efetivo|strlen) > 0) {
        // Bundle sem assinatura materializada ainda — regras do catalogo
        function.run fn_fp_plano_regras_default {
          input = {produto: $input.produto, plano: $plano_efetivo}
        } as $regras
      }
    }
  
    var $modulos_out {
      value = {}
    }
  
    var $limites_out {
      value = {}
    }
  
    var $retencao_out {
      value = 30
    }
  
    conditional {
      if ($regras != null) {
        var.update $modulos_out {
          value = $regras.modulos_json|first_notempty:{}
        }
      
        var.update $limites_out {
          value = $regras.limites_json|first_notempty:{}
        }
      
        var.update $retencao_out {
          value = $regras.retencao_dias|first_notempty:30
        }
      }
    }
  
    var $addons_pendentes {
      value = []
    }
  
    var $addons_contratados {
      value = []
    }
  
    conditional {
      if ($check.assinatura != null) {
        var.update $addons_pendentes {
          value = $check.assinatura.addons_pendentes_json|first_notempty:[]
        }
      
        var.update $addons_contratados {
          value = $check.assinatura.addons_json|first_notempty:[]
        }
      }
    }
  
    var $assinatura_out {
      value = $check.assinatura
    }
  
    conditional {
      if ($check.assinatura != null && $regras != null) {
        var.update $assinatura_out {
          value = $check.assinatura
            |set:"modulos_json":$regras.modulos_json
            |set:"limites_json":$regras.limites_json
        }
      }
    }
  }

  response = {
    liberado          : $liberado
    motivo            : $motivo
    origem            : $acesso_efetivo.origem
    plano_efetivo     : $plano_efetivo
    produto           : $input.produto
    assinatura        : $assinatura_out
    modulos_json      : $modulos_out
    limites_json      : $limites_out
    retencao_dias     : $retencao_out
    addons_efetivos   : $addons_efetivos
    addons_pendentes  : $addons_pendentes
    addons_contratados: $addons_contratados
  }
}