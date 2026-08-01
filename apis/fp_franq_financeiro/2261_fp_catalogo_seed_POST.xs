// Seed catalogo — CEN na propria Central; Break-glass pode passar id_central
query fp_catalogo_seed verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    // assinaturas | confvision_licenca | tudo
    text escopo?=assinaturas filters=trim
    text produto? filters=trim
    // Somente Break-glass: seed para outra Central
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
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
  
    var $escopo {
      value = $input.escopo|first_notempty:"assinaturas"|to_lower
    }
  
    var $resultado {
      value = null
    }
  
    conditional {
      if ($admin.userTipo == "CEN") {
        function.run fn_fp_catalogo_seed {
          input = {
            user_tipo : "CEN"
            id_vinculo: $id_central
            escopo    : $escopo
          }
        } as $seed_cen
      
        // Break-glass ou Central em modo piso: trava o piso BG = valores do seed
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
      
        var.update $resultado {
          value = $seed_cen
            |set:"acao":"catalogo_central"
            |set:"id_central":$id_central
            |set:"modo_preco":$modo_res.modo_preco
        }
      }
    
      elseif ($admin.userTipo == "REP") {
        function.run fn_fp_preco_rep_seed {
          input = {
            id_representante: $admin.idVinculo
            id_central      : $id_central
            produto         : $input.produto
            escopo          : $escopo
          }
        } as $seed_rep
      
        var.update $resultado {
          value = $seed_rep|set:"acao":"precos_representante"|set:"user_tipo":"REP"|set:"escopo":$escopo
        }
      }
    
      else {
        precondition (false) {
          error = "Perfil nao autorizado para seed"
        }
      }
    }
  }

  response = $resultado
}
