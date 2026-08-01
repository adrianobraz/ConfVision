// Atualizar no MediaMTX
query "vis_mediamtx_node/{vis_mediamtx_node_id}" verb=PUT {
  api_group = "confVision"

  input {
    int vis_mediamtx_node_id? filters=min:1
    dblink {
      table = "vis_mediamtx_node"
    }
  }

  stack {
    db.get vis_mediamtx_node {
      field_name = "id"
      field_value = $input.vis_mediamtx_node_id
    } as $existe

    precondition ($existe != null) {
      error_type = "notfound"
      error = "No MediaMTX nao encontrado"
    }

    db.patch vis_mediamtx_node {
      field_name = "id"
      field_value = $input.vis_mediamtx_node_id
      data = {
        nome         : $input.nome
        rtmp_public  : $input.rtmp_public
        hls_public   : $input.hls_public
        rtsp_internal: $input.rtsp_internal
        max_cameras  : $input.max_cameras
        ordem        : $input.ordem
        status       : $input.status
        observacao   : $input.observacao
      }
    } as $model

    function.run fn_vis_mediamtx_node_sync_status {
      input = {vis_mediamtx_node_id: $input.vis_mediamtx_node_id}
    } as $sync
  }

  response = $model
}
