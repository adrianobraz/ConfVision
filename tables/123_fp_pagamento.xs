table fp_pagamento {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int fp_fatura_id? {
      table = "fp_fatura"
    }
  
    decimal valor?
    text metodo?=manual filters=trim
    timestamp? pago_em?
    text id_externo? filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "fp_fatura_id", op: "asc"}]}
    {type: "btree", field: [{name: "pago_em", op: "asc"}]}
  ]
}