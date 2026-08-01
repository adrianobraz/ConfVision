table vis_cliente_grade_exec {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }

    int slot_id? {
      table = "vis_cliente_grade_slot"
    }

    text data_ref? filters=trim
    text hora_ref? filters=trim
    timestamp? executado_em?
    text resultado? filters=trim
    json detalhe?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree|unique"
      field: [
        {name: "slot_id", op: "asc"}
        {name: "data_ref", op: "asc"}
        {name: "hora_ref", op: "asc"}
      ]
    }
  ]
}
