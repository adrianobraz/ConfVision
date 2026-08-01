// Fecha o mes gravando snapshot (nao bloqueia lancamentos retroativos)
query fp_fin_fechamento_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
    text observacao? filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    precondition (($escopo|get:"userTipo":"") != "REP") {
      error = "Fechamento mensal e exclusivo da Central"
    }
  
    precondition (($input.competencia|strlen) >= 7) {
      error = "Informe a competencia no formato YYYY-MM"
    }
  
    db.query fp_fechamento_mes {
      where = $db.fp_fechamento_mes.competencia == $input.competencia && $db.fp_fechamento_mes.status == "fechado"
      return = {type: "list"}
    } as $existentes
  
    precondition (($existentes|count) == 0) {
      error = "Esta competencia ja foi fechada"
    }
  
    var $id_cen {
      value = $escopo|get:"idCentral":""
    }
  
    function.run fn_fp_fin_resumo {
      input = {
        competencia     : $input.competencia
        id_central      : $id_cen
        admin_usuario   : $escopo|get:"usuario":""
        permitir_global : $escopo|get:"permite_global":false
      }
    } as $resumo
  
    db.add fp_fechamento_mes {
      data = {
        created_at           : "now"
        competencia          : $input.competencia
        status               : "fechado"
        fechado_em           : now
        fechado_por          : $input.admin_usuario|first_notempty:($escopo|get:"usuario":"")
        id_central           : $id_cen
        saldo_caixa_inicio   : 0
        saldo_caixa_fim      : $resumo.saldo_caixa
        total_entradas       : $resumo.entradas_periodo
        total_saidas         : $resumo.total_despesas_periodo
        total_receitas       : $resumo.receita_periodo
        total_despesas       : $resumo.total_despesas_periodo
        total_a_receber      : $resumo.valor_em_aberto
        total_vencido        : $resumo.valor_vencido
        qtd_inadimplentes    : $resumo.qtd_inadimplentes
        assinaturas_ativas   : $resumo.assinaturas_ativas
        assinaturas_pendentes: $resumo.assinaturas_pendentes
        faturas_pagas        : $resumo.faturas_pagas
        faturas_abertas      : $resumo.faturas_abertas
        snapshot_json        : $resumo|json_encode
        observacao           : $input.observacao
      }
    } as $fechamento
  
    function.run fn_fp_financeiro_log {
      input = {
        acao         : "fechamento_mes"
        ref_tipo     : "fp_fechamento_mes"
        ref_id       : $fechamento.id|to_text
        detalhe      : "competencia=" ~ $input.competencia
        origem       : "admin"
        valor        : $resumo.resultado_periodo
        admin_usuario: $input.admin_usuario|first_notempty:($escopo|get:"usuario":"")
        id_central   : $id_cen
      }
    } as $log
  }

  response = {fechamento: $fechamento, resumo: $resumo}
}