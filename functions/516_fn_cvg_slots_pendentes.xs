// Slots ativos para dia/hora que ainda nao foram executados na data_ref
function fn_cvg_slots_pendentes {
  input {
    int dia_semana? filters=min:1|max:7
    text hora? filters=trim
    text data_ref? filters=trim
  }

  stack {
    precondition ($input.dia_semana != null) {
      error = "dia_semana obrigatorio (1=seg .. 7=dom)"
    }

    precondition (($input.hora|is_empty) == false) {
      error = "hora obrigatoria (HH:MM)"
    }

    precondition (($input.data_ref|is_empty) == false) {
      error = "data_ref obrigatorio"
    }

    db.query vis_cliente_grade_slot {
      where = $db.vis_cliente_grade_slot.dia_semana == $input.dia_semana && $db.vis_cliente_grade_slot.hora == $input.hora && $db.vis_cliente_grade_slot.ativo == true
      sort = {vis_cliente_grade_slot.id: "asc"}
      return = {type: "list"}
    } as $slots_raw

    var $pendentes {
      value = []
    }

    foreach ($slots_raw) {
      each as $slot {
        var $id_disp_slot {
          value = $slot.id_dispositivo|first_notempty:""
        }

        function.run fn_cvg_escopo_grade_ativa {
          input = {
            id_cliente    : $slot.id_cliente
            id_dispositivo: $id_disp_slot
          }
        } as $grade_ativa_escopo

        conditional {
          if ($grade_ativa_escopo) {
            db.query vis_cliente_grade_exec {
              where = $db.vis_cliente_grade_exec.slot_id == $slot.id && $db.vis_cliente_grade_exec.data_ref == $input.data_ref && $db.vis_cliente_grade_exec.hora_ref == $input.hora
              return = {type: "single"}
            } as $exec

            conditional {
              if ($exec == null) {
                array.push $pendentes {
                  value = $slot
                }
              }
            }
          }
        }
      }
    }
  }

  response = $pendentes
}
