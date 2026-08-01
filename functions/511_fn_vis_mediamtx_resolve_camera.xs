// Resolve no + URLs RTMP/HLS/RTSP para uma camera (atribui no se ausente)
function fn_vis_mediamtx_resolve_camera {
  input {
    int vis_camera_id? filters=min:1
    bool? atribuir_se_ausente?=true
  }

  stack {
    db.get vis_camera {
      field_name = "id"
      field_value = $input.vis_camera_id
    } as $cam

    precondition ($cam != null) {
      error_type = "notfound"
      error = "Camera nao encontrada"
    }

    var $node_id {
      value = $cam.vis_mediamtx_node_id
    }

    conditional {
      if ($node_id == null && ($input.atribuir_se_ausente|first_notempty:true)) {
        function.run fn_vis_mediamtx_pick_node {
        } as $picked

        db.patch vis_camera {
          field_name = "id"
          field_value = $cam.id
          data = {vis_mediamtx_node_id: $picked.id}
        } as $cam_upd

        var.update $node_id {
          value = $picked.id
        }

        function.run fn_vis_mediamtx_node_sync_status {
          input = {vis_mediamtx_node_id: $picked.id}
        } as $sync1
      }
    }

    precondition ($node_id != null) {
      error = "Camera sem no MediaMTX — nenhum servidor disponivel"
    }

    db.get vis_mediamtx_node {
      field_name = "id"
      field_value = $node_id
    } as $node

    precondition ($node != null) {
      error = "No MediaMTX da camera nao encontrado"
    }

    db.query vis_camera {
      where = $db.vis_camera.vis_mediamtx_node_id == $node.id
      return = {type: "count"}
    } as $total

    var $rtmp_public {
      value = $node.rtmp_public|trim
    }

    var $hls_public {
      value = $node.hls_public|trim
    }

    var $rtsp_internal {
      value = $node.rtsp_internal|trim
    }

    conditional {
      if (($rtmp_public|is_empty) == true) {
        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_rtmp_public"
        } as $cfg_rtmp

        conditional {
          if ($cfg_rtmp != null && ($cfg_rtmp.valor|is_empty) == false) {
            var.update $rtmp_public {
              value = $cfg_rtmp.valor|trim
            }
          }
        }
      }
    }

    conditional {
      if (($hls_public|is_empty) == true) {
        db.get fp_config_financeiro {
          field_name = "chave"
          field_value = "confvision_mediamtx_default_hls_public"
        } as $cfg_hls

        conditional {
          if ($cfg_hls != null && ($cfg_hls.valor|is_empty) == false) {
            var.update $hls_public {
              value = $cfg_hls.valor|trim
            }
          }
        }
      }
    }

    conditional {
      if (($rtsp_internal|is_empty) == true) {
        var.update $rtsp_internal {
          value = "rtsp://127.0.0.1:8554"
        }
      }
    }
  }

  response = {
    vis_camera_id      : $cam.id
    vis_mediamtx_node_id: $node.id
    nome               : $node.nome
    rtmp_public        : $rtmp_public
    hls_public         : $hls_public
    rtsp_internal      : $rtsp_internal
    max_cameras        : $node.max_cameras|first_notempty:200
    cameras_atribuidas : $total
    status             : $node.status
  }
}
