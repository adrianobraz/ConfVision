// Admin — gera ou adianta fatura do ciclo atual da assinatura
query fp_fatura_gerar_manual verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    int assinatura_id? filters=min:1
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    precondition ($input.assinatura_id != null) {
      error = "assinatura_id obrigatorio"
    }
  
    function.run fn_fp_fatura_gerar {
      input = {
        assinatura_id: $input.assinatura_id
        origem       : "admin"
        admin_usuario: $input.admin_usuario
        acao_log     : "fatura_gerar_manual"
      }
    } as $resultado
  }

  response = $resultado
}