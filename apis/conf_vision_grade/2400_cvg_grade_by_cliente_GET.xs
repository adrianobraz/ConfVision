// Retorna config e slots da grade por cliente (filtro opcional por dispositivo)
query cvg_grade_by_cliente verb=GET {
  api_group = "confVisionGrade"

  input {
    text id_cliente? filters=trim
    text id_franqueado? filters=trim
    text id_dispositivo? filters=trim
  }

  stack {
    precondition (($input.id_cliente|is_empty) == false) {
      error = "id_cliente obrigatorio"
    }

    db.query vis_cliente_grade_config {
      where = $db.vis_cliente_grade_config.id_cliente == $input.id_cliente
      return = {type: "single"}
    } as $cfg

    var $filtro_disp {
      value = $input.id_dispositivo|first_notempty:""
    }

    function.run fn_cvg_escopo_grade_ativa {
      input = {
        id_cliente    : $input.id_cliente
        id_dispositivo: $filtro_disp
      }
    } as $grade_ativa_escopo

    var $grade_ativa {
      value = $grade_ativa_escopo|first_notempty:false
    }

    db.query vis_cliente_grade_slot {
      where = $db.vis_cliente_grade_slot.id_cliente == $input.id_cliente
      sort = {
        vis_cliente_grade_slot.dia_semana: "asc"
        vis_cliente_grade_slot.hora      : "asc"
      }
      return = {type: "list"}
    } as $slots_raw

    var $slots {
      value = []
    }

    foreach ($slots_raw) {
      each as $s {
        var $incluir {
          value = true
        }

        conditional {
          if (($filtro_disp|is_empty) == false) {
            var.update $incluir {
              value = ($s.id_dispositivo|first_notempty:"") == $filtro_disp
            }
          }

          else {
            var.update $incluir {
              value = ($s.id_dispositivo|is_empty)
            }
          }
        }

        conditional {
          if ($incluir) {
            array.push $slots {
              value = $s
            }
          }
        }
      }
    }

    var $config_id {
      value = 0
    }

    conditional {
      if ($cfg != null) {
        var.update $config_id {
          value = $cfg.id|first_notempty:0
        }
      }
    }
  }

  response = {
    id_cliente : $input.id_cliente
    grade_ativa: $grade_ativa
    config_id  : $config_id
    slots      : $slots
  }
}
