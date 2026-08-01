// Lista movimentacoes de caixa exclusivas do usuario logado
query fp_caixa_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text tipo? filters=trim
    int limite?=200 filters=min:1|max:500
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin_check
  
    var $usuario {
      value = $admin_check.usuario|first_notempty:""|trim
    }
  
    precondition (($usuario|is_empty) == false) {
      error = "Sessao sem usuario — faca logout e login novamente"
    }
  
    db.query fp_caixa_movimento {
      where = ($db.fp_caixa_movimento.id > 0) && ($db.fp_caixa_movimento.tipo ==? $input.tipo) && ($db.fp_caixa_movimento.admin_usuario == $usuario)
      sort = {fp_caixa_movimento.movimento_em: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista, total: $lista|count}
}
