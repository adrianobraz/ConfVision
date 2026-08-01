table vis_licenca {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_franqueado? filters=trim
    text plano? filters=trim
    text unidade?=camera filters=trim
    decimal valor?
    timestamp? pago_em?
    timestamp? valido_ate?
    text status?=disponivel filters=trim
    text id_dispositivo? filters=trim
    text id_fatura? filters=trim
    text id_pagamento? filters=trim
    text observacao? filters=trim
    int? vis_camera_id? {
      table = "vis_camera"
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "id_dispositivo", op: "asc"}]}
    {type: "btree", field: [{name: "valido_ate", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "status", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "unidade", op: "asc"}
        {name: "status", op: "asc"}
      ]
    }
  ]
}