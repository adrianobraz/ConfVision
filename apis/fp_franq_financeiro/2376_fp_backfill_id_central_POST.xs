// Dispara backfill de id_central (assinatura/fatura) — soh CEN/break-glass
query fp_backfill_id_central verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    precondition (($escopo|get:"userTipo":"") == "CEN" || ($escopo|get:"breakglass":false)) {
      error = "Somente Central ou Break-glass pode executar backfill"
    }
  
    function.run fn_fp_backfill_id_central {
      input = {limite: 2000}
    } as $resultado
  }

  response = $resultado
}