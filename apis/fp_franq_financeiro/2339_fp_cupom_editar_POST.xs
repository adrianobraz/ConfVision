// Edita cupom ativo (ainda nao usado) — admConfmonit
query fp_cupom_editar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int cupom_id? filters=min:1
    text codigo? filters=trim
    text tipo?=percentual filters=trim
    decimal valor?
    text produto?=franqueadopro filters=trim
    text id_franqueado? filters=trim
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    function.run fn_fp_cupom_editar {
      input = {
        cupom_id     : $input.cupom_id
        codigo       : $input.codigo
        tipo         : $input.tipo
        valor        : $input.valor
        produto      : $input.produto
        id_franqueado: $input.id_franqueado
        observacao   : $input.observacao
        admin_usuario: $input.admin_usuario
      }
    } as $resultado
  }

  response = $resultado
}