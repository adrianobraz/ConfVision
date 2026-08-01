// Lista modulos a la carte — CEN/REP na propria Central; Break-glass pode filtrar id_central
query fp_modulo_catalogo_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text produto? filters=trim
    text ativo? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    precondition ($admin.userTipo == "CEN" || $admin.userTipo == "REP") {
      error = "Sem permissao para listar catalogo de modulos a la carte"
    }
  
    // Catalogo = Central da sessao (CEN: idVinculo; REP: idCentral)
    var $id_central {
      value = $admin.idCentral|first_notempty:($admin.idCentralCatalogo|first_notempty:"")
    }
  
    conditional {
      if ($admin.userTipo == "CEN" && ($id_central|is_empty) == true) {
        var.update $id_central {
          value = $admin.idVinculo|trim
        }
      }
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
  
    function.run fn_fp_central_modo_preco {
      input = {id_central: $id_central}
    } as $modo_res
  
    db.query fp_modulo_catalogo {
      where = $db.fp_modulo_catalogo.id_central == $id_central && $db.fp_modulo_catalogo.id_representante == "" && $db.fp_modulo_catalogo.produto ==? $input.produto && $db.fp_modulo_catalogo.ativo ==? $input.ativo
      sort = {
        fp_modulo_catalogo.ordem: "asc"
        fp_modulo_catalogo.grupo: "asc"
      }
    
      return = {type: "list"}
    } as $lista_raw
  
    var $lista {
      value = []
    }
  
    foreach ($lista_raw) {
      each as $item {
        var.update $lista {
          value = $lista|push:($item
            |set:"modo_preco":$modo_res.modo_preco
            |set:"valor_piso_minimo":($item.valor_piso_breakglass|first_notempty:0)
          )
        }
      }
    }
  }

  response = {
    dados     : $lista
    total     : $lista|count
    id_central: $id_central
    modo_preco: $modo_res.modo_preco
  }
}
