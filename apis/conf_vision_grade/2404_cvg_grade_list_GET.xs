// Lista grades por franqueado (cliente+dispositivo com slots salvos)
query cvg_grade_list verb=GET {
  api_group = "confVisionGrade"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    db.query vis_cliente_grade_slot {
      where = $db.vis_cliente_grade_slot.id_franqueado == $input.id_franqueado
      sort = {
        vis_cliente_grade_slot.id_cliente   : "asc"
        vis_cliente_grade_slot.id_dispositivo: "asc"
      }
      return = {type: "list"}
    } as $slots_raw

    var $itens {
      value = []
    }

    foreach ($slots_raw) {
      each as $s {
        var $disp {
          value = $s.id_dispositivo|first_notempty:""
        }

        var $ja_tem {
          value = false
        }

        foreach ($itens) {
          each as $item {
            var $item_disp {
              value = $item.id_dispositivo|first_notempty:""
            }

            conditional {
              if ($item.id_cliente == $s.id_cliente && $item_disp == $disp) {
                var.update $ja_tem {
                  value = true
                }
              }
            }
          }
        }

        conditional {
          if ($ja_tem == false) {
            function.run fn_cvg_escopo_grade_ativa {
              input = {
                id_cliente    : $s.id_cliente
                id_dispositivo: $disp
              }
            } as $grade_ativa

            array.push $itens {
              value = {
                id_cliente    : $s.id_cliente
                id_dispositivo: $disp
                grade_ativa   : $grade_ativa
              }
            }
          }
        }
      }
    }
  }

  response = {itens: $itens}
}
