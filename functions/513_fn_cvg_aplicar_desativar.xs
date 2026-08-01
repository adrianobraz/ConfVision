// Desativar: pausar cameras + desarmar (se plano armado)
function fn_cvg_aplicar_desativar {
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

    var $cameras_pausar {
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
        try_catch {
          try {
            function.run fn_cvg_camera_pausar {
              input = {
                vis_camera_id: $cam.id
                pausado      : true
              }
            } as $p

            array.push $cameras_pausar {
              value = $cam.id
            }

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
                            acao          : 0
                          }
                        } as $d

                        array.push $dispositivos {
                          value = $d
                        }
                      }

                      catch {
                        array.push $erros {
                          value = "desarmar:" ~ $cam.id_dispositivo
                        }
                      }
                    }
                  }
                }
              }
            }
          }

          catch {
            array.push $erros {
              value = "pausar:" ~ $cam.id
            }
          }
        }
      }
    }
  }

  response = {
    acao          : "desativar"
    cameras_pausar: $cameras_pausar
    dispositivos  : $dispositivos
    erros         : $erros
  }
}
