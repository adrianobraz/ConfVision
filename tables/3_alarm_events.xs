table alarm_events {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idEvento? filters=trim
    text codigo? filters=trim
    text particao? filters=trim
    text zonaUser? filters=trim
    text nivel? filters=trim
    text dataEntrada? filters=trim
    text img? filters=trim
    text idProcesso? filters=trim
    text idDispositivo? filters=trim
    text idCliente? filters=trim
    text nomeCliente? filters=trim
    text emailCliente? filters=trim
    text ctiGrupo? filters=trim
    text ctiDescricao? filters=trim
    text idFranqueado? filters=trim
    text codigoBenuvem? filters=trim
    text conta? filters=trim
    text carmeraAtiva? filters=trim
    date? Data?=now
    text usa_confvision?=N filters=trim
    text provedor_video?=nenhum filters=trim
    int vis_evento_id? {
      table = "vis_evento"
    }
  
    text snapshot_url? filters=trim
    bool tem_midia?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "idCliente", op: "asc"}]}
    {type: "btree", field: [{name: "idFranqueado", op: "asc"}]}
    {type: "btree", field: [{name: "codigoBenuvem", op: "asc"}]}
    {type: "btree", field: [{name: "emailCliente", op: "asc"}]}
    {type: "btree", field: [{name: "idEvento", op: "asc"}]}
    {type: "btree", field: [{name: "Data", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "id", op: "desc"}, {name: "nivel", op: "desc"}]
    }
    {type: "btree", field: [{name: "ctiDescricao", op: "asc"}]}
    {type: "btree", field: [{name: "ctiGrupo", op: "asc"}]}
    {type: "btree", field: [{name: "codigo", op: "asc"}]}
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {type: "btree", field: [{name: "idProcesso", op: "asc"}]}
    {type: "btree", field: [{name: "dataEntrada", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "idProcesso", op: "asc"}
        {name: "idDispositivo", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "zonaUser", op: "asc"}
        {name: "idCliente", op: "asc"}
        {name: "ctiGrupo", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idProcesso", op: "asc"}
        {name: "created_at", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "created_at", op: "asc"}
      ]
    }
  ]
}