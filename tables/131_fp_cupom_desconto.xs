table fp_cupom_desconto {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text codigo? filters=trim
    text tipo?=percentual filters=trim
    decimal valor?
    text produto?=franqueadopro filters=trim
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text id_central? filters=trim
    text criado_por_tipo? filters=trim
    text alvo_nivel? filters=trim
    text status?=ativo filters=trim
    text criado_por? filters=trim
    text observacao? filters=trim
    timestamp? usado_em?
    text usado_por_id_franqueado? filters=trim
    text ref_tipo? filters=trim
    text ref_id? filters=trim
    decimal valor_base?
    decimal valor_desconto?
    decimal valor_final?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "status", op: "asc"}]}
    {type: "btree", field: [{name: "codigo", op: "asc"}]}
    {type: "btree", field: [{name: "produto", op: "asc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "id_representante", op: "asc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
    {type: "btree", field: [{name: "criado_por_tipo", op: "asc"}]}
    {type: "btree", field: [{name: "alvo_nivel", op: "asc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
  ]
}
