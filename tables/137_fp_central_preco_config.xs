table fp_central_preco_config {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_central? filters=trim
    text modo_preco?=livre filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
    {type: "btree", field: [{name: "modo_preco", op: "asc"}]}
  ]
}