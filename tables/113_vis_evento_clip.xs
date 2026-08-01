table vis_evento_clip {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    int vis_evento_id? {
      table = "vis_evento"
    }
  
    int seq?
    text video_url? filters=trim
    int duracao_seg?
    text snapshot_url? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "vis_evento_id", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "vis_evento_id", op: "asc"}, {name: "seq", op: "asc"}]
    }
  ]
}