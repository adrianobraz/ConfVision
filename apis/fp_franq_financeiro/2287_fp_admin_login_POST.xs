// Login admConfmonit — CEN/REP (Master + AdmFinanceiro) ou break-glass
query fp_admin_login verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text usuario? filters=trim
    text senha? filters=trim
  }

  stack {
    function.run fn_fp_admin_login {
      input = {usuario: $input.usuario, senha: $input.senha}
    } as $login_result
  }

  response = $login_result
}
