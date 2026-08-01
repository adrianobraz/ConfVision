// Cria ou atualiza grade_ativa por escopo cliente+dispositivo
function fn_cvg_escopo_salvar {
  input {
    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    bool grade_ativa?
  }

  stack {
    var $disp {
      value = $input.id_dispositivo|first_notempty:""
    }

    var $ativa {
      value = $input.grade_ativa|first_notempty:false
    }

    db.query vis_cliente_grade_escopo {
      where = $db.vis_cliente_grade_escopo.id_cliente == $input.id_cliente
      return = {type: "list"}
    } as $escopos_raw

    var $exist {
      value = null
    }

    foreach ($escopos_raw) {
      each as $e {
        var $ed {
          value = $e.id_dispositivo|first_notempty:""
        }

        conditional {
          if ($ed == $disp) {
            var.update $exist {
              value = $e
            }
          }
        }
      }
    }

    conditional {
      if ($exist == null) {
        db.add vis_cliente_grade_escopo {
          data = {
            id_franqueado : $input.id_franqueado
            id_cliente    : $input.id_cliente
            id_dispositivo: $disp
            grade_ativa   : $ativa
          }
        } as $novo
      }

      else {
        db.patch vis_cliente_grade_escopo {
          field_name = "id"
          field_value = $exist.id
          data = {
            id_franqueado: $input.id_franqueado
            grade_ativa  : $ativa
          }
        } as $patch
      }
    }
  }

  response = {ok: true, grade_ativa: $ativa}
}
