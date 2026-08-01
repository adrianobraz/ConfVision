// Lista usuarios por vinculo — CEN/BG: qualquer nivel; REP: so o proprio ou franqueado da carteira
query fp_usuario_listar_vinculo verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text id_vinculo? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $user_tipo {
      value = $escopo|get:"userTipo":""|to_upper|trim
    }
  
    var $bg {
      value = ($escopo|get:"breakglass":false) == true
    }
  
    precondition ($user_tipo == "CEN" || $user_tipo == "REP" || $bg) {
      error = "Somente Central ou Representante pode listar usuarios da hierarquia"
    }

    precondition (($input.id_vinculo|is_empty) == false) {
      error = "id_vinculo obrigatorio"
    }

    var $id_vin {
      value = $input.id_vinculo|trim
    }

    // REP: so o proprio ID ou franqueado da carteira
    conditional {
      if ($user_tipo == "REP" && $bg == false) {
        var $id_rep {
          value = $escopo
            |get:"idVinculo":($escopo|get:"idRepresentante":"")
            |trim
        }

        precondition (($id_rep|is_empty) == false) {
          error = "Sessao Representante sem idVinculo"
        }

        conditional {
          if ($id_vin != $id_rep) {
            function.run fn_fp_admin_assert_escopo_franqueado {
              input = {
                admin_token  : $input.admin_token
                id_franqueado: $id_vin
              }
            } as $escopo_fra
          }
        }
      }
    }

    var $id_cen {
      value = $input.id_central|trim
    }

    conditional {
      if ($id_cen|is_empty) {
        var.update $id_cen {
          value = $escopo|get:"idCentral":""|trim
        }
      }
    }

    function.run fn_fp_usuario_listar_vinculo {
      input = {
        id_vinculo: $id_vin
        id_central: $id_cen
      }
    } as $lista
  }

  response = $lista
}
