// Resolve franqueado + logo pelo Host (login ConfVision/FP sem auth)
query fp_whitelabel_by_fqdn verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text fqdn? filters=trim|lower
  }

  stack {
    precondition (($input.fqdn|is_empty) == false) {
      error = "fqdn obrigatorio"
    }
  
    // Prefere ConfVision (fqdn_cv); fallback FranqueadoPro (fqdn)
    db.query fp_whitelabel {
      where = $db.fp_whitelabel.fqdn_cv == $input.fqdn
      return = {type: "single"}
    } as $row_cv
  
    var $row {
      value = $row_cv
    }
  
    var $app {
      value = "confvision"
    }
  
    conditional {
      if ($row == null) {
        db.query fp_whitelabel {
          where = $db.fp_whitelabel.fqdn == $input.fqdn
          return = {type: "single"}
        } as $row_fp
      
        var.update $row {
          value = $row_fp
        }
      
        var.update $app {
          value = "franqueadopro"
        }
      }
    }
  }

  response = {dados: $row, app: $app, fqdn: $input.fqdn}
}