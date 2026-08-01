table bot_finalizaeventoauto {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    timestamp? ultimo_evento_ts?
    text status? filters=trim
    timestamp? rodar_em?
    int tentativas?
    text? lock_token? filters=trim
    timestamp? lock_at?
    text motivo_final? filters=trim
    text acao_final? filters=trim
    text erro? filters=trim
    json payload_resultado?
    timestamp? updated_at?=now
    int alarm_events_id? {
      table = "alarm_events"
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [{name: "status", op: "asc"}, {name: "rodar_em", op: "asc"}]
    }
    {type: "btree", field: [{name: "idProcesso", op: "asc"}]}
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "alarm_events_id", op: "asc"}]
    }
  ]
}