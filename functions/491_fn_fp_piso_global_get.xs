// Busca valor minimo global (Break-glass) para produto/plano
function fn_fp_piso_global_get {
  input {
    text produto? filters=trim
    text plano? filters=trim
  }

  stack {
    var $valor_minimo {
      value = 0
    }
  
    var $row {
      value = null
    }
  
    conditional {
      if (($input.produto|is_empty) == false && ($input.plano|is_empty) == false) {
        db.query fp_preco_piso_global {
          where = $db.fp_preco_piso_global.produto == $input.produto && $db.fp_preco_piso_global.plano == $input.plano && $db.fp_preco_piso_global.ativo == "S"
          return = {type: "single"}
        } as $row
      
        conditional {
          if ($row != null) {
            var.update $valor_minimo {
              value = $row.valor_minimo|first_notempty:0
            }
          }
        }
      }
    }
  
    conditional {
      if ($valor_minimo < 0) {
        var.update $valor_minimo {
          value = 0
        }
      }
    }
  }

  response = {
    produto     : $input.produto
    plano       : $input.plano
    valor_minimo: $valor_minimo
    piso        : $row
  }
}