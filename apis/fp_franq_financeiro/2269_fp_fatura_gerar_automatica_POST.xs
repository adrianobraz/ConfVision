// Worker — gera fatura automatica para assinatura
query fp_fatura_gerar_automatica verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text worker_key? filters=trim
    int assinatura_id? filters=min:1
    text ciclo_ref? filters=trim
  }

  stack {
    function.run fn_fp_worker_validar {
      input = {worker_key: $input.worker_key}
    } as $worker_check
  
    function.run fn_fp_fatura_gerar {
      input = {
        assinatura_id: $input.assinatura_id
        ciclo_ref    : $input.ciclo_ref
        origem       : "worker"
        acao_log     : "fatura_gerar_automatica"
      }
    } as $resultado
  }

  response = $resultado
}