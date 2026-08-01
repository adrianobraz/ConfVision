table fp_admin_sessao {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text token? filters=trim
    text id_usuario? filters=trim
    // CEN = ID_Central; REP = ID_Representante
    text id_vinculo? filters=trim
    // ID_Central do catalogo (obrigatorio para REP; igual a id_vinculo para CEN)
    text id_central? filters=trim
    text user_tipo? filters=trim
    text usuario? filters=trim
    text nome? filters=trim
    text master?=N filters=trim
    text breakglass?=N filters=trim
    text ativo?=S filters=trim
    timestamp? expira_em?
    timestamp? ultimo_acesso_em?
    text id_central? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "token", op: "asc"}]}
    {type: "btree", field: [{name: "id_usuario", op: "asc"}]}
    {type: "btree", field: [{name: "user_tipo", op: "asc"}]}
    {type: "btree", field: [{name: "ativo", op: "asc"}]}
  ]
}