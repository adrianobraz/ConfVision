// Seed Cota 50/200/800
query fp_pacote_cota_seed verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN" || $admin.userTipo == "REP") {
      error = "Sem permissao"
    }
  
    var $id_central {
      value = $admin.idCentralCatalogo
        |first_notempty:($admin.idVinculo|trim)
    }
  
    conditional {
      if ($admin.userTipo == "REP") {
        var.update $id_central {
          value = $admin.idCentralCatalogo|first_notempty:"CENTRAL"|trim
        }
      }
    
      elseif ($admin.breakglass && (($input.id_central|is_empty) == false)) {
        var.update $id_central {
          value = $input.id_central|trim
        }
      }
    }
  
    conditional {
      if ($id_central|is_empty) {
        var.update $id_central {
          value = "CENTRAL"
        }
      }
    }
  
    function.run fn_fp_pacote_cota_seed {
      input = {id_central: $id_central}
    } as $resultado
  }

  response = $resultado
}