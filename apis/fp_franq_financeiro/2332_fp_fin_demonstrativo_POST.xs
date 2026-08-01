// Demonstrativo de receita e despesas / DRE gerencial / balancete
query fp_fin_demonstrativo verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
    timestamp? data_inicio?
    timestamp? data_fim?
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    function.run fn_fp_fin_resumo {
      input = {
        competencia     : $input.competencia
        data_inicio     : $input.data_inicio
        data_fim        : $input.data_fim
        id_representante: $escopo|get:"idRepresentante":""
        id_central      : $escopo|get:"idCentral":""
        admin_usuario   : $escopo|get:"usuario":""
        permitir_global : $escopo|get:"permite_global":false
      }
    } as $resumo
  }

  response = {
    receita_bruta         : $resumo.receita_periodo
    despesas_retiradas    : $resumo.retiradas_periodo
    despesas_contas_pagar : $resumo.despesas_periodo
    despesas_repasses     : $resumo.repasses_periodo
    total_despesas        : $resumo.total_despesas_periodo
    resultado             : $resumo.resultado_periodo
    valor_em_aberto       : $resumo.valor_em_aberto
    valor_pago            : $resumo.valor_pago
    valor_vencido         : $resumo.valor_vencido
    saldo_caixa           : $resumo.saldo_caixa
    contas_a_receber      : $resumo.valor_em_aberto
    contas_a_pagar        : $resumo.valor_a_pagar
    valor_repasses_abertos: $resumo.valor_repasses_abertos
  }
}
