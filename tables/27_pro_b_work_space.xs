table proB_WorkSpace {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int pro_adm_id? {
      table = "Pro_adm"
    }
  
    int proa_tipoprojeto_id? {
      table = "proA_TipoProjeto"
    }
  
    date? data?
    bool Aberto?
    text areatrabalho? filters=trim
    text descricao? filters=trim
    date? fechado?
    int Proprietario? {
      table = "pro_adm_user"
    }
  
    int[] Membros? {
      table = "pro_adm_user"
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}