// Lista representantes da Central logada (CEN / Break-glass com Central)
query fp_representante_listar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text admin_token? filters=trim
    text busca? filters=trim
    text id_central? filters=trim
  }

  stack {
    function.run fn_fp_admin_escopo {
      input = {admin_token: $input.admin_token}
    } as $escopo
  
    var $user_tipo {
      value = $escopo|get:"userTipo":""|to_upper|trim
    }
  
    precondition ($user_tipo == "CEN" || ($escopo|get:"breakglass":false) == true) {
      error = "Somente Central pode listar representantes"
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
  
    precondition (($id_cen|is_empty) == false) {
      error = "Central nao identificada na sessao — selecione a Central (break-glass) ou faca login novamente"
    }
  
    function.run fn_fp_representante_listar {
      input = {
        id_central: $id_cen
        busca     : $input.busca
      }
    } as $lista
  }

  response = $lista
}
