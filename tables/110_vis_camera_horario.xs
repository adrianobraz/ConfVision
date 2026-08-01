table vis_camera_horario {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int vis_camera_id? {
      table = "vis_camera"
    }
  
    int dia_semana?
    text hora_inicio? filters=trim
    text hora_fim? filters=trim
    bool ativo?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}