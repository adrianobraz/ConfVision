// Franqueado contrata plano personalizado a la carte
query fp_assinatura_contratar_alacarte verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    json addons?
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.plano|is_empty) == false) {
      error = "plano base obrigatorio"
    }
  
    function.run fn_fp_assinatura_contratar_alacarte {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
        plano        : $input.plano
        addons       : $input.addons
        origem       : "franqueado"
      }
    } as $resultado
  }

  response = $resultado
}