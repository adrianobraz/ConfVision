table co_rastro_disparo {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text chave? filters=trim
    int mapa_ambiente_id?
    text idCliente? filters=trim
    text idFranqueado? filters=trim
    text idProcesso? filters=trim
    text idSetor? filters=trim
    text labelSetor? filters=trim
    decimal posX?
    decimal posY?
    int sequencia?
    text predito_id_setor? filters=trim
    text direcao? filters=trim
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