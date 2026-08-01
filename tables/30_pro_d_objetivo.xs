table proD_Objetivo {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    // avaliacao
    int[] proda_objavaliacao_id? {
      table = "proDA_objAvaliacao"
    }
  
    int colocaboradores? {
      table = "pro_adm_user"
    }
  
    timestamp? controletempo?
    date? controlTime?
    date? cronograma?
    date? data?
  
    // depende_de
    int prod_objetivo_id? {
      table = "proD_Objetivo"
    }
  
    // equipe
    int prob_workspace_id? {
      table = "proB_WorkSpace"
    }
  
    int esforco_estimado?
  
    // fase
    int pro_st_fase_id? {
      table = "pro_ST_fase"
    }
  
    text guia_criativo? filters=trim
    text hora? filters=trim
    text id_objetivo? filters=trim
  
    // importancia
    int pro_st_importancia_id? {
      table = "pro_ST_importancia"
    }
  
    // Lider
    int pro_users_id? {
      table = "pro_adm_user"
    }
  
    text link? filters=trim
    text[] ListaEtiquetas? filters=trim
    bool marcar?
    text notas? filters=trim
    int numeros?
  
    // objDescricao
    int proc_objetivodesc_id? {
      table = "proC_objetivoDesc"
    }
  
    // prioridade
    int pro_st_prioridade_id? {
      table = "pro_ST_prioridade"
    }
  
    // progresso
    int pro_st_progresso_id? {
      table = "pro_ST_progresso"
    }
  
    int progressoBar?
    text requisitosFuncional? filters=trim
    text semana? filters=trim
    text[] tags? filters=trim
  
    // tarefas
    int prodb_tarefas_id? {
      table = "proDB_tarefas"
    }
  
    text telefone? filters=trim
    text Tema? filters=trim
    text Texto? filters=trim
  
    // voto
    int prodc_objvoto_id? {
      table = "proDC_objvoto"
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