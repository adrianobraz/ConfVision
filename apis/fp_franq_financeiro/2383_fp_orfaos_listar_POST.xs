// Break-glass: lista dados orfaos / inconsistencias (assinatura/fatura)
query fp_orfaos_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_franqueado? filters=trim
    int limite?=200
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    precondition (($escopo|get:"breakglass":false) == true) {
      error = "Somente Break-glass pode listar dados orfaos"
    }
  
    function.run fn_fp_orfaos_listar {
      input = {
        id_franqueado: $input.id_franqueado
        limite       : $input.limite|first_notempty:200
      }
    } as $resultado
  }

  response = $resultado
}
