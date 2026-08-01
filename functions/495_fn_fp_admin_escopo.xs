// Resolve escopo financeiro da sessao. CEN/REP normal SEMPRE exige id_central.
// Break-glass soh ve global se NAO tiver Central selecionada (id_central vazio).
function fn_fp_admin_escopo {
  input {
    text admin_token? filters=trim
  }

  stack {
    function.run fn_fp_admin_validar {
      input = {admin_token: $input.admin_token}
    } as $admin
  
    var $user_tipo {
      value = $admin|get:"userTipo":""|to_upper|trim
    }
  
    // Aceita SOMENTE true / "S" / "true" — nunca string "N" (evita liberar global por engano)
    var $bg_raw {
      value = $admin|get:"breakglass":false
    }
  
    var $breakglass {
      value = false
    }
  
    conditional {
      if ($bg_raw == true) {
        var.update $breakglass {
          value = true
        }
      }
    
      elseif ($bg_raw == "S" || $bg_raw == "s" || $bg_raw == "true" || $bg_raw == "TRUE") {
        var.update $breakglass {
          value = true
        }
      }
    }
  
    var $id_rep {
      value = ""
    }
  
    var $id_cen {
      value = ""
    }
  
    var $usuario {
      value = $admin|get:"usuario":""|trim
    }
  
    var $id_usuario {
      value = $admin|get:"idUsuario":""|trim
    }
  
    var $id_vinculo {
      value = $admin|get:"idVinculo":""|trim
    }
  
    var $permite_global {
      value = false
    }
  
    conditional {
      if ($user_tipo == "REP") {
        var.update $id_rep {
          value = $admin
            |get:"idVinculo":($admin|get:"idRepresentante":"")
            |trim
        }
      
        var.update $id_cen {
          value = $admin
            |get:"idCentral":($admin|get:"idCentralCatalogo":"")
            |trim
        }
      
        precondition (($id_rep|is_empty) == false) {
          error = "Sessao Representante sem idVinculo — faca logout e login novamente"
        }
      
        precondition (($id_cen|is_empty) == false) {
          error = "Sessao Representante sem idCentral — faca logout e login novamente"
        }
      }
    
      elseif ($user_tipo == "CEN") {
        var.update $id_cen {
          value = $admin
            |get:"idCentral":($admin|get:"idCentralCatalogo":($admin|get:"idVinculo":""))
            |trim
        }
      
        conditional {
          if ($breakglass) {
            // BG sem Central escolhida: pode ver consolidado global
            // "CENTRAL" e ID da matriz — NAO e sinonimo de global
            conditional {
              if (($id_cen|is_empty) == true) {
                var.update $permite_global {
                  value = true
                }
              }
            }
          }
        
          else {
            precondition (($id_cen|is_empty) == false) {
              error = "Sessao Central sem idCentral — faca logout e login novamente"
            }
          }
        }
      }
    
      else {
        precondition (false) {
          error = "Sessao sem perfil financeiro valido (CEN/REP)"
        }
      }
    }
  }

  response = {
    admin          : $admin
    userTipo       : $user_tipo
    breakglass     : $breakglass
    idVinculo      : $id_vinculo
    idRepresentante: $id_rep
    idCentral      : $id_cen
    idUsuario      : $id_usuario
    usuario        : $usuario
    permite_global : $permite_global
  }
}
