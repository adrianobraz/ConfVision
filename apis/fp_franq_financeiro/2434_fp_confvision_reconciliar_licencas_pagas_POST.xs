// Recuperacao: ativa licencas pendentes de faturas pagas (admConfmonit) — ID 2434 (2433 = renovacao portal)
query fp_confvision_reconciliar_licencas_pagas verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    int fp_fatura_id?
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check

    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    conditional {
      if ($admin_check.userTipo == "REP") {
        function.run fn_fp_admin_assert_escopo_franqueado {
          input = {
            admin_token              : $input.admin_token
            id_franqueado            : $input.id_franqueado
            id_representante_registro: ""
          }
        } as $escopo
      }
    }

    var $resultado {
      value = null
    }

    conditional {
      if ($input.fp_fatura_id != null && $input.fp_fatura_id > 0) {
        function.run fn_fp_confvision_reconciliar_licencas_pagas {
          input = {
            id_franqueado: $input.id_franqueado
            fp_fatura_id : $input.fp_fatura_id
          }
        } as $resultado
      }

      else {
        function.run fn_fp_confvision_reconciliar_licencas_pagas {
          input = {id_franqueado: $input.id_franqueado}
        } as $resultado
      }
    }

    function.run fn_fp_financeiro_log {
      input = {
        acao         : "confvision_reconciliar_licencas"
        id_franqueado: $input.id_franqueado
        ref_tipo     : "franqueado"
        ref_id       : $input.id_franqueado
        detalhe      : ($resultado|get:"ativadas":0)|to_text ~ " licencas"
        origem       : "admin"
        admin_usuario: $input.admin_usuario
      }
    } as $log
  }

  response = $resultado
}
