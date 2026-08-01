table fp_conta_pagar {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text fornecedor? filters=trim
    text descricao? filters=trim
    text categoria? filters=trim
    decimal valor?
    timestamp? vencimento_em?
    text status?=aberta filters=trim
    timestamp? pago_em?
    text admin_usuario? filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "status", op: "asc"}]}
    {type: "btree", field: [{name: "vencimento_em", op: "asc"}]}
  ]
}