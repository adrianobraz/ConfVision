// Orquestra execucao de todos os slots pendentes do minuto
function fn_cvg_worker_tick {
  input {
    int dia_semana? filters=min:1|max:7
    text hora? filters=trim
    text data_ref? filters=trim
  }

  stack {
    function.run fn_cvg_slots_pendentes {
      input = {
        dia_semana: $input.dia_semana
        hora      : $input.hora
        data_ref  : $input.data_ref
      }
    } as $pendentes

    var $executados {
      value = 0
    }

    var $ignorados {
      value = 0
    }

    var $erros {
      value = 0
    }

    var $detalhes {
      value = []
    }

    foreach ($pendentes) {
      each as $slot {
        try_catch {
          try {
            function.run fn_cvg_executar_slot {
              input = {
                slot_id : $slot.id
                data_ref: $input.data_ref
                hora_ref: $input.hora
              }
            } as $res

            conditional {
              if ($res.ignorado) {
                var.update $ignorados {
                  value = $ignorados + 1
                }
              }

              else {
                var.update $executados {
                  value = $executados + 1
                }

                conditional {
                  if ($res.resultado == "erro" || $res.resultado == "parcial") {
                    var.update $erros {
                      value = $erros + 1
                    }
                  }
                }
              }
            }

            array.push $detalhes {
              value = $res|set:"slot_id":$slot.id
            }
          }

          catch {
            var.update $erros {
              value = $erros + 1
            }

            array.push $detalhes {
              value = {
                slot_id: $slot.id
                erro   : "execucao_falhou"
              }
            }
          }
        }
      }
    }
  }

  response = {
    data_ref  : $input.data_ref
    hora      : $input.hora
    dia_semana: $input.dia_semana
    total     : $pendentes|count
    executados: $executados
    ignorados : $ignorados
    erros     : $erros
    detalhes  : $detalhes
  }
}
