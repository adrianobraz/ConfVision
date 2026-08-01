// Lista cameras alvo de um slot (analitico ativo, filtro cliente/dispositivo)
function fn_cvg_listar_cameras_slot {
  input {
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
  }

  stack {
    precondition (($input.id_cliente|is_empty) == false) {
      error = "id_cliente obrigatorio"
    }

    db.query vis_camera {
      where = $db.vis_camera.id_cliente == $input.id_cliente && $db.vis_camera.ativo == true
      sort = {vis_camera.id: "asc"}
      return = {type: "list"}
    } as $raw

    var $cameras {
      value = []
    }

    foreach ($raw) {
      each as $cam {
        var $incluir {
          value = true
        }

        conditional {
          if (($input.id_dispositivo|is_empty) == false) {
            conditional {
              if ($cam.id_dispositivo != $input.id_dispositivo) {
                var.update $incluir {
                  value = false
                }
              }
            }
          }
        }

        conditional {
          if ($incluir) {
            function.run fn_cvg_plano_tipo {
              input = {plano: $cam.plano}
            } as $pt

            conditional {
              if ($pt.tipo != "outro") {
                array.push $cameras {
                  value = $cam
                    |set:"plano_tipo":$pt.tipo
                }
              }
            }
          }
        }
      }
    }
  }

  response = $cameras
}
