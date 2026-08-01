table fp_financeiro_log {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text acao? filters=trim
    text id_franqueado? filters=trim
    text ref_tipo? filters=trim
    text ref_id? filters=trim
    text detalhe? filters=trim
    text origem? filters=trim
    decimal valor?
    text produto? filters=trim
    text plano? filters=trim
    text admin_usuario? filters=trim
    text id_central? filters=trim
    text id_representante? filters=trim
    text id_usuario? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "acao", op: "asc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
    {type: "btree", field: [{name: "id_representante", op: "asc"}]}
    {type: "btree", field: [{name: "id_usuario", op: "asc"}]}
    {type: "btree", field: [{name: "admin_usuario", op: "asc"}]}
  ]
}
