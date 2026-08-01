table fp_preco_pacote_cota {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_representante? filters=trim
    int fp_pacote_cota_id? {
      table = "fp_pacote_cota"
    }
  
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
        {name: "fp_pacote_cota_id", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
  ]
}