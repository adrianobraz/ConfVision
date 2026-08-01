table alarmEvent_Finalizados {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idprocesso? filters=trim
    text telefone? filters=trim
    text token? filters=trim
    text mensagem? filters=trim
    text ip? filters=trim
    text ipcidade? filters=trim
    text ipestado? filters=trim
    text ipcep? filters=trim
    int geolat?
    int geolon?
    text georua? filters=trim
    text geonumero? filters=trim
    text geobairro? filters=trim
    text geocidade? filters=trim
    text geoestado? filters=trim
    text device? filters=trim
    text browser? filters=trim
    text sitema? filters=trim
    text timezone? filters=trim
    text fingerprint? filters=trim
    text idUsuario? filters=trim
    text Usuario? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}