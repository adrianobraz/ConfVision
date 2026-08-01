// Seed modulos a la carte — Break-glass pode passar id_central
query fp_modulo_catalogo_seed verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN") {
      error = "Somente a Central pode rodar o seed de modulos a la carte"
    }
  
    var $id_central {
      value = $admin.idVinculo|trim
    }
  
    conditional {
      if ($admin.breakglass == true) {
        conditional {
          if (($input.id_central|is_empty) == false) {
            var.update $id_central {
              value = $input.id_central|trim
            }
          }
        
          elseif (($id_central|is_empty) == true) {
            var.update $id_central {
              value = "CENTRAL"
            }
          }
        }
      }
    
      else {
        precondition (($id_central|is_empty) == false) {
          error = "Sessao sem Central vinculada — faca logout e login novamente"
        }
      }
    }
  
    function.run fn_fp_modulo_catalogo_seed {
      input = {
        user_tipo : "CEN"
        id_vinculo: $id_central
      }
    } as $seed
  
    function.run fn_fp_central_modo_preco {
      input = {id_central: $id_central}
    } as $modo_res
  
    conditional {
      if ($admin.breakglass == true || $modo_res.modo_preco == "piso") {
        function.run fn_fp_catalogo_sync_piso_bg {
          input = {id_central: $id_central}
        } as $sync
      }
    }
  }

  response = $seed|set:"id_central":$id_central|set:"modo_preco":$modo_res.modo_preco
}
