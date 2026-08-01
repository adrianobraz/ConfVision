// Nó MediaMTX (VPS) — ingestão RTMP / leitura RTSP por cluster
table vis_mediamtx_node {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }

    text nome? filters=trim
    text rtmp_public? filters=trim
    text hls_public? filters=trim
    text rtsp_internal? filters=trim
    int max_cameras?=200 filters=min:1
    int ordem?=1 filters=min:1
    text status? filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "status", op: "asc"}]}
    {type: "btree", field: [{name: "ordem", op: "asc"}]}
  ]
}
