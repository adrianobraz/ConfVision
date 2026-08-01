// Cancela cupom ativo — admConfmonit
query fp_cupom_cancelar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int cupom_id? filters=min:1
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    function.run fn_fp_cupom_cancelar {
      input = {
        cupom_id     : $input.cupom_id
        observacao   : $input.observacao
        admin_usuario: $input.admin_usuario
      }
    } as $resultado
  }

  response = $resultado
}