table vis_cliente_grade_escopo {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }

    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    bool grade_ativa?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_cliente", op: "asc"}]}
    {type: "btree", field: [{name: "id_dispositivo", op: "asc"}]}
    {
      type : "btree|unique"
      field: [
        {name: "id_cliente", op: "asc"}
        {name: "id_franqueado", op: "asc"}
        {name: "id_dispositivo", op: "asc"}
      ]
    }
  ]
}
