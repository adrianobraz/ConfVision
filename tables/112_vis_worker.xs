table vis_worker {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text worker_id? filters=trim
    text hostname? filters=trim
    text versao? filters=trim
    int cameras_ativas?
    timestamp? ultimo_ping_em?
    bool ativo?
    int shard_index?
    int shard_total?
    int max_cameras?
    text yolo_device? filters=trim
    text queue_backend? filters=trim
    int vis_mediamtx_node_id? {
      table = "vis_mediamtx_node"
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
  ]
}