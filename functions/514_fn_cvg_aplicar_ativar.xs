// Ativar: armar (se plano armado) + despausar cameras
function fn_cvg_aplicar_ativar {
  input {
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
  }

  stack {
    function.run fn_cvg_listar_cameras_slot {
      input = {
        id_cliente    : $input.id_cliente
        id_dispositivo: $input.id_dispositivo
      }
    } as $cameras

    var $cameras_despausar {
      value = []
    }

    var $dispositivos {
      value = []
    }

    var $erros {
      value = []
    }

    var $disp_feitos {
      value = ","
    }

    foreach ($cameras) {
      each as $cam {
        conditional {
          if ($cam.plano_tipo == "armado" && ($cam.id_dispositivo|is_empty) == false) {
            var $chave_disp {
              value = "," ~ $cam.id_dispositivo ~ ","
            }

            conditional {
              if (($disp_feitos|contains:$chave_disp) == false) {
                var.update $disp_feitos {
                  value = $disp_feitos ~ $cam.id_dispositivo ~ ","
                }

                try_catch {
                  try {
                    function.run fn_cvg_dispositivo_armar {
                      input = {
                        id_dispositivo: $cam.id_dispositivo
                        acao          : 1
                      }
                    } as $d

                    array.push $dispositivos {
                      value = $d
                    }
                  }

                  catch {
                    array.push $erros {
                      value = "armar:" ~ $cam.id_dispositivo
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    foreach ($cameras) {
      each as $cam {
        try_catch {
          try {
            function.run fn_cvg_camera_pausar {
              input = {
                vis_camera_id: $cam.id
                pausado      : false
              }
            } as $p

            array.push $cameras_despausar {
              value = $cam.id
            }
          }

          catch {
            array.push $erros {
              value = "despausar:" ~ $cam.id
            }
          }
        }
      }
    }
  }

  response = {
    acao             : "ativar"
    dispositivos     : $dispositivos
    cameras_despausar: $cameras_despausar
    erros            : $erros
  }
}
