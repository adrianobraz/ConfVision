// Cameras com gravacao ativa — sync worker DVR / motion / timelapse (independente de ativo analitico)
query vis_camera_query_gravacao_ativas verb=GET {
  api_group = "confVision"

  input {
    text worker_id? filters=trim
    int vis_mediamtx_node_id? filters=min:1
  }

  stack {
    db.query vis_camera {
      where = $db.vis_camera.gravacao_status == "ativa" && $db.vis_camera.vis_licenca_gravacao_id != null && $db.vis_camera.worker_id ==? $input.worker_id && $db.vis_camera.vis_mediamtx_node_id ==? $input.vis_mediamtx_node_id
      return = {type: "list"}
    } as $cameras
  
    var $result {
      value = []
    }
  
    foreach ($cameras) {
      each as $cam {
        db.get vis_licenca {
          field_name = "id"
          field_value = $cam.vis_licenca_gravacao_id
        } as $licenca
      
        conditional {
          if ($licenca == null) {
            continue
          }
        }
      
        function.run fn_vis_plano_flags {
          input = {plano: $licenca.plano}
        } as $flags
      
        conditional {
          if ($flags.grava_continua == false && $flags.grava_movimento == false && $flags.grava_timelapse == false) {
            continue
          }
        }
      
        db.query vis_gravacao_storage {
          where = $db.vis_gravacao_storage.id_franqueado == $cam.id_franqueado && $db.vis_gravacao_storage.status == "ativo"
          return = {type: "single"}
        } as $storage
      
        var $rtsp_base {
          value = ""
        }

        conditional {
          if ($cam.vis_mediamtx_node_id != null) {
            db.get vis_mediamtx_node {
              field_name = "id"
              field_value = $cam.vis_mediamtx_node_id
            } as $mtx_node

            conditional {
              if ($mtx_node != null) {
                var.update $rtsp_base {
                  value = $mtx_node.rtsp_internal|trim
                }
              }
            }
          }
        }

        conditional {
          if ($storage != null) {
            var $modo_gravacao {
              value = "continua"
            }
          
            conditional {
              if ($flags.grava_movimento) {
                var.update $modo_gravacao {
                  value = "movimento"
                }
              }
            
              elseif ($flags.grava_timelapse) {
                var.update $modo_gravacao {
                  value = "timelapse"
                }
              }
            }
          
            array.push $result {
              value = {
                id                     : $cam.id
                nome                   : $cam.nome
                id_franqueado          : $cam.id_franqueado
                id_cliente             : $cam.id_cliente
                rtsp_url_sec           : $cam.rtsp_url_sec
                mediamtx_rtsp_base     : $rtsp_base
                vis_mediamtx_node_id   : $cam.vis_mediamtx_node_id
                retencao_dias          : $cam.retencao_dias
                segmento_minutos       : $storage.segmento_minutos
                vis_gravacao_storage_id: $storage.id
                modo_gravacao          : $modo_gravacao
                grava_continua         : $flags.grava_continua
                grava_movimento        : $flags.grava_movimento
                grava_timelapse        : $flags.grava_timelapse
                plano_gravacao         : $licenca.plano
                gravacao_flush_pedido  : $cam.gravacao_flush_pedido
              }
            }
          }
        }
      }
    }
  }

  response = {dados: $result}
}