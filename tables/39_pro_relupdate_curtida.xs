table pro_RelupdateCurtida {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    // atualizacao
    int pro_relatualiza_id? {
      table = "pro_Relatualiza"
    }
  
    date? data?
    int pro_users_id? {
      table = "pro_adm_user"
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