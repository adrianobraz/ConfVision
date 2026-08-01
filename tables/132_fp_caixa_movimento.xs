table fp_caixa_movimento {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text tipo?=entrada filters=trim
    decimal valor?
    text descricao? filters=trim
    timestamp? movimento_em?
    text admin_usuario? filters=trim
    text ref_tipo? filters=trim
    text ref_id? filters=trim
    text id_franqueado? filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "movimento_em", op: "asc"}]}
    {type: "btree", field: [{name: "tipo", op: "asc"}]}
    {type: "btree", field: [{name: "ref_tipo", op: "asc"}]}
    {type: "btree", field: [{name: "ref_id", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "ref_tipo", op: "asc"}, {name: "ref_id", op: "asc"}]
    }
  ]
}