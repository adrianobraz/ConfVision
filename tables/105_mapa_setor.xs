table mapa_setor {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int mapa_ambiente_id? {
      table = "mapa_ambiente"
    }
  
    text idSetor? filters=trim
    text idDispositivo? filters=trim
    text idCliente? filters=trim
    text setornome? filters=trim
    text dispositivoNome? filters=trim
    decimal posX?
    decimal poxY?
    text icone? filters=trim
    text label? filters=trim
    text descricao? filters=trim
    text clienteNome? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}