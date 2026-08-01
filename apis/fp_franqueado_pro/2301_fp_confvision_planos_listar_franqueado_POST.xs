// Lista planos ConfVision para self-service do franqueado
query fp_confvision_planos_listar_franqueado verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_confvision_planos_listar {
      input = {id_franqueado: $input.id_franqueado}
    } as $resultado
  }

  response = $resultado
}