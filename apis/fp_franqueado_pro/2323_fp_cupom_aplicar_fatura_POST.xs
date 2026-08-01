// Aplica cupom em fatura aberta ou paga (credito) — FranqueadoPro
query fp_cupom_aplicar_fatura verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text codigo? filters=trim
    int fatura_id? filters=min:1
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_cupom_aplicar_fatura {
      input = {
        codigo       : $input.codigo
        id_franqueado: $input.id_franqueado
        fatura_id    : $input.fatura_id
        origem       : "franqueado"
      }
    } as $resultado
  }

  response = $resultado
}