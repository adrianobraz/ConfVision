table UserCliConfLog {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    uuid? usercliconf_id? {
      table = "UserCliConf"
    }
  
    text idFranqueado? filters=trim
    text idCliente? filters=trim
    date? Data?=now
    text Descricao? filters=trim
    text Ocorrencia? filters=trim
    text Evento? filters=trim
    text log? filters=trim
    text idDisp? filters=trim
    text DispApp? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}