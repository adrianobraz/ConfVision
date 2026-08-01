table status_link {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text token? filters=trim
    text idDispositivo? filters=trim
    timestamp? expiresAt?
    int alarm_events_id? {
      table = "alarm_events"
    }
  
    text particao? filters=trim
    text zonauser? filters=trim
    text grupo? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "token", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "zonauser", op: "asc"}
        {name: "alarm_events_id", op: "asc"}
        {name: "created_at", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "zonauser", op: "asc"}
        {name: "expiresAt", op: "asc"}
        {name: "grupo", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "alarm_events_id", op: "asc"}
        {name: "grupo", op: "asc"}
      ]
    }
  ]
}