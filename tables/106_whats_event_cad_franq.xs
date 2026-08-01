table WhatsEventCadFranq {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idFranqueado? filters=trim
    text telMonitoramento? filters=trim
    text telViatura? filters=trim
    text telGerente? filters=trim
    text horaIni? filters=trim
    text horaFin? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}