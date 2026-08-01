query vis_camera_query_ativas verb=GET {
  api_group = "confVision"

  input {
    text worker_id? filters=trim
    int vis_mediamtx_node_id? filters=min:1
  }

  stack {
    db.query vis_camera {
      where = $db.vis_camera.ativo == true && $db.vis_camera.deteccao_humano == true && ($db.vis_camera.analitico_pausado != true) && $db.vis_camera.worker_id ==? $input.worker_id && $db.vis_camera.vis_mediamtx_node_id ==? $input.vis_mediamtx_node_id
      return = {type: "list"}
    } as $cameras

    var $result {
      value = []
    }

    foreach ($cameras) {
      each as $cam {
        var $rtsp_base {
          value = ""
        }

        conditional {
          if ($cam.vis_mediamtx_node_id != null) {
            db.get vis_mediamtx_node {
              field_name = "id"
              field_value = $cam.vis_mediamtx_node_id
            } as $node

            conditional {
              if ($node != null) {
                var.update $rtsp_base {
                  value = $node.rtsp_internal|trim
                }
              }
            }
          }
        }

        array.push $result {
          value = $cam|set:"mediamtx_rtsp_base":$rtsp_base
        }
      }
    }
  }

  response = {dados: $result}
}
