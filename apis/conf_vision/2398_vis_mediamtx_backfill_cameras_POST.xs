// Atribui no MediaMTX a cameras legadas sem vis_mediamtx_node_id
query vis_mediamtx_backfill_cameras verb=POST {
  api_group = "confVision"

  input {
    int vis_mediamtx_node_id? filters=min:1
    int limite?=500 filters=min:1|max:2000
  }

  stack {
    var $limite {
      value = $input.limite|first_notempty:500
    }

    db.query vis_camera {
      where = $db.vis_camera.vis_mediamtx_node_id == null
      sort = {vis_camera.id: "asc"}
      return = {type: "list", paging: {page: 1, per_page: $limite, metadata: false}}
    } as $cameras_raw

    var $cameras {
      value = $cameras_raw|get:"items":[]
    }

    var $atribuidas {
      value = 0
    }

    foreach ($cameras) {
      each as $cam {
        conditional {
          if ($input.vis_mediamtx_node_id != null) {
            db.patch vis_camera {
              field_name = "id"
              field_value = $cam.id
              data = {vis_mediamtx_node_id: $input.vis_mediamtx_node_id}
            } as $upd

            var.update $atribuidas {
              value = $atribuidas + 1
            }
          }

          else {
            function.run fn_vis_mediamtx_resolve_camera {
              input = {vis_camera_id: $cam.id, atribuir_se_ausente: true}
            } as $res

            var.update $atribuidas {
              value = $atribuidas + 1
            }
          }
        }
      }
    }

    function.run fn_vis_mediamtx_node_sync_status {
      input = {vis_mediamtx_node_id: $input.vis_mediamtx_node_id}
    } as $sync
  }

  response = {atribuidas: $atribuidas, processadas: $cameras|count}
}
