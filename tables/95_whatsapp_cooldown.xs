table whatsapp_cooldown {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text idDispositivo? filters=trim
    text ctiGrupo? filters=trim
    text ctiDescricao? filters=trim
    int totalEventos?
    timestamp? ultimoEnvio?
    text zonaUser? filters=trim
    text particao? filters=trim
    text idCliente? filters=trim
    text nomeCliente? filters=trim
    text emailCliente? filters=trim
    text idFranqueado? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "ctiDescricao", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "zonaUser", op: "asc"}
        {name: "ctiDescricao", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "ctiGrupo", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "ctiGrupo", op: "asc"}
        {name: "idDispositivo", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "idDispositivo", op: "asc"}]}
    {type: "btree", field: [{name: "ctiGrupo", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "idDispositivo", op: "asc"}
        {name: "idCliente", op: "asc"}
        {name: "ctiGrupo", op: "asc"}
      ]
    }
  ]
}