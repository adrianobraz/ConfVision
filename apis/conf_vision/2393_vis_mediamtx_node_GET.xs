// Lista nos MediaMTX com contagem de cameras atribuidas
query vis_mediamtx_node verb=GET {
  api_group = "confVision"

  input {
    text status? filters=trim
  }

  stack {
    db.query vis_mediamtx_node {
      where = $db.vis_mediamtx_node.status ==? $input.status
      sort = {vis_mediamtx_node.ordem: "asc"}
      return = {type: "list"}
    } as $nodes

    var $resultado {
      value = []
    }

    foreach ($nodes) {
      each as $node {
        db.query vis_camera {
          where = $db.vis_camera.vis_mediamtx_node_id == $node.id
          return = {type: "count"}
        } as $total

        array.push $resultado {
          value = {
            id                : $node.id
            created_at        : $node.created_at
            nome              : $node.nome
            rtmp_public       : $node.rtmp_public
            hls_public        : $node.hls_public
            rtsp_internal     : $node.rtsp_internal
            max_cameras       : $node.max_cameras|first_notempty:200
            ordem             : $node.ordem|first_notempty:1
            status            : $node.status
            observacao        : $node.observacao
            cameras_atribuidas: $total
            vagas_restantes   : ($node.max_cameras|first_notempty:200) - $total
          }
        }
      }
    }
  }

  response = {dados: $resultado}
}
