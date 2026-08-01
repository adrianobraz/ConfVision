// Lista pisos globais Break-glass
query fp_preco_piso_global_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text produto? filters=trim
    text ativo? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.breakglass || $admin.userTipo == "CEN") {
      error = "Sem permissao para listar piso global"
    }
  
    db.query fp_preco_piso_global {
      where = $db.fp_preco_piso_global.produto ==? $input.produto && $db.fp_preco_piso_global.ativo ==? $input.ativo
      sort = {
        fp_preco_piso_global.produto: "asc"
        fp_preco_piso_global.plano  : "asc"
      }
    
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista, total: $lista|count}
}