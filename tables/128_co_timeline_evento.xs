table co_timeline_evento {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text chave? filters=trim
    int mapa_ambiente_id?
    int idCliente?
    int idFranqueado?
    int idProcesso?
    int idSetor?
    int idDispositivo?
    int idOperador?
    text nomeOperador? filters=trim
    text tipo? filters=trim
    text origem? filters=trim
    text codigo? filters=trim
    text zonaUser? filters=trim
    text particao? filters=trim
    text texto? filters=trim
    timestamp? evento_ts?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "chave", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "mapa_ambiente_id", op: "asc"}]
    }
    {type: "btree", field: [{name: "idProcesso", op: "asc"}]}
    {type: "btree", field: [{name: "evento_ts", op: "asc"}]}
  ]
}