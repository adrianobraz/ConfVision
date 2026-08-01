// Escolhe no MediaMTX com vaga (menor ocupacao, ordem crescente)
function fn_vis_mediamtx_pick_node {
  input {
  }

  stack {
    function.run fn_vis_mediamtx_node_ensure_default {
    } as $default_ok

    db.query vis_mediamtx_node {
      where = $db.vis_mediamtx_node.status == "ativo"
      sort = {vis_mediamtx_node.ordem: "asc"}
      return = {type: "list"}
    } as $nodes

    precondition (($nodes|count) > 0) {
      error = "Nenhum no MediaMTX ativo — cadastre vis_mediamtx_node ou libere nos cheios"
    }

    var $escolhido {
      value = null
    }

    var $menor_ocupacao {
      value = 999999
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

        conditional {
          if ($total < $max && $total < $menor_ocupacao) {
            var.update $menor_ocupacao {
              value = $total
            }

            var.update $escolhido {
              value = $node|set:"cameras_atribuidas":$total
            }
          }
        }
      }
    }

    precondition ($escolhido != null) {
      error = "Todos os nos MediaMTX estao cheios — cadastre um novo servidor ou aumente max_cameras"
    }
  }

  response = $escolhido
}
