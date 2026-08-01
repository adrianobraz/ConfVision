// Detalhe da fatura com itens
query fp_fatura_get verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    int fatura_id? filters=min:1
  }

  stack {
    db.get fp_fatura {
      field_name = "id"
      field_value = $input.fatura_id
    } as $fatura
  
    precondition ($fatura != null) {
      error = "Fatura nao encontrada"
    }
  
    db.query fp_fatura_item {
      where = $db.fp_fatura_item.fp_fatura_id == $input.fatura_id
      sort = {fp_fatura_item.id: "asc"}
      return = {type: "list"}
    } as $itens
  
    db.query fp_pagamento {
      where = $db.fp_pagamento.fp_fatura_id == $input.fatura_id
      sort = {fp_pagamento.pago_em: "desc"}
      return = {type: "list"}
    } as $pagamentos
  }

  response = {
    fatura    : $fatura
    itens     : $itens
    pagamentos: $pagamentos
  }
}