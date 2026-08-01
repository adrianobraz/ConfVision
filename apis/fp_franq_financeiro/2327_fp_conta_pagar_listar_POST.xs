// Lista contas a pagar exclusivas do usuario logado
query fp_conta_pagar_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text status? filters=trim
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
  
    db.query fp_conta_pagar {
      where = ($db.fp_conta_pagar.id > 0) && ($db.fp_conta_pagar.status ==? $input.status) && ($db.fp_conta_pagar.admin_usuario == $usuario)
      sort = {fp_conta_pagar.vencimento_em: "asc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista, total: $lista|count}
}
