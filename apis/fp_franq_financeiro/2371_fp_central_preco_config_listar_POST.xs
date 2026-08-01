// Lista Centrais + modo_preco — somente Break-glass
query fp_central_preco_config_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.breakglass == true) {
      error = "Somente Break-glass pode gerenciar modo de preco das Centrais"
    }
  
    function.run fn_fp_centrais_listar {
      input = {}
    } as $lista
  }

  response = $lista
}
