table vis_gravacao_segmento {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int vis_camera_id? {
      table = "vis_camera"
    }
  
    int vis_gravacao_storage_id? {
      table = "vis_gravacao_storage"
    }
  
    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    timestamp? inicio_em?
    timestamp? fim_em?
    int duracao_seg?
    text s3_key? filters=trim
    text s3_url? filters=trim
    int tamanho_bytes?
    text status? filters=trim
    text erro_msg? filters=trim
    timestamp? uploaded_em?
    timestamp? expira_em?
    text tipo? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "vis_camera_id", op: "asc"}]}
    {type: "btree", field: [{name: "inicio_em", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "vis_camera_id", op: "asc"}
        {name: "inicio_em", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "inicio_em", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "status", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "status", op: "asc"}, {name: "created_at", op: "asc"}]
    }
    {type: "btree", field: [{name: "expira_em", op: "asc"}]}
  ]
}