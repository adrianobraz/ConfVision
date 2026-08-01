table pro_ST_fase {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text cor? filters=trim
    text descricao? filters=trim
    int ordem_exibicao?
    int proa_tipoprojeto_id? {
      table = "proA_TipoProjeto"
    }
  
    int pro_adm_id? {
      table = "Pro_adm"
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}