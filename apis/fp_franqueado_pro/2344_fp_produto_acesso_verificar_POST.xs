// Verifica acesso efetivo a um produto (assinatura ou bundle FP)
query fp_produto_acesso_verificar verb=POST {
  api_group = "fp_franqueadoPro"

  input {
    text id_franqueado? filters=trim
    text produto? filters=trim
  }

  stack {
    function.run fn_fp_produto_acesso_efetivo {
      input = {
        id_franqueado: $input.id_franqueado
        produto      : $input.produto
      }
    } as $acesso
  }

  response = $acesso
}