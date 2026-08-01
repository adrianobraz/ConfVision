table vis_gravacao_storage {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_franqueado? filters=trim
    text s3_endpoint? filters=trim
    text s3_bucket? filters=trim
    text s3_tenant_id? filters=trim
    text s3_access_key? filters=trim
    text s3_secret_key? filters=trim
    int segmento_minutos?
    text status? filters=trim
    timestamp? provisionado_em?
    timestamp? cancelado_em?
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "status", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "status", op: "asc"}]}
  ]
}