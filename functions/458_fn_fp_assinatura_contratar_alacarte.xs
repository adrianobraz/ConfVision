// Cria assinatura pendente a la carte (plano base + modulos avulsos)
function fn_fp_assinatura_contratar_alacarte {
  input {
    text id_franqueado? filters=trim
    text produto?=franqueadopro filters=trim
    text plano? filters=trim
    json addons?
    text origem?=franqueado filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    function.run fn_fp_assinatura_salvar_alacarte {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
        plano        : $input.plano
        addons       : $input.addons
        origem       : $input.origem
        admin_usuario: $input.admin_usuario
      }
    } as $resultado
  }

  response = $resultado
}