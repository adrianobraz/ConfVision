// Franqueado contrata plano — cria assinatura pendente + fatura
query fp_assinatura_contratar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    int fp_pacote_cota_id?
    text cupom_codigo? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.plano|is_empty) == false) {
      error = "plano obrigatorio"
    }
  
    function.run fn_fp_assinatura_contratar {
      input = {
        id_franqueado    : $input.id_franqueado
        produto          : $input.produto
        plano            : $input.plano
        fp_pacote_cota_id: $input.fp_pacote_cota_id
        cupom_codigo     : $input.cupom_codigo
        origem           : "franqueado"
      }
    } as $resultado
  }

  response = $resultado
}