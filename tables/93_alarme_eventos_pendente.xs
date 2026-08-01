table alarmeEventos_pendente {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idDispositivo? filters=trim
    text zonaUser? filters=trim
    int id_tblAlarm_events? {
      table = "alarm_events"
    }
  
    timestamp? agendadoPara?
    int janelaSegundos?
    bool cancelado?
    bool enviado?
    text grupo? filters=trim
    text particao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "zonaUser", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "cancelado", op: "asc"}
        {name: "enviado", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "cancelado", op: "asc"}
        {name: "enviado", op: "asc"}
        {name: "agendadoPara", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "zonaUser", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "id_tblAlarm_events", op: "asc"}
      ]
    }
  ]
}