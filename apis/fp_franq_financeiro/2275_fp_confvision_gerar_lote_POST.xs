// Gerar lote de licencas ConfVision (admin) — DEPRECATED: use fp_confvision_fatura_venda
query fp_confvision_gerar_lote verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text id_franqueado? filters=trim
    text plano? filters=trim
    int quantidade?=1 filters=min:1
    timestamp? valido_ate?
    decimal valor?
    text observacao? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }
  
    precondition (($input.plano|is_empty) == false) {
      error = "plano obrigatorio"
    }
  
    var $itens {
      value = [
        {
          plano     : $input.plano
          quantidade: $input.quantidade|first_notempty:1
        }
      ]
    }
  
    function.run "" {
      input = {
        id_franqueado: $input.id_franqueado
        itens        : $itens
        origem       : "admin"
      }
    } as $venda
  }

  response = {
    deprecated: true
    mensagem  : "Use fp_confvision_fatura_venda — licencas criadas como pendente ate pagamento"
    venda     : $venda
  }
}