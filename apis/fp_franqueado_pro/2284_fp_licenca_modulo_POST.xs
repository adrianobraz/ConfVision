// Verifica se modulo/chave_menu esta liberado pelo plano
query fp_licenca_modulo verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text chave_menu? filters=trim
    text produto?=franqueadopro filters=trim
  }

  stack {
    precondition (($input.chave_menu|is_empty) == false) {
      error = "chave_menu obrigatorio"
    }
  
    function.run fn_fp_licenca_modulo_check {
      input = {
        id_franqueado: $input.id_franqueado
        chave_menu   : $input.chave_menu
        produto      : $input.produto
      }
    } as $result
  }

  response = $result
}