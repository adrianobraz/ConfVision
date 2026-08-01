// Salva grade_ativa e slots do cliente (substitui slots enviados)
query cvg_grade_by_cliente verb=PUT {
  api_group = "confVisionGrade"

  input {
    text id_cliente? filters=trim
    text id_franqueado? filters=trim
    text id_dispositivo_escopo? filters=trim
    bool grade_ativa?
    object[] slots? {
      schema {
        int id?
        text id_dispositivo? filters=trim
        int dia_semana? filters=min:1|max:7
        text hora? filters=trim
        text acao? filters=trim
        bool ativo?
      }
    }
  }

  stack {
    precondition (($input.id_cliente|is_empty) == false) {
      error = "id_cliente obrigatorio"
    }

    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    var $grade_ativa {
      value = $input.grade_ativa|first_notempty:false
    }

    var $escopo {
      value = $input.id_dispositivo_escopo|first_notempty:""
    }

    db.query vis_cliente_grade_config {
      where = $db.vis_cliente_grade_config.id_cliente == $input.id_cliente
      return = {type: "single"}
    } as $cfg_exist

    conditional {
      if ($cfg_exist == null) {
        db.add vis_cliente_grade_config {
          data = {
            id_franqueado: $input.id_franqueado
            id_cliente   : $input.id_cliente
            grade_ativa  : false
          }
        } as $cfg_nova
      }
    }

    function.run fn_cvg_escopo_salvar {
      input = {
        id_franqueado : $input.id_franqueado
        id_cliente    : $input.id_cliente
        id_dispositivo: $escopo
        grade_ativa   : $grade_ativa
      }
    } as $escopo_salvo

    db.query vis_cliente_grade_slot {
      where = $db.vis_cliente_grade_slot.id_cliente == $input.id_cliente
      return = {type: "list"}
    } as $slots_antigos

    foreach ($slots_antigos) {
      each as $old {
        var $disp_old {
          value = $old.id_dispositivo|first_notempty:""
        }

        var $apagar {
          value = $disp_old == $escopo
        }

        conditional {
          if ($apagar) {
            db.del vis_cliente_grade_slot {
              field_name = "id"
              field_value = $old.id
            }
          }
        }
      }
    }

    var $slots_salvos {
      value = []
    }

    foreach ($input.slots|first_notempty:[]) {
      each as $sl {
        var $hora_norm {
          value = $sl.hora|first_notempty:""
        }

        var $acao_norm {
          value = $sl.acao|first_notempty:""
        }

        precondition (($hora_norm|is_empty) == false) {
          error = "hora obrigatoria em cada slot (HH:MM)"
        }

        precondition ($acao_norm == "ativar" || $acao_norm == "desativar") {
          error = "acao deve ser ativar ou desativar"
        }

        db.add vis_cliente_grade_slot {
          data = {
            id_franqueado : $input.id_franqueado
            id_cliente    : $input.id_cliente
            id_dispositivo: $escopo
            dia_semana    : $sl.dia_semana
            hora          : $hora_norm
            acao          : $acao_norm
            ativo         : $sl.ativo|first_notempty:true
          }
        } as $slot_novo

        array.push $slots_salvos {
          value = $slot_novo
        }
      }
    }
  }

  response = {
    success    : true
    id_cliente : $input.id_cliente
    grade_ativa: $grade_ativa
    slots      : $slots_salvos
  }
}
