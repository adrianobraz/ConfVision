table pro_Relatualiza {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text conteudo? filters=trim
    date? data?
  
    // ID_Atualizacao
    int pro_relatualiza_id? {
      table = "pro_Relatualiza"
    }
  
    // mencoes
    int[] pro_users_id? {
      table = "pro_adm_user"
    }
  
    // tarefa
    int prodb_tarefas_id? {
      table = "proDB_tarefas"
    }
  
    // user
    int pro_users2_id? {
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