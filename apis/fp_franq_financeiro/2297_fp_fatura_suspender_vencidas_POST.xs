// Worker — suspende assinaturas com fatura aberta vencida
query fp_fatura_suspender_vencidas verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text worker_key? filters=trim
    text admin_token? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    conditional {
      if (($input.worker_key|is_empty) == false) {
        function.run fn_fp_worker_validar {
          input = {worker_key: $input.worker_key}
        } as $worker_check
      }
    
      else {
        function.run fn_fp_admin_validar {
          input = {admin_token: $input.admin_token}
        } as $admin_check
      }
    }
  
    function.run "" {
      input = {
        origem       : ($input.worker_key|is_empty) == false ? "worker" : "admin"
        admin_usuario: $input.admin_usuario
      }
    } as $resultado
  }

  response = $resultado
}