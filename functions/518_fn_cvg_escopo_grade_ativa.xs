// Retorna grade_ativa por escopo cliente+dispositivo (fallback vis_cliente_grade_config)
function fn_cvg_escopo_grade_ativa {
  input {
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
  }

  stack {
    var $disp {
      value = $input.id_dispositivo|first_notempty:""
    }

    db.query vis_cliente_grade_escopo {
      where = $db.vis_cliente_grade_escopo.id_cliente == $input.id_cliente
      return = {type: "list"}
    } as $escopos_raw

    var $escopo {
      value = null
    }

    foreach ($escopos_raw) {
      each as $e {
        var $ed {
          value = $e.id_dispositivo|first_notempty:""
        }

        conditional {
          if ($ed == $disp) {
            var.update $escopo {
              value = $e
            }
          }
        }
      }
    }

    var $ativa {
      value = false
    }

    conditional {
      if ($escopo != null) {
        var.update $ativa {
          value = $escopo.grade_ativa|first_notempty:false
        }
      }

      else {
        db.query vis_cliente_grade_config {
          where = $db.vis_cliente_grade_config.id_cliente == $input.id_cliente
          return = {type: "single"}
        } as $cfg

        conditional {
          if ($cfg != null) {
            var.update $ativa {
              value = $cfg.grade_ativa|first_notempty:false
            }
          }
        }
      }
    }
  }

  response = $ativa
}
