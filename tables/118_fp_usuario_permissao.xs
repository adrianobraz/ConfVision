table fp_usuario_permissao {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    // ID do franqueado (vinculo)
    text id_franqueado? filters=trim
  
    // ID do responsavel (usuarios.ID_Usuario MySQL)
    text id_usuario? filters=trim
  
    // Chave do menu ex: relatorio, relatorio.eventos
    text chave_menu? filters=trim
  
    // S=liberado N=bloqueado
    text liberado?=S filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "id_usuario", op: "asc"}
        {name: "chave_menu", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "id_usuario", op: "asc"}
      ]
    }
  ]
}