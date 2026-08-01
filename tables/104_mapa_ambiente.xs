table mapa_ambiente {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text descricao? filters=trim
    text imagem_url? filters=trim
    text idCliente? filters=trim
    text idFranqueado? filters=trim
    text nomeCliente? filters=trim
    text nomeFranqueado? filters=trim
    int ordem?
    bool ativo?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}