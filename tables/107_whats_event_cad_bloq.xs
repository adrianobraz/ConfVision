table WhatsEventCadBloq {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idFranqueado? filters=trim
    text idCliente? filters=trim
    text nomeCliente? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "idFranqueado", op: "asc"}]}
    {type: "btree", field: [{name: "idCliente", op: "asc"}]}
    {type: "btree", field: [{name: "nomeCliente", op: "asc"}]}
  ]
}