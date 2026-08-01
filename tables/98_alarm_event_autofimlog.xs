table alarmEvent_autofimlog {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int idEvento?
    text idProcesso? filters=trim
    text motivo? filters=trim
    text acao? filters=trim
    int qtdCiclos5m?
    bool bloqueio3x5?
    bool temFalhas?
    bool temAlarme?
    bool temDesarme?
    bool temRestaure?
    bool temParAlarmeRest50?
    json retornoProcessoEnd?
    text idDispositivo? filters=trim
    text regraVersao?=v1 filters=trim
    text erro? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "idEvento", op: "asc"}]}
    {type: "btree", field: [{name: "id", op: "asc"}]}
    {type: "btree", field: [{name: "idProcesso", op: "asc"}]}
  ]
}