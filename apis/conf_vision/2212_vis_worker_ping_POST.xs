// Add vis_worker record
query vis_worker_ping verb=POST {
  api_group = "confVision"

  input {
    dblink {
      table = "vis_worker"
    }
  }

  stack {
    db.add vis_worker {
      enforce_hidden_fields = false
      data = {
        created_at          : "now"
        worker_id           : $input.worker_id
        hostname            : $input.hostname
        versao              : $input.versao
        cameras_ativas      : $input.cameras_ativas
        shard_index         : $input.shard_index
        shard_total         : $input.shard_total
        max_cameras         : $input.max_cameras
        yolo_device         : $input.yolo_device
        queue_backend       : $input.queue_backend
        vis_mediamtx_node_id: $input.vis_mediamtx_node_id
        ultimo_ping_em      : $input.ultimo_ping_em
        ativo               : $input.ativo
      }
    } as $model
  }

  response = $model
}
