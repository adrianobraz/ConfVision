table UserCliConf_Equipamentos {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idDisp? filters=trim
    uuid? usercliconf_id? {
      table = "UserCliConf"
    }
  
    text DispApp? filters=trim
    text idCliente? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "usercliconf_id", op: "asc"}]}
    {type: "btree", field: [{name: "idDisp", op: "asc"}]}
    {type: "btree", field: [{name: "idCliente", op: "asc"}]}
  ]
}