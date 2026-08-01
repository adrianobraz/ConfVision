table fp_whitelabel {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_franqueado? filters=trim
    json tema_json?
    text logo_data? filters=trim
    timestamp? updated_at?
    text subdominio? filters=trim|lower
    text dominio? filters=trim|lower
    text fqdn? filters=trim|lower
    text dominio_status? filters=trim|lower
    text dominio_erro? filters=trim
    timestamp? provisionado_em?
    text subdominio_cv? filters=trim|lower
    text fqdn_cv? filters=trim|lower
    text dominio_status_cv? filters=trim|lower
    text dominio_erro_cv? filters=trim
    timestamp? provisionado_em_cv?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "fqdn", op: "asc"}]}
    {type: "btree", field: [{name: "fqdn_cv", op: "asc"}]}
  ]
}