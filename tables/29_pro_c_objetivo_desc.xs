table proC_objetivoDesc {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int prob_workspace_id? {
      table = "proB_WorkSpace"
    }
  
    date? Data?
    text Descricao? filters=trim
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