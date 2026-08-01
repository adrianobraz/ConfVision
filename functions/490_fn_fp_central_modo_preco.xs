// Retorna modo_preco da Central (livre|piso). Default: livre
function fn_fp_central_modo_preco {
  input {
    text id_central? filters=trim
  }

  stack {
    var $id_central {
      value = $input.id_central|first_notempty:"CENTRAL"|trim
    }
  
    var $modo {
      value = "livre"
    }
  
    var $cfg {
      value = null
    }
  
    db.query fp_central_preco_config {
      where = $db.fp_central_preco_config.id_central == $id_central
      return = {type: "single"}
    } as $cfg
  
    conditional {
      if ($cfg != null) {
        var.update $modo {
          value = $cfg.modo_preco
            |first_notempty:"livre"
            |to_lower
            |trim
        }
      }
    }
  
    conditional {
      if ($modo != "piso") {
        var.update $modo {
          value = "livre"
        }
      }
    }
  }

  response = {
    id_central: $id_central
    modo_preco: $modo
    config    : $cfg
  }
}