table fp_pacote_cota {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_central?=CENTRAL filters=trim
    text nome? filters=trim
    int quantidade?
    decimal valor?
    text ativo?=S filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [{name: "id_central", op: "asc"}, {name: "ativo", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "id_central", op: "asc"}
        {name: "quantidade", op: "asc"}
      ]
    }
  ]
}