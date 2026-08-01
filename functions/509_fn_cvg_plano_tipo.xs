// Classifica plano da camera: armado | 24h | outro
function fn_cvg_plano_tipo {
  input {
    text plano? filters=trim
  }

  stack {
    function.run fn_vis_plano_flags {
      input = {plano: $input.plano}
    } as $flags

    var $tipo {
      value = "outro"
    }

    conditional {
      if ($flags.captura_analitico == false) {
        var.update $tipo {
          value = "outro"
        }
      }

      elseif ($flags.somente_armado) {
        var.update $tipo {
          value = "armado"
        }
      }

      else {
        var.update $tipo {
          value = "24h"
        }
      }
    }
  }

  response = {
    tipo              : $tipo
    captura_analitico : $flags.captura_analitico
    somente_armado    : $flags.somente_armado
  }
}
