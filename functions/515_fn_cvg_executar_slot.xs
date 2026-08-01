// Executa um slot com idempotencia (vis_cliente_grade_exec)
function fn_cvg_executar_slot {
  input {
    int slot_id? filters=min:1
    text data_ref? filters=trim
    text hora_ref? filters=trim
  }

  stack {
    precondition (($input.data_ref|is_empty) == false) {
      error = "data_ref obrigatorio"
    }

    precondition (($input.hora_ref|is_empty) == false) {
      error = "hora_ref obrigatorio"
    }

    var $out {
      value = {
        slot_id : $input.slot_id
        ignorado: false
      }
    }

    db.query vis_cliente_grade_exec {
      where = $db.vis_cliente_grade_exec.slot_id == $input.slot_id && $db.vis_cliente_grade_exec.data_ref == $input.data_ref && $db.vis_cliente_grade_exec.hora_ref == $input.hora_ref
      return = {type: "single"}
    } as $ja_exec

    conditional {
      if ($ja_exec != null) {
        var.update $out {
          value = {
            slot_id : $input.slot_id
            ignorado: true
            motivo  : "ja_executado"
            exec_id : $ja_exec.id
          }
        }
      }
    }

    conditional {
      if ($ja_exec == null) {
        db.get vis_cliente_grade_slot {
          field_name = "id"
          field_value = $input.slot_id
        } as $slot

        precondition ($slot != null) {
          error = "Slot nao encontrado"
        }

        precondition ($slot.ativo) {
          error = "Slot inativo"
        }

        var $id_disp_slot {
          value = $slot.id_dispositivo|first_notempty:""
        }

        function.run fn_cvg_escopo_grade_ativa {
          input = {
            id_cliente    : $slot.id_cliente
            id_dispositivo: $id_disp_slot
          }
        } as $grade_ativa_escopo

        precondition ($grade_ativa_escopo) {
          error = "Grade desativada para este dispositivo"
        }

        var $detalhe {
          value = {}
        }

        var $resultado {
          value = "ok"
        }

        conditional {
          if ($slot.acao == "desativar") {
            function.run fn_cvg_aplicar_desativar {
              input = {
                id_cliente    : $slot.id_cliente
                id_dispositivo: $id_disp_slot
              }
            } as $detalhe
          }

          elseif ($slot.acao == "ativar") {
            function.run fn_cvg_aplicar_ativar {
              input = {
                id_cliente    : $slot.id_cliente
                id_dispositivo: $id_disp_slot
              }
            } as $detalhe
          }

          else {
            var.update $resultado {
              value = "erro"
            }

            var.update $detalhe {
              value = {erro: "acao invalida"}
            }
          }
        }

        conditional {
          if ($detalhe.erros != null && ($detalhe.erros|count) > 0) {
            var.update $resultado {
              value = "parcial"
            }
          }
        }

        db.add vis_cliente_grade_exec {
          data = {
            slot_id     : $input.slot_id
            data_ref    : $input.data_ref
            hora_ref    : $input.hora_ref
            executado_em: "now"
            resultado   : $resultado
            detalhe     : $detalhe
          }
        } as $exec_row

        var.update $out {
          value = {
            slot_id  : $input.slot_id
            ignorado : false
            resultado: $resultado
            detalhe  : $detalhe
            exec_id  : $exec_row.id
          }
        }
      }
    }
  }

  response = $out
}
