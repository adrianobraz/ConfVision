// Franqueado compra pacote de cota extra (soma capacidade)
query fp_cota_comprar_extra verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    int fp_pacote_cota_id?
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_cota_comprar_extra {
      input = {
        id_franqueado    : $input.id_franqueado
        produto          : $input.produto
        fp_pacote_cota_id: $input.fp_pacote_cota_id
        origem           : "franqueado"
      }
    } as $resultado
  }

  response = $resultado
}
