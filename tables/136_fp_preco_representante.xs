table fp_preco_representante {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_representante? filters=trim
    int fp_produto_catalogo_id? {
      table = "fp_produto_catalogo"
    }
  
    text produto? filters=trim
    text plano? filters=trim
    decimal valor_venda?
    text ativo?=S filters=trim
    text observacao? filters=trim
    timestamp? atualizado_em?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [{name: "id_representante", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "id_representante", op: "asc"}
        {name: "fp_produto_catalogo_id", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "id_representante", op: "asc"}
        {name: "produto", op: "asc"}
        {name: "plano", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
  ]
}