// Compra de licencas ConfVision pelo franqueado (self-service)
query fp_confvision_comprar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text admin_usuario? filters=trim
    object[] itens? {
      schema {
        text plano? filters=trim
        int quantidade?=1 filters=min:1
      }
    }
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    function.run fn_fp_confvision_fatura_venda {
      input = {
        id_franqueado: $input.id_franqueado
        itens        : $input.itens
        origem       : "franqueado"
        admin_usuario: $input.admin_usuario
      }
    } as $resultado
  }

  response = $resultado
}