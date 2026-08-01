table fp_fatura_item {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int fp_fatura_id? {
      table = "fp_fatura"
    }
  
    text descricao? filters=trim
    int quantidade?=1
    decimal valor_unitario?
    decimal valor_total?
    text ref_tipo? filters=trim
    text ref_id? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "fp_fatura_id", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "ref_tipo", op: "asc"}, {name: "ref_id", op: "asc"}]
    }
  ]
}