table fp_produto_catalogo {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
  
  
    // Pacote de cota vinculado (somente franqueadopro): licenca + cota = total
    int fp_pacote_cota_id? {
      table = "fp_pacote_cota"
    }
  

  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "id_central", op: "asc"}
        {name: "id_representante", op: "asc"}
        {name: "produto", op: "asc"}
        {name: "plano", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
  ]
}
