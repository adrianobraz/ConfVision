table fp_modulo_catalogo {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_central?=CENTRAL filters=trim
    text id_representante? filters=trim
    text produto?=franqueadopro filters=trim
    text chave? filters=trim
    text label? filters=trim
    text grupo? filters=trim
    text descricao? filters=trim
    text plano_minimo?=lite filters=trim
    decimal valor_mensal?
    decimal valor_piso_breakglass?
    text ativo?=S filters=trim
    int ordem?
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
        {name: "chave", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
    {type: "btree", field: [{name: "ordem", op: "asc"}]}
  ]
}
