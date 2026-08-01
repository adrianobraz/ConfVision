table UserCliConf {
  auth = false

  schema {
    uuid id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    email email filters=trim|lower
    password senha? {
      sensitive = true
      visibility = "internal"
    }
  
    date? Data?=now
    text idFranqueado? filters=trim
    text idCliente? filters=trim
    text nome? filters=trim
    text telefone1? filters=trim
    text telefone2? filters=trim
    text Descricao? filters=trim
    enum[] Acesso? {
      values = [
        "MEUSDADOS"
        "GRADEHORARIO"
        "MSGATENDENTE"
        "RELATORIO"
        "WHATSAPP"
        "CAMERA"
      ]
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "id", op: "asc"}]}
    {type: "btree", field: [{name: "email", op: "asc"}]}
    {type: "btree|unique", field: [{name: "email", op: "asc"}]}
    {type: "btree", field: [{name: "idCliente", op: "asc"}]}
  ]
}