// Lista planos ConfVision (18 slugs) para admConfmonit / ConfVision
query fp_confvision_planos_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    bool sincronizar?
    text admin_usuario? filters=trim
  }

  stack {
    conditional {
      if (($input.admin_token|is_empty) == false) {
        function.run fn_fp_admin_validar {
          input = {admin_token: $input.admin_token}
        } as $admin_check
      }
    }
  
    function.run fn_fp_confvision_planos_listar {
      input = {id_franqueado: $input.id_franqueado}
    } as $resultado
  
    var $sync {
      value = null
    }
  
    conditional {
      if ($input.sincronizar) {
        precondition (($input.id_franqueado|is_empty) == false) {
          error = "id_franqueado obrigatorio"
        }
      
        function.run fn_franqueado_sync_usa_confvision {
          input = {id_franqueado: $input.id_franqueado}
        } as $sync
      
        var $det_sync {
          value = $sync.acesso.motivo
        }
      
        var.update $det_sync {
          value = $det_sync ~ " -> " ~ $sync.acesso.usa_confvision
        }
      
        function.run fn_fp_financeiro_log {
          input = {
            acao         : "confvision_sync_manual"
            id_franqueado: $input.id_franqueado
            ref_tipo     : "franqueado"
            ref_id       : $input.id_franqueado
            detalhe      : $det_sync
            origem       : "admin"
            admin_usuario: $input.admin_usuario
          }
        } as $log_sync
      }
    }
  }

  response = {
    planos           : $resultado.planos
    desconto_pro_plus: $resultado.desconto_pro_plus
    total            : $resultado.total
    sync             : $sync
  }
}