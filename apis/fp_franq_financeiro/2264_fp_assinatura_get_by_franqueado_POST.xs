// Assinaturas de um franqueado
query fp_assinatura_get_by_franqueado verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text id_franqueado? filters=trim
    text produto? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    db.query fp_assinatura_produto {
      where = $db.fp_assinatura_produto.id_franqueado == $input.id_franqueado && $db.fp_assinatura_produto.produto ==? $input.produto
      sort = {fp_assinatura_produto.created_at: "desc"}
      return = {type: "list"}
    } as $lista
  }

  response = {dados: $lista, total: $lista|count}
}