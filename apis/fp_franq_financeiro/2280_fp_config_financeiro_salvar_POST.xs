// Salvar config financeiro
query fp_config_financeiro_salvar verb=POST {
  api_group = "fp_franqFinanceiro"

  input {
    text chave? filters=trim
    text valor? filters=trim
    text descricao? filters=trim
  }

  stack {
    precondition (($input.chave|is_empty) == false) {
      error = "chave obrigatoria"
    }
  
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = $input.chave
    } as $existe
  
    conditional {
      if ($existe != null) {
        db.patch fp_config_financeiro {
          field_name = "chave"
          field_value = $input.chave
          data = {valor: $input.valor, descricao: $input.descricao}
        } as $model
      }
    
      else {
        db.add fp_config_financeiro {
          data = {
            created_at: "now"
            chave     : $input.chave
            valor     : $input.valor
            descricao : $input.descricao
          }
        } as $model
      }
    }
  }

  response = $model
}