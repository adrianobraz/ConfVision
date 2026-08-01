// Atualiza status ativo/cheio conforme contagem de cameras atribuidas
function fn_vis_mediamtx_node_sync_status {
  input {
    int vis_mediamtx_node_id? filters=min:1
  }

  stack {
    var $nodes {
      value = []
    }

    conditional {
      if ($input.vis_mediamtx_node_id != null) {
        db.get vis_mediamtx_node {
          field_name = "id"
          field_value = $input.vis_mediamtx_node_id
        } as $uno

        precondition ($uno != null) {
          error = "No MediaMTX nao encontrado"
        }

        var.update $nodes {
          value = [$uno]
        }
      }

      else {
        db.query vis_mediamtx_node {
          sort = {vis_mediamtx_node.ordem: "asc"}
          return = {type: "list"}
        } as $lista

        var.update $nodes {
          value = $lista
        }
      }
    }

    var $atualizados {
      value = 0
    }

    foreach ($nodes) {
      each as $node {
        db.query vis_camera {
          where = $db.vis_camera.vis_mediamtx_node_id == $node.id
          return = {type: "count"}
        } as $total

        var $max {
          value = $node.max_cameras|first_notempty:200
        }

        var $novo_status {
          value = $node.status|first_notempty:"ativo"
        }

        conditional {
          if ($total >= $max) {
            var.update $novo_status {
              value = "cheio"
            }
          }

          elseif ($node.status == "cheio") {
            var.update $novo_status {
              value = "ativo"
            }
          }
        }

        conditional {
          if ($novo_status != $node.status) {
            db.patch vis_mediamtx_node {
              field_name = "id"
              field_value = $node.id
              data = {status: $novo_status}
            } as $pat

            var.update $atualizados {
              value = $atualizados + 1
            }
          }
        }
      }
    }
  }

  response = {atualizados: $atualizados}
}
