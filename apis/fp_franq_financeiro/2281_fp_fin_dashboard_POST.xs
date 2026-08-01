// Dashboard financeiro admin (KPIs consolidados no escopo CEN/REP)
query fp_fin_dashboard verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $id_rep {
      value = $escopo|get:"idRepresentante":""
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    var $usuario {
      value = $escopo|get:"usuario":""
    }
  
    var $perm_global {
      value = $escopo|get:"permite_global":false
    }
  
    function.run fn_fp_fin_resumo {
      input = {
        competencia     : $input.competencia
        id_representante: $id_rep
        id_central      : $id_cen
        admin_usuario   : $usuario
        permitir_global : $perm_global
      }
    } as $resumo
  }

  response = $resumo
}
