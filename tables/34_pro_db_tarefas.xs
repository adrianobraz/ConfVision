table proDB_tarefas {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    // Ciclo
    int prodba_ciclo_naosei_id? {
      table = "proDBA_ciclo"
    }
  
    bool ciclo_ativo?
    bool ciclo_concluido?
    text Descricao? filters=trim
    int esforco_estimado?
    int Esforco_real?
    text id_tarefa? filters=trim
    bool naoprevisto?
    text nome? filters=trim
    int prod_objetivo_id? {
      table = "proD_Objetivo"
    }
  
    // Responsavel
    int pro_users_id? {
      table = "pro_adm_user"
    }
  
    // tarDescricao
    int prodbb_tardescricao_id? {
      table = "proDBB_tarDescricao"
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