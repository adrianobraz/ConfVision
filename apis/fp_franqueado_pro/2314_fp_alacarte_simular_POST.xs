// Simula calculo a la carte (preview de preco)
query fp_alacarte_simular verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    json addons?
  }

  stack {
    precondition (($input.plano|is_empty) == false) {
      error = "plano base obrigatorio"
    }
  
    function.run fn_fp_alacarte_calcular {
      input = {
        produto: $input.produto
        plano  : $input.plano
        addons : $input.addons
      }
    } as $calc
  }

  response = $calc
}