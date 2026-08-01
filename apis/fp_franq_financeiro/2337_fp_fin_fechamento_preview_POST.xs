// Preview do fechamento do mes (nao grava)
query fp_fin_fechamento_preview verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text competencia? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    precondition (($input.competencia|strlen) >= 7) {
      error = "Informe a competencia no formato YYYY-MM"
    }
  
    function.run fn_fp_fin_resumo {
      input = {
        competencia     : $input.competencia
        id_representante: $escopo|get:"idRepresentante":""
        id_central      : $escopo|get:"idCentral":""
        admin_usuario   : $escopo|get:"usuario":""
        permitir_global : $escopo|get:"permite_global":false
      }
    } as $resumo
  
    db.query fp_fechamento_mes {
      where = $db.fp_fechamento_mes.competencia == $input.competencia && $db.fp_fechamento_mes.status == "fechado"
      return = {type: "list"}
    } as $existentes
  }

  response = {
    competencia: $input.competencia
    ja_fechado : ($existentes|count) > 0
    preview    : $resumo
    idCentral  : $escopo|get:"idCentral":""
  }
}
