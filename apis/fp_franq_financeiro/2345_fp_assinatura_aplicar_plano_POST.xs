// Aplica snapshot de modulos/limites do catalogo na assinatura (admin)
query fp_assinatura_aplicar_plano verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int assinatura_id? filters=min:1
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    precondition ($input.assinatura_id != null) {
      error = "assinatura_id obrigatorio"
    }
  
    function.run fn_fp_assinatura_aplicar_catalogo {
      input = {assinatura_id: $input.assinatura_id}
    } as $model
  }

  response = $model
}