table vis_camera_area {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int vis_camera_id? {
      table = "vis_camera"
    }
  
    text nome? filters=trim
    bool ativo?
    text poligono_json? filters=trim
    text cor? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "vis_camera_id", op: "asc"}]}
    {type: "btree", field: [{name: "nome", op: "asc"}]}
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
  ]
}