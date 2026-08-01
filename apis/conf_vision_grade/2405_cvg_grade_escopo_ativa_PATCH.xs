// Ativa/desativa grade por escopo cliente+dispositivo (sem alterar slots)
query cvg_grade_escopo_ativa verb=PATCH {
  api_group = "confVisionGrade"

  input {
    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    bool grade_ativa?
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    precondition (($input.id_cliente|is_empty) == false) {
      error = "id_cliente obrigatorio"
    }

    var $disp {
      value = $input.id_dispositivo|first_notempty:""
    }

    var $grade_ativa {
      value = $input.grade_ativa|first_notempty:false
    }

    db.query vis_cliente_grade_slot {
      where = $db.vis_cliente_grade_slot.id_cliente == $input.id_cliente
      return = {type: "list"}
    } as $slots_raw

    var $qtd_escopo {
      value = 0
    }

    foreach ($slots_raw) {
      each as $sl {
        var $sd {
          value = $sl.id_dispositivo|first_notempty:""
        }

        conditional {
          if ($sd == $disp) {
            var.update $qtd_escopo {
              value = $qtd_escopo + 1
            }
          }
        }
      }
    }

    precondition ($qtd_escopo > 0) {
      error = "Nenhum horario salvo para este escopo"
    }

    function.run fn_cvg_escopo_salvar {
      input = {
        id_franqueado : $input.id_franqueado
        id_cliente    : $input.id_cliente
        id_dispositivo: $disp
        grade_ativa   : $grade_ativa
      }
    } as $salvo
  }

  response = {
    success     : true
    id_cliente  : $input.id_cliente
    id_dispositivo: $disp
    grade_ativa : $grade_ativa
  }
}
