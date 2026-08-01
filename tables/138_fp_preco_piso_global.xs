table fp_preco_piso_global {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text produto? filters=trim
    text plano? filters=trim
    text nome_exibicao? filters=trim
    decimal valor_minimo?
    text ativo?=S filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [{name: "produto", op: "asc"}, {name: "plano", op: "asc"}]
    }
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
  ]
}